package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
}

type Unit struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type DataStore struct {
	Units         []Unit `json:"units"`
	DefaultUnitID string `json:"defaultUnitId"`
	Tasks         []Task `json:"tasks"`
}

type App struct {
	ctx           context.Context
	tasks         []Task
	units         []Unit
	defaultUnitID string
	dataPath      string
	mu            sync.RWMutex
}

func NewApp() *App { return &App{} }

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

func (a *App) CreateTask(input TaskInput) (Task, error) {
	if err := a.validate(input, ""); err != nil {
		return Task{}, err
	}
	now := time.Now().Format(time.RFC3339)
	task := Task{ID: uuid.NewString(), Name: strings.TrimSpace(input.Name), Kind: input.Kind, ParentID: input.ParentID, Issuer: strings.TrimSpace(input.Issuer), Unit: strings.TrimSpace(input.Unit), PublishDate: input.PublishDate, DueDate: input.DueDate, TaskType: input.TaskType, Status: "in_progress", CreatedAt: now, UpdatedAt: now}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tasks = append(a.tasks, task)
	return task, a.saveLocked()
}

func (a *App) UpdateTask(task Task) (Task, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.indexOf(task.ID)
	if idx < 0 {
		return Task{}, fmt.Errorf("任务不存在")
	}
	input := TaskInput{Name: task.Name, Kind: task.Kind, ParentID: task.ParentID, Issuer: task.Issuer, Unit: task.Unit, PublishDate: task.PublishDate, DueDate: task.DueDate, TaskType: task.TaskType}
	if err := a.validateLocked(input, task.ID); err != nil {
		return Task{}, err
	}
	old := a.tasks[idx]
	task.Name, task.Issuer = strings.TrimSpace(task.Name), strings.TrimSpace(task.Issuer)
	task.CreatedAt, task.UpdatedAt = old.CreatedAt, time.Now().Format(time.RFC3339)
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
	a.tasks[idx] = task
	return task, a.saveLocked()
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
		if task.ID == input.ParentID && task.Kind == "main" {
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
	if err := json.Unmarshal(raw, &store); err == nil && (store.Tasks != nil || store.Units != nil) {
		a.tasks, a.units, a.defaultUnitID = store.Tasks, store.Units, store.DefaultUnitID
		a.migrateUnitsLocked()
		return nil
	}
	// Backward compatibility with the original array-only JSON format.
	if err := json.Unmarshal(raw, &a.tasks); err != nil {
		return err
	}
	a.migrateUnitsLocked()
	return nil
}
func (a *App) saveLocked() error {
	raw, err := json.MarshalIndent(DataStore{Units: a.units, DefaultUnitID: a.defaultUnitID, Tasks: a.tasks}, "", "  ")
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
	raw, err := json.MarshalIndent(DataStore{Units: a.units, DefaultUnitID: a.defaultUnitID, Tasks: a.tasks}, "", "  ")
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
	if err = json.Unmarshal(raw, &store); err != nil || (store.Tasks == nil && store.Units == nil) {
		var tasks []Task
		if e := json.Unmarshal(raw, &tasks); e != nil {
			return fmt.Errorf("导入 JSON 格式无效")
		}
		store.Tasks = tasks
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tasks, a.units, a.defaultUnitID = store.Tasks, store.Units, store.DefaultUnitID
	a.migrateUnitsLocked()
	return a.saveLocked()
}
