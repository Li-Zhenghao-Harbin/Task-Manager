package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Task struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	ParentID    string `json:"parentId"`
	Issuer      string `json:"issuer"`
	Unit        string `json:"unit"`
	PublishDate string `json:"publishDate"`
	DueDate     string `json:"dueDate"`
	TaskType    string `json:"taskType"`
	Status      string `json:"status"`
	Deliverable string `json:"deliverable"`
	// AgentID only applies to code tasks; the service clears it for other types.
	AgentID     string `json:"agentId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	CompletedAt string `json:"completedAt"`
	Output      Output `json:"output"`
}

// Output records the deliverable result and its structured quantity metrics.
type Output struct {
	Result            string `json:"result"`
	NewFeatures       int    `json:"newFeatures"`
	ModifiedFeatures  int    `json:"modifiedFeatures"`
	Applications      int    `json:"applications"`
	Skills            int    `json:"skills"`
	NewDocuments      int    `json:"newDocuments"`
	ModifiedDocuments int    `json:"modifiedDocuments"`
	Supports          int    `json:"supports"`
}

type TaskInput struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	ParentID    string `json:"parentId"`
	Issuer      string `json:"issuer"`
	Unit        string `json:"unit"`
	PublishDate string `json:"publishDate"`
	DueDate     string `json:"dueDate"`
	TaskType    string `json:"taskType"`
	AgentID     string `json:"agentId"`
}

type Unit struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AgentApp registers one Agent application together with the model it runs.
// Projects are existing tasks whose type is "code"; there is no separate project entity.
//
// ProjectIDs is a derived view, not an independent relation: it always mirrors the code
// tasks whose AgentID points at this application (see syncAgentProjectsLocked). Task is the
// single source of truth so both edit directions cannot drift apart.
type AgentApp struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Model      string   `json:"model"`
	ProjectIDs []string `json:"projectIds"`
	Available  bool     `json:"available"`
	// ResetAt uses the datetime-local format (minute precision), e.g. "2026-09-21T17:30".
	ResetAt string `json:"resetAt"`
	// Prompts maps a task ID to the prompts written for it on this application: the same
	// application needs different instructions per task. The key is not an independent
	// relation — it rides on Task.AgentID — but it survives unlinking on purpose, so
	// re-attaching a task brings its prompts back. Only deleting the task clears them.
	Prompts   map[string][]AgentPrompt `json:"prompts"`
	CreatedAt string                   `json:"createdAt"`
	UpdatedAt string                   `json:"updatedAt"`
}

// AgentPrompt is one prompt stored against a (Agent application, task) pair. Text is the
// whole payload: a prompt has no title or note of its own.
type AgentPrompt struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type AgentAppInput struct {
	Name       string   `json:"name"`
	Model      string   `json:"model"`
	ProjectIDs []string `json:"projectIds"`
	Available  bool     `json:"available"`
	ResetAt    string   `json:"resetAt"`
}

// agentResetLayout pins the minute precision used by the frontend datetime-local control.
const agentResetLayout = "2006-01-02T15:04"

// maxAgentProjects caps a single Agent application's project selection.
const maxAgentProjects = 200

// maxPromptsPerPair caps how many prompts one (application, task) pair can hold.
const maxPromptsPerPair = 50

// maxPromptLength caps one prompt's text, counted in runes so CJK is not penalised.
const maxPromptLength = 20000

// agentResetCheckInterval is how often the running app looks for reset times that passed.
// Reset times have minute precision, so a sub-minute poll is enough to feel immediate.
const agentResetCheckInterval = 30 * time.Second

// agentResetEvent is the frontend event emitted after Agent applications are restored.
const agentResetEvent = "agents:reset"

type DataStore struct {
	Units         []Unit     `json:"units"`
	DefaultUnitID string     `json:"defaultUnitId"`
	Tasks         []Task     `json:"tasks"`
	Agents        []AgentApp `json:"agents"`
	Settings      Settings   `json:"settings"`
}

type Settings struct {
	BackgroundColor string `json:"backgroundColor"`
	AutoStart       bool   `json:"autoStart"`
}

type App struct {
	ctx           context.Context
	tasks         []Task
	units         []Unit
	agents        []AgentApp
	defaultUnitID string
	settings      Settings
	dataPath      string
	// notifier delivers a desktop notification. It is a field so tests can replace it;
	// nil falls back to the real Windows toast implementation.
	notifier func(title, body string) error
	mu       sync.RWMutex
}

func NewApp() *App { return &App{} }

// notify sends a desktop notification through the configured notifier.
func (a *App) notify(title, body string) {
	send := a.notifier
	if send == nil {
		send = showWindowsToast
	}
	// A failed notification must never affect state restoration, so the error is dropped.
	_ = send(title, body)
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Always store beside the executable so the data file stays portable.
	exePath, err := os.Executable()
	dir := ""
	if err == nil {
		dir = filepath.Dir(exePath)
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	a.dataPath = filepath.Join(dir, "task-manager-data.json")
	_ = a.load()
	// Restore Agent applications whose reset time passed while the app was closed, then
	// keep watching so a reset that happens while the app runs also recovers and notifies.
	a.checkAgentResets()
	go a.watchAgentResets()
}

// checkAgentResets restores every Agent whose reset time has passed, persists the change and
// notifies. The notification is sent outside the lock so no child process starts under it.
func (a *App) checkAgentResets() {
	a.mu.Lock()
	restored := a.applyAgentResetLocked()
	if len(restored) > 0 {
		_ = a.saveLocked()
	}
	ctx := a.ctx
	a.mu.Unlock()
	if len(restored) == 0 {
		return
	}
	for _, label := range restored {
		a.notify("Agent 已恢复可用", label+" 的额度重置时间已到，现已恢复可用。")
	}
	if ctx != nil {
		// Let the UI refresh its list without the user having to touch anything.
		runtime.EventsEmit(ctx, agentResetEvent, restored)
	}
}

// watchAgentResets polls for reset times that pass while the app is running and stops with ctx.
func (a *App) watchAgentResets() {
	ticker := time.NewTicker(agentResetCheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			a.checkAgentResets()
		case <-a.ctx.Done():
			return
		}
	}
}

func (a *App) ListTasks() []Task {
	a.mu.RLock()
	defer a.mu.RUnlock()
	// Return an empty JSON array ([]) rather than null when there are no tasks.
	// The Vue UI can then safely render task counts on a fresh installation.
	items := make([]Task, len(a.tasks))
	copy(items, a.tasks)
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	return items
}

func (a *App) ListUnits() []Unit {
	a.mu.RLock()
	defer a.mu.RUnlock()
	items := make([]Unit, len(a.units))
	copy(items, a.units)
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func (a *App) GetDefaultUnitID() string { a.mu.RLock(); defer a.mu.RUnlock(); return a.defaultUnitID }

func (a *App) GetSettings() Settings { a.mu.RLock(); defer a.mu.RUnlock(); return a.settings }

func (a *App) UpdateSettings(settings Settings) (Settings, error) {
	if settings.BackgroundColor == "" {
		settings.BackgroundColor = "#f7faff"
	}
	if len(settings.BackgroundColor) != 7 || settings.BackgroundColor[0] != '#' {
		return Settings{}, fmt.Errorf("背景色格式无效")
	}
	if err := setAutoStart(settings.AutoStart); err != nil {
		return Settings{}, err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.settings = settings
	return settings, a.saveLocked()
}

func setAutoStart(enabled bool) error {
	const key = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if enabled {
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		return exec.Command("reg", "add", key, "/v", "Task Manager", "/t", "REG_SZ", "/d", `"`+exePath+`"`, "/f").Run()
	}
	cmd := exec.Command("reg", "delete", key, "/v", "Task Manager", "/f")
	if err := cmd.Run(); err != nil {
		return nil
	}
	return nil
}

func (a *App) CreateUnit(name string) (Unit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Unit{}, fmt.Errorf("请填写单位名称")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, unit := range a.units {
		if unit.Name == name {
			return Unit{}, fmt.Errorf("单位已存在")
		}
	}
	unit := Unit{ID: uuid.NewString(), Name: name}
	a.units = append(a.units, unit)
	return unit, a.saveLocked()
}

func (a *App) UpdateUnit(id, name string) (Unit, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Unit{}, fmt.Errorf("请填写单位名称")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, unit := range a.units {
		if unit.ID != id && unit.Name == name {
			return Unit{}, fmt.Errorf("单位已存在")
		}
	}
	for i := range a.units {
		if a.units[i].ID == id {
			old := a.units[i].Name
			a.units[i].Name = name
			for j := range a.tasks {
				if a.tasks[j].Unit == old {
					a.tasks[j].Unit = name
				}
			}
			return a.units[i], a.saveLocked()
		}
	}
	return Unit{}, fmt.Errorf("单位不存在")
}

func (a *App) SetDefaultUnit(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, unit := range a.units {
		if unit.ID == id {
			a.defaultUnitID = id
			return a.saveLocked()
		}
	}
	return fmt.Errorf("单位不存在")
}

func (a *App) DeleteUnit(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := -1
	for i, unit := range a.units {
		if unit.ID == id {
			idx = i
			for _, task := range a.tasks {
				if task.Unit == unit.Name {
					return fmt.Errorf("该单位仍有关联任务，请先处理相关任务")
				}
			}
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("单位不存在")
	}
	a.units = append(a.units[:idx], a.units[idx+1:]...)
	if a.defaultUnitID == id {
		a.defaultUnitID = ""
		if len(a.units) == 1 {
			a.defaultUnitID = a.units[0].ID
		}
	}
	return a.saveLocked()
}

// ---------- Agent applications ----------

// ListAgentApps returns the registered applications, first restoring any whose reset time has
// passed. It routes through checkAgentResets so a restore triggered by opening the list still
// notifies, instead of silently consuming the reset before the watcher sees it.
func (a *App) ListAgentApps() []AgentApp {
	a.checkAgentResets()
	a.mu.RLock()
	defer a.mu.RUnlock()
	items := make([]AgentApp, len(a.agents))
	copy(items, a.agents)
	for i := range items {
		if items[i].ProjectIDs == nil {
			// Keep projectIds as [] rather than null so the Vue list can iterate safely.
			items[i].ProjectIDs = []string{}
		}
		if items[i].Prompts == nil {
			items[i].Prompts = map[string][]AgentPrompt{}
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items
}

func (a *App) CreateAgentApp(input AgentAppInput) (AgentApp, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.validateAgentLocked(input); err != nil {
		return AgentApp{}, err
	}
	now := time.Now().Format(time.RFC3339)
	agent := AgentApp{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(input.Name),
		Model:     strings.TrimSpace(input.Model),
		Available: input.Available,
		ResetAt:   strings.TrimSpace(input.ResetAt),
		Prompts:   map[string][]AgentPrompt{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if agent.Available {
		agent.ResetAt = ""
	}
	a.agents = append(a.agents, agent)
	// Seed the relation from the task side, then let the tasks own the resulting list.
	a.assignAgentProjectsLocked(agent.ID, input.ProjectIDs)
	a.syncAgentProjectsLocked()
	return a.agents[len(a.agents)-1], a.saveLocked()
}

func (a *App) UpdateAgentApp(id string, input AgentAppInput) (AgentApp, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.agentIndexOf(id)
	if idx < 0 {
		return AgentApp{}, fmt.Errorf("Agent 应用不存在")
	}
	if err := a.validateAgentLocked(input); err != nil {
		return AgentApp{}, err
	}
	updated := a.agents[idx]
	updated.Name = strings.TrimSpace(input.Name)
	updated.Model = strings.TrimSpace(input.Model)
	updated.Available = input.Available
	updated.ResetAt = strings.TrimSpace(input.ResetAt)
	if updated.Available {
		updated.ResetAt = ""
	}
	updated.UpdatedAt = time.Now().Format(time.RFC3339)
	a.agents[idx] = updated
	// The submitted project list is a desired set: it reassigns or detaches tasks, and the
	// tasks then feed the rebuilt list back into every Agent.
	a.assignAgentProjectsLocked(id, input.ProjectIDs)
	a.syncAgentProjectsLocked()
	return a.agents[idx], a.saveLocked()
}

func (a *App) DeleteAgentApp(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.agentIndexOf(id)
	if idx < 0 {
		return fmt.Errorf("Agent 应用不存在")
	}
	a.agents = append(a.agents[:idx], a.agents[idx+1:]...)
	// Drop references held by tasks so no task points at a deleted Agent, then rebuild the
	// surviving Agents' lists from those tasks.
	now := time.Now().Format(time.RFC3339)
	for i := range a.tasks {
		if a.tasks[i].AgentID == id {
			a.tasks[i].AgentID = ""
			a.tasks[i].UpdatedAt = now
		}
	}
	a.syncAgentProjectsLocked()
	return a.saveLocked()
}

func (a *App) agentIndexOf(id string) int {
	for i := range a.agents {
		if a.agents[i].ID == id {
			return i
		}
	}
	return -1
}

func (a *App) agentExistsLocked(id string) bool { return a.agentIndexOf(id) >= 0 }

// normalizeProjectIDs removes blanks and duplicates while preserving the user's order.
func (a *App) normalizeProjectIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
		if len(result) >= maxAgentProjects {
			break
		}
	}
	return result
}

func (a *App) validateAgentLocked(input AgentAppInput) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Model) == "" {
		return fmt.Errorf("请填写应用名和模型名")
	}
	if reset := strings.TrimSpace(input.ResetAt); reset != "" {
		if _, err := time.ParseInLocation(agentResetLayout, reset, time.Local); err != nil {
			return fmt.Errorf("下次重置时间格式无效")
		}
	}
	for _, id := range a.normalizeProjectIDs(input.ProjectIDs) {
		found := false
		for i := range a.tasks {
			// Associated projects must be tasks of type "code".
			if a.tasks[i].ID == id && a.tasks[i].TaskType == "code" {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("关联项目必须是代码类型的任务")
		}
	}
	return nil
}

// applyAgentResetLocked restores availability once the reset time has passed and returns the
// labels of the restored applications so the caller can notify about them.
func (a *App) applyAgentResetLocked() []string {
	now := time.Now()
	restored := make([]string, 0, 1)
	for i := range a.agents {
		if a.agents[i].Available || strings.TrimSpace(a.agents[i].ResetAt) == "" {
			continue
		}
		reset, err := time.ParseInLocation(agentResetLayout, a.agents[i].ResetAt, time.Local)
		if err != nil || now.Before(reset) {
			continue
		}
		a.agents[i].Available = true
		a.agents[i].ResetAt = ""
		a.agents[i].UpdatedAt = now.Format(time.RFC3339)
		restored = append(restored, agentLabel(a.agents[i]))
	}
	return restored
}

// agentLabel renders an Agent application the way the UI shows it: "应用名-模型名".
func agentLabel(agent AgentApp) string { return agent.Name + "-" + agent.Model }

// ---------- Agent prompts ----------
//
// A prompt belongs to the (application, task) pair: the same application needs different
// instructions per task. Prompts therefore live on the Agent, keyed by task ID, and only
// ever shrink for one reason — the task itself being deleted.

// normalizePrompts drops blank entries, fills in missing IDs and caps the list.
// Truncating beats rejecting: a pasted oversized prompt should not lose the whole save.
func normalizePrompts(prompts []AgentPrompt) []AgentPrompt {
	result := make([]AgentPrompt, 0, len(prompts))
	seen := make(map[string]bool, len(prompts))
	for _, prompt := range prompts {
		if strings.TrimSpace(prompt.Text) == "" {
			// An entry the user added but never typed into; it must not reach the data file.
			continue
		}
		item := AgentPrompt{ID: strings.TrimSpace(prompt.ID), Text: prompt.Text}
		if item.ID == "" || seen[item.ID] {
			item.ID = uuid.NewString()
		}
		seen[item.ID] = true
		if runes := []rune(item.Text); len(runes) > maxPromptLength {
			item.Text = string(runes[:maxPromptLength])
		}
		result = append(result, item)
		if len(result) >= maxPromptsPerPair {
			break
		}
	}
	return result
}

func samePrompts(left, right []AgentPrompt) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// normalizeAgentPromptsLocked makes every Agent's prompt map safe to work with: a nil map
// becomes empty, keys left behind by deleted tasks are removed and each list is folded
// through normalizePrompts. Idempotent, so it is safe to run on every load.
func (a *App) normalizeAgentPromptsLocked() bool {
	changed := false
	known := make(map[string]bool, len(a.tasks))
	for i := range a.tasks {
		known[a.tasks[i].ID] = true
	}
	for i := range a.agents {
		if a.agents[i].Prompts == nil {
			a.agents[i].Prompts = map[string][]AgentPrompt{}
			changed = true
			continue
		}
		for taskID, list := range a.agents[i].Prompts {
			if taskID == "" || !known[taskID] {
				// The task is gone, so its prompts have nothing left to describe.
				delete(a.agents[i].Prompts, taskID)
				changed = true
				continue
			}
			normalized := normalizePrompts(list)
			if len(normalized) == 0 {
				delete(a.agents[i].Prompts, taskID)
				changed = true
				continue
			}
			if !samePrompts(list, normalized) {
				a.agents[i].Prompts[taskID] = normalized
				changed = true
			}
		}
	}
	return changed
}

// validatePromptTargetLocked resolves a (application, task) pair and rejects anything that
// is not an existing code task currently attached to that application.
func (a *App) validatePromptTargetLocked(agentID, taskID string) (int, error) {
	idx := a.agentIndexOf(agentID)
	if idx < 0 {
		return -1, fmt.Errorf("Agent 应用不存在")
	}
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return -1, fmt.Errorf("请先选择一个关联任务")
	}
	for i := range a.tasks {
		if a.tasks[i].ID != taskID {
			continue
		}
		if a.tasks[i].TaskType != "code" {
			return -1, fmt.Errorf("只有代码类型的任务可以管理 Prompt")
		}
		if a.tasks[i].AgentID != agentID {
			return -1, fmt.Errorf("该任务未关联到此 Agent 应用")
		}
		return idx, nil
	}
	return -1, fmt.Errorf("任务不存在")
}

// ListAgentPrompts returns the prompts stored for one (application, task) pair.
func (a *App) ListAgentPrompts(agentID, taskID string) ([]AgentPrompt, error) {
	a.checkAgentResets()
	a.mu.RLock()
	defer a.mu.RUnlock()
	idx, err := a.validatePromptTargetLocked(agentID, taskID)
	if err != nil {
		return nil, err
	}
	list := a.agents[idx].Prompts[strings.TrimSpace(taskID)]
	items := make([]AgentPrompt, len(list))
	copy(items, list)
	return items, nil
}

// UpdateAgentPrompts replaces the whole prompt list of one (application, task) pair.
// The panel edits a local list and saves it as a whole, so add / edit / delete all arrive
// here as one replacement — which also makes "cleared the text box" mean "delete it".
func (a *App) UpdateAgentPrompts(agentID, taskID string, prompts []AgentPrompt) ([]AgentPrompt, error) {
	a.checkAgentResets()
	a.mu.Lock()
	defer a.mu.Unlock()
	idx, err := a.validatePromptTargetLocked(agentID, taskID)
	if err != nil {
		return nil, err
	}
	taskID = strings.TrimSpace(taskID)
	normalized := normalizePrompts(prompts)
	if a.agents[idx].Prompts == nil {
		a.agents[idx].Prompts = map[string][]AgentPrompt{}
	}
	if len(normalized) == 0 {
		delete(a.agents[idx].Prompts, taskID)
	} else {
		a.agents[idx].Prompts[taskID] = normalized
	}
	a.agents[idx].UpdatedAt = time.Now().Format(time.RFC3339)
	if err := a.saveLocked(); err != nil {
		return nil, err
	}
	items := make([]AgentPrompt, len(normalized))
	copy(items, normalized)
	return items, nil
}

// ---------- Agent / task relation ----------

// syncAgentProjectsLocked rebuilds every Agent's project list from the tasks, which are the
// single source of truth for the relation. Returns whether anything changed.
func (a *App) syncAgentProjectsLocked() bool {
	changed := false
	now := time.Now().Format(time.RFC3339)
	for i := range a.agents {
		ids := make([]string, 0, 4)
		for j := range a.tasks {
			if a.tasks[j].TaskType == "code" && a.tasks[j].AgentID == a.agents[i].ID {
				ids = append(ids, a.tasks[j].ID)
			}
		}
		if sameIDs(a.agents[i].ProjectIDs, ids) {
			continue
		}
		a.agents[i].ProjectIDs = ids
		a.agents[i].UpdatedAt = now
		changed = true
	}
	return changed
}

// assignAgentProjectsLocked treats ids as the desired project set of one Agent: listed code
// tasks are moved onto this Agent, code tasks that left the set are detached from it.
func (a *App) assignAgentProjectsLocked(agentID string, ids []string) {
	wanted := map[string]bool{}
	for _, id := range a.normalizeProjectIDs(ids) {
		wanted[id] = true
	}
	for i := range a.tasks {
		if a.tasks[i].TaskType != "code" {
			continue
		}
		switch {
		case wanted[a.tasks[i].ID]:
			a.tasks[i].AgentID = agentID
		case a.tasks[i].AgentID == agentID:
			// The task carrying the pointer decides; a project list that no longer holds it
			// means the user unchecked it here.
			a.tasks[i].AgentID = ""
		}
	}
}

// repairAgentLinksLocked fixes data written before the relation was unified: an Agent could
// list a project that never made it onto the task. Tasks that already point somewhere keep
// their pointer, so the repair never overrides an explicit task-side choice. Idempotent.
func (a *App) repairAgentLinksLocked() bool {
	changed := false
	for i := range a.agents {
		for _, projectID := range a.agents[i].ProjectIDs {
			for j := range a.tasks {
				if a.tasks[j].ID != projectID || a.tasks[j].TaskType != "code" {
					continue
				}
				if a.tasks[j].AgentID == "" {
					a.tasks[j].AgentID = a.agents[i].ID
					changed = true
				}
				break
			}
		}
	}
	if a.syncAgentProjectsLocked() {
		changed = true
	}
	return changed
}

func sameIDs(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func (a *App) CreateTask(input TaskInput) (Task, error) {
	if err := a.validate(input, ""); err != nil {
		return Task{}, err
	}
	now := time.Now().Format(time.RFC3339)
	task := Task{ID: uuid.NewString(), Name: strings.TrimSpace(input.Name), Kind: input.Kind, ParentID: input.ParentID, Issuer: strings.TrimSpace(input.Issuer), Unit: strings.TrimSpace(input.Unit), PublishDate: input.PublishDate, DueDate: input.DueDate, TaskType: input.TaskType, AgentID: input.AgentID, Status: "in_progress", CreatedAt: now, UpdatedAt: now}
	if task.TaskType != "code" {
		// Only code tasks may carry an Agent association.
		task.AgentID = ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tasks = append(a.tasks, task)
	// Keep every Agent application's project list in step with the task side.
	a.syncAgentProjectsLocked()
	return task, a.saveLocked()
}

func (a *App) DeleteTask(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.indexOf(id)
	if idx < 0 {
		return fmt.Errorf("任务不存在")
	}
	for _, task := range a.tasks {
		if task.ParentID == id {
			return fmt.Errorf("该任务仍有子任务，请先删除子任务")
		}
	}
	a.tasks = append(a.tasks[:idx], a.tasks[idx+1:]...)
	// Prompts describe an (application, task) pair, so the task's prompts leave with it.
	for i := range a.agents {
		delete(a.agents[i].Prompts, id)
	}
	// Detach the deleted task from every Agent application's project list by rebuilding them.
	a.syncAgentProjectsLocked()
	return a.saveLocked()
}

func (a *App) UpdateTask(task Task) (Task, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.indexOf(task.ID)
	if idx < 0 {
		return Task{}, fmt.Errorf("任务不存在")
	}
	input := TaskInput{Name: task.Name, Kind: task.Kind, ParentID: task.ParentID, Issuer: task.Issuer, Unit: task.Unit, PublishDate: task.PublishDate, DueDate: task.DueDate, TaskType: task.TaskType, AgentID: task.AgentID}
	if err := a.validateLocked(input, task.ID); err != nil {
		return Task{}, err
	}
	old := a.tasks[idx]
	task.Name, task.Issuer = strings.TrimSpace(task.Name), strings.TrimSpace(task.Issuer)
	task.CreatedAt, task.UpdatedAt = old.CreatedAt, time.Now().Format(time.RFC3339)
	if task.TaskType != "code" {
		// Only code tasks may carry an Agent association.
		task.AgentID = ""
	}
	if !validStatus(task.Status) {
		task.Status = old.Status
	}
	if task.Output.NewFeatures < 0 {
		task.Output.NewFeatures = 0
	}
	if task.Output.ModifiedFeatures < 0 {
		task.Output.ModifiedFeatures = 0
	}
	if task.Output.Applications < 0 {
		task.Output.Applications = 0
	}
	if task.Output.Skills < 0 {
		task.Output.Skills = 0
	}
	if task.Output.NewDocuments < 0 {
		task.Output.NewDocuments = 0
	}
	if task.Output.ModifiedDocuments < 0 {
		task.Output.ModifiedDocuments = 0
	}
	if task.Output.Supports < 0 {
		task.Output.Supports = 0
	}
	if task.TaskType == "support" && task.Status == "completed" {
		task.Output.Supports = 1
	}
	if task.TaskType == "support" && task.Status != "completed" {
		task.Output.Supports = 0
	}
	a.tasks[idx] = task
	// An agentId change here moves the task between Agent project lists.
	a.syncAgentProjectsLocked()
	return a.tasks[idx], a.saveLocked()
}

func (a *App) UpdateTaskStatus(id, status string) (Task, error) {
	if !validStatus(status) {
		return Task{}, fmt.Errorf("无效的任务状态")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.indexOf(id)
	if idx < 0 {
		return Task{}, fmt.Errorf("任务不存在")
	}
	a.tasks[idx].Status, a.tasks[idx].UpdatedAt = status, time.Now().Format(time.RFC3339)
	if status == "completed" && a.tasks[idx].CompletedAt == "" {
		a.tasks[idx].CompletedAt = time.Now().Format("2006-01-02")
	}
	if status != "completed" {
		a.tasks[idx].CompletedAt = ""
	}
	if a.tasks[idx].TaskType == "support" && status == "completed" {
		a.tasks[idx].Output.Supports = 1
	}
	if a.tasks[idx].TaskType == "support" && status != "completed" {
		a.tasks[idx].Output.Supports = 0
	}
	return a.tasks[idx], a.saveLocked()
}

func (a *App) validate(input TaskInput, excludeID string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.validateLocked(input, excludeID)
}
func (a *App) validateLocked(input TaskInput, excludeID string) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Issuer) == "" || strings.TrimSpace(input.Unit) == "" {
		return fmt.Errorf("请填写单位、任务名和下发人")
	}
	if input.Kind != "main" && input.Kind != "branch" && input.Kind != "child" {
		return fmt.Errorf("请选择任务层级")
	}
	if input.TaskType != "code" && input.TaskType != "document" && input.TaskType != "support" {
		return fmt.Errorf("请选择任务类型")
	}
	// A non-code task simply drops its Agent association; only a code task's
	// Agent reference has to resolve to an existing application.
	if input.TaskType == "code" && input.AgentID != "" && !a.agentExistsLocked(input.AgentID) {
		return fmt.Errorf("关联的 Agent 不存在")
	}
	pub, e1 := time.Parse("2006-01-02", input.PublishDate)
	if e1 != nil {
		return fmt.Errorf("请选择有效的发布日期和截止日期")
	}
	if input.DueDate != "" {
		due, e2 := time.Parse("2006-01-02", input.DueDate)
		if e2 != nil || due.Before(pub) {
			return fmt.Errorf("截止日期不能早于发布日期")
		}
	}
	if input.Kind == "main" || input.Kind == "branch" {
		if input.ParentID != "" {
			return fmt.Errorf("主线任务和支线任务不可关联父任务")
		}
		return nil
	}
	if input.Kind != "child" {
		return fmt.Errorf("请选择任务层级")
	}
	if input.ParentID == "" || input.ParentID == excludeID {
		return fmt.Errorf("请选择所属主线任务")
	}
	for _, task := range a.tasks {
		if task.ID == input.ParentID && (task.Kind == "main" || task.Kind == "branch") {
			return nil
		}
	}
	return fmt.Errorf("所属主线任务不存在")
}
func validStatus(status string) bool {
	return status == "in_progress" || status == "completed" || status == "cancelled"
}
func (a *App) indexOf(id string) int {
	for i := range a.tasks {
		if a.tasks[i].ID == id {
			return i
		}
	}
	return -1
}
func (a *App) load() error {
	raw, err := os.ReadFile(a.dataPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var store DataStore
	// Legacy files were a bare task array; current files are a DataStore object.
	// Deciding by the top-level JSON token avoids guessing from which fields happen
	// to be non-empty, so a store holding only agents (or only settings) still loads.
	if isJSONArray(raw) {
		if err := json.Unmarshal(raw, &a.tasks); err != nil {
			return err
		}
		a.migrateUnitsLocked()
		return nil
	}
	if err := json.Unmarshal(raw, &store); err != nil {
		return err
	}
	a.tasks, a.units, a.agents, a.defaultUnitID, a.settings = store.Tasks, store.Units, store.Agents, store.DefaultUnitID, store.Settings
	if a.settings.BackgroundColor == "" {
		a.settings.BackgroundColor = "#f7faff"
	}
	a.migrateUnitsLocked()
	// Normalising prompts also drops keys whose task is no longer in the file.
	prunedPrompts := a.normalizeAgentPromptsLocked()
	// Data written before the relation was unified may hold an Agent-side project that never
	// reached the task; repair it once here and persist so the two views agree from now on.
	if a.repairAgentLinksLocked() || prunedPrompts {
		_ = a.saveLocked()
	}
	return nil
}

func isJSONArray(raw []byte) bool {
	for _, b := range raw {
		switch b {
		case ' ', '\t', '\r', '\n':
			continue
		case '[':
			return true
		default:
			return false
		}
	}
	return false
}
func (a *App) saveLocked() error {
	raw, err := json.MarshalIndent(DataStore{Units: a.units, DefaultUnitID: a.defaultUnitID, Tasks: a.tasks, Agents: a.agents, Settings: a.settings}, "", "  ")
	if err != nil {
		return err
	}
	temp := a.dataPath + ".tmp"
	if err = os.WriteFile(temp, raw, 0644); err != nil {
		return err
	}
	return os.Rename(temp, a.dataPath)
}

func (a *App) migrateUnitsLocked() {
	known := map[string]bool{}
	for _, u := range a.units {
		known[u.Name] = true
	}
	for _, task := range a.tasks {
		name := strings.TrimSpace(task.Unit)
		if name != "" && !known[name] {
			a.units = append(a.units, Unit{ID: uuid.NewString(), Name: name})
			known[name] = true
		}
	}
	if a.defaultUnitID == "" && len(a.units) == 1 {
		a.defaultUnitID = a.units[0].ID
	}
}

func (a *App) ExportData() error {
	a.mu.RLock()
	raw, err := json.MarshalIndent(DataStore{Units: a.units, DefaultUnitID: a.defaultUnitID, Tasks: a.tasks, Agents: a.agents, Settings: a.settings}, "", "  ")
	a.mu.RUnlock()
	if err != nil {
		return err
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{DefaultFilename: "task-manager-data.json", Filters: []runtime.FileFilter{{DisplayName: "JSON 文件", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return err
	}
	return os.WriteFile(path, raw, 0644)
}
func (a *App) ImportData() error {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Filters: []runtime.FileFilter{{DisplayName: "JSON 文件", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var store DataStore
	if isJSONArray(raw) {
		var tasks []Task
		if err = json.Unmarshal(raw, &tasks); err != nil {
			return fmt.Errorf("导入 JSON 格式无效")
		}
		store.Tasks = tasks
	} else if err = json.Unmarshal(raw, &store); err != nil {
		return fmt.Errorf("导入 JSON 格式无效")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tasks, a.units, a.agents, a.defaultUnitID, a.settings = store.Tasks, store.Units, store.Agents, store.DefaultUnitID, store.Settings
	if a.settings.BackgroundColor == "" {
		a.settings.BackgroundColor = "#f7faff"
	}
	a.migrateUnitsLocked()
	a.normalizeAgentPromptsLocked()
	// Imported files may come from the pre-unification format, so repair before persisting.
	a.repairAgentLinksLocked()
	return a.saveLocked()
}
