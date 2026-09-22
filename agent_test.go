package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newTestApp returns an App bound to a throwaway data file so saves are harmless.
func newTestApp(t *testing.T) *App {
	t.Helper()
	app := NewApp()
	app.dataPath = filepath.Join(t.TempDir(), "task-manager-data.json")
	return app
}

func seedCodeTask(t *testing.T, app *App, name, taskType string) Task {
	t.Helper()
	task, err := app.CreateTask(TaskInput{
		Name:        name,
		Kind:        "main",
		Issuer:      "本人",
		Unit:        "研发",
		PublishDate: "2026-09-01",
		DueDate:     "2026-09-30",
		TaskType:    taskType,
	})
	if err != nil {
		t.Fatalf("CreateTask(%s) failed: %v", name, err)
	}
	return task
}

func TestAgentRequiresNameAndModel(t *testing.T) {
	app := newTestApp(t)
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "  ", Model: "claude"}); err == nil {
		t.Fatal("expected an error when the application name is blank")
	}
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: ""}); err == nil {
		t.Fatal("expected an error when the model name is blank")
	}
}

func TestAgentTasksMustBeCodeTasks(t *testing.T) {
	app := newTestApp(t)
	code := seedCodeTask(t, app, "代码任务", "code")
	doc := seedCodeTask(t, app, "文档任务", "document")

	if _, err := app.CreateAgentApp(AgentAppInput{Name: "A", Model: "m", TaskIDs: []string{doc.ID}, Available: true}); err == nil {
		t.Fatal("expected a non-code task to be rejected as an attachable task")
	}

	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", TaskIDs: []string{code.ID, code.ID, ""}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if len(agent.TaskIDs) != 1 || agent.TaskIDs[0] != code.ID {
		t.Fatalf("task ids should be de-duplicated, got %#v", agent.TaskIDs)
	}
}

func TestAgentResetTimeMustHaveMinutePrecision(t *testing.T) {
	app := newTestApp(t)
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "A", Model: "m", Available: false, ResetAt: "2026-09-21 17:30"}); err == nil {
		t.Fatal("expected a space-separated reset time to be rejected")
	}
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "A", Model: "m", Available: false, ResetAt: "2026-09-21T17:30"}); err != nil {
		t.Fatalf("minute-precision reset time should be accepted, got %v", err)
	}
}

func TestAgentAutoRestoresAvailabilityAfterReset(t *testing.T) {
	app := newTestApp(t)
	past := time.Now().Add(-time.Hour).Format(agentResetLayout)
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: false, ResetAt: past})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if agent.Available {
		t.Fatal("agent should start as unavailable")
	}

	listed := app.ListAgentApps()
	if len(listed) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(listed))
	}
	if !listed[0].Available {
		t.Fatal("agent should auto-restore once the reset time has passed")
	}
	if listed[0].ResetAt != "" {
		t.Fatalf("reset time should be cleared, got %q", listed[0].ResetAt)
	}
}

func TestAgentKeepsResetTimeWhenStillInFuture(t *testing.T) {
	app := newTestApp(t)
	future := time.Now().Add(3 * time.Hour).Format(agentResetLayout)
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: false, ResetAt: future})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	listed := app.ListAgentApps()
	if listed[0].Available {
		t.Fatal("agent should stay unavailable before the reset time")
	}
	if listed[0].ResetAt != agent.ResetAt {
		t.Fatalf("reset time should be preserved, got %q", listed[0].ResetAt)
	}
}

func TestAvailableAgentDropsResetTime(t *testing.T) {
	app := newTestApp(t)
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: true, ResetAt: time.Now().Add(time.Hour).Format(agentResetLayout)})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if agent.ResetAt != "" {
		t.Fatalf("an available agent should not keep a reset time, got %q", agent.ResetAt)
	}
}

func TestOnlyCodeTasksKeepAgentAssociation(t *testing.T) {
	app := newTestApp(t)
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}

	code, err := app.CreateTask(TaskInput{Name: "代码任务", Kind: "main", Issuer: "本人", Unit: "研发", PublishDate: "2026-09-01", DueDate: "2026-09-30", TaskType: "code", AgentID: agent.ID})
	if err != nil {
		t.Fatalf("code task with an agent should be accepted: %v", err)
	}
	if code.AgentID != agent.ID {
		t.Fatalf("code task should keep agentId, got %q", code.AgentID)
	}

	doc, err := app.CreateTask(TaskInput{Name: "文档任务", Kind: "main", Issuer: "本人", Unit: "研发", PublishDate: "2026-09-01", DueDate: "2026-09-30", TaskType: "document", AgentID: agent.ID})
	if err != nil {
		t.Fatalf("non-code task with an agent should still be created: %v", err)
	}
	if doc.AgentID != "" {
		t.Fatalf("non-code task should drop agentId, got %q", doc.AgentID)
	}

	// Switching an existing code task away from "code" clears the association.
	code.TaskType = "document"
	updated, err := app.UpdateTask(code)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if updated.AgentID != "" {
		t.Fatalf("switching to a non-code type should clear agentId, got %q", updated.AgentID)
	}
}

func TestUnknownAgentIsRejected(t *testing.T) {
	app := newTestApp(t)
	_, err := app.CreateTask(TaskInput{Name: "代码任务", Kind: "main", Issuer: "本人", Unit: "研发", PublishDate: "2026-09-01", DueDate: "2026-09-30", TaskType: "code", AgentID: "missing"})
	if err == nil {
		t.Fatal("expected an unknown agent id to be rejected")
	}
}

func TestDeletingAgentClearsTaskReferences(t *testing.T) {
	app := newTestApp(t)
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	task, err := app.CreateTask(TaskInput{Name: "代码任务", Kind: "main", Issuer: "本人", Unit: "研发", PublishDate: "2026-09-01", DueDate: "2026-09-30", TaskType: "code", AgentID: agent.ID})
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	if err := app.DeleteAgentApp(agent.ID); err != nil {
		t.Fatalf("DeleteAgentApp failed: %v", err)
	}
	if len(app.ListAgentApps()) != 0 {
		t.Fatal("agent should be gone")
	}
	if got := app.ListTasks()[0]; got.ID != task.ID || got.AgentID != "" {
		t.Fatalf("task should keep existing but lose its agentId, got %#v", got)
	}
}

func TestDeletingTaskDetachesItFromAgentTasks(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "代码任务", "code")
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if err := app.DeleteTask(task.ID); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}
	listed := app.ListAgentApps()
	if len(listed) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(listed))
	}
	if len(listed[0].TaskIDs) != 0 {
		t.Fatalf("deleted task should be removed from the task list, got %#v", listed[0].TaskIDs)
	}
	if listed[0].ID != agent.ID {
		t.Fatalf("unexpected agent id %q", listed[0].ID)
	}
}

func TestAgentPersistsAcrossReload(t *testing.T) {
	app := newTestApp(t)
	reset := time.Now().Add(2 * time.Hour).Format(agentResetLayout)
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: false, ResetAt: reset}); err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}

	reloaded := NewApp()
	reloaded.dataPath = app.dataPath
	if err := reloaded.load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	listed := reloaded.ListAgentApps()
	if len(listed) != 1 || listed[0].Name != "Claude Code" || listed[0].Model != "Opus" {
		t.Fatalf("agent should survive a reload, got %#v", listed)
	}
	if listed[0].Available || listed[0].ResetAt != reset {
		t.Fatalf("availability and reset time should survive a reload, got %#v", listed[0])
	}
	if listed[0].TaskIDs == nil {
		t.Fatal("taskIds should serialise as an empty array, not null")
	}
}

// ---------- unified relation: Agent.TaskIDs mirrors Task.AgentID ----------

func TestSelectingTaskAssignsTheAgent(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "代码任务", "code")
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}

	updated, err := app.UpdateAgentApp(agent.ID, AgentAppInput{Name: "Claude Code", Model: "Opus", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("UpdateAgentApp failed: %v", err)
	}
	if len(updated.TaskIDs) != 1 || updated.TaskIDs[0] != task.ID {
		t.Fatalf("task list should hold the selected task, got %#v", updated.TaskIDs)
	}
	if got := app.ListTasks()[0].AgentID; got != agent.ID {
		t.Fatalf("checking a task must point the task at this agent, got %q", got)
	}

	// Unchecking the task must detach the task again.
	cleared, err := app.UpdateAgentApp(agent.ID, AgentAppInput{Name: "Claude Code", Model: "Opus", Available: true})
	if err != nil {
		t.Fatalf("UpdateAgentApp failed: %v", err)
	}
	if len(cleared.TaskIDs) != 0 {
		t.Fatalf("task list should be empty, got %#v", cleared.TaskIDs)
	}
	if got := app.ListTasks()[0].AgentID; got != "" {
		t.Fatalf("unchecking a task must clear the task association, got %q", got)
	}
}

func TestSelectingAgentOnTaskUpdatesAgentTasks(t *testing.T) {
	app := newTestApp(t)
	first := seedCodeTask(t, app, "任务一", "code")
	second := seedCodeTask(t, app, "任务二", "code")
	alpha, err := app.CreateAgentApp(AgentAppInput{Name: "Alpha", Model: "m1", Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	beta, err := app.CreateAgentApp(AgentAppInput{Name: "Beta", Model: "m2", Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}

	first.AgentID = alpha.ID
	if _, err = app.UpdateTask(first); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	second.AgentID = alpha.ID
	if _, err = app.UpdateTask(second); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if got := agentTasksOf(t, app, alpha.ID); len(got) != 2 {
		t.Fatalf("alpha should own both tasks, got %#v", got)
	}

	// Moving one task to another Agent must take it out of the first Agent's list.
	second.AgentID = beta.ID
	if _, err = app.UpdateTask(second); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if got := agentTasksOf(t, app, alpha.ID); len(got) != 1 || got[0] != first.ID {
		t.Fatalf("alpha should keep only the first task, got %#v", got)
	}
	if got := agentTasksOf(t, app, beta.ID); len(got) != 1 || got[0] != second.ID {
		t.Fatalf("beta should own the second task, got %#v", got)
	}
}

func TestSwitchingTaskTypeDetachesItFromAgentTasks(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "代码任务", "code")
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Alpha", Model: "m1", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if len(agent.TaskIDs) != 1 {
		t.Fatalf("the agent should list the task, got %#v", agent.TaskIDs)
	}

	task.TaskType = "document"
	updated, err := app.UpdateTask(task)
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}
	if updated.AgentID != "" {
		t.Fatalf("a non-code task must drop agentId, got %q", updated.AgentID)
	}
	if got := agentTasksOf(t, app, agent.ID); len(got) != 0 {
		t.Fatalf("the task should leave the agent task list, got %#v", got)
	}
}

func TestSharedTaskFallsBackToTaskSide(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "共享任务", "code")
	alpha, err := app.CreateAgentApp(AgentAppInput{Name: "Alpha", Model: "m1", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	// A second Agent claiming the same task must not steal it from the task-side pointer.
	beta, err := app.CreateAgentApp(AgentAppInput{Name: "Beta", Model: "m2", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if got := agentTasksOf(t, app, alpha.ID); len(got) != 0 {
		t.Fatalf("alpha should lose the task since beta claimed it last, got %#v", got)
	}
	if got := agentTasksOf(t, app, beta.ID); len(got) != 1 || got[0] != task.ID {
		t.Fatalf("beta should own the shared task, got %#v", got)
	}
	if got := app.ListTasks()[0].AgentID; got != beta.ID {
		t.Fatalf("the task should point at exactly one agent, got %q", got)
	}
}

func TestLegacyAgentSideLinksAreRepairedOnLoad(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "Codex 关联的任务", "code")
	other := seedCodeTask(t, app, "已有归属的任务", "code")
	owner, err := app.CreateAgentApp(AgentAppInput{Name: "Owner", Model: "m", Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	other.AgentID = owner.ID
	if _, err = app.UpdateTask(other); err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Reproduce the pre-unification file: an Agent lists tasks that never reached the
	// tasks, while a task that already had an owner keeps its own pointer.
	legacy := DataStore{
		Units: app.units,
		Tasks: app.tasks,
		Agents: []AgentApp{
			{ID: "codex", Name: "Codex", Model: "GPT", TaskIDs: []string{task.ID, other.ID}, Available: true},
			{ID: "owner", Name: "Owner", Model: "m", TaskIDs: []string{other.ID}, Available: true},
		},
	}
	for i := range legacy.Tasks {
		legacy.Tasks[i].AgentID = ""
	}
	legacy.Tasks[1].AgentID = "owner"
	raw, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err = os.WriteFile(app.dataPath, raw, 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	reloaded := NewApp()
	reloaded.dataPath = app.dataPath
	if err = reloaded.load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	byID := map[string]Task{}
	for _, item := range reloaded.ListTasks() {
		byID[item.ID] = item
	}
	if got := byID[task.ID].AgentID; got != "codex" {
		t.Fatalf("the orphaned task should be attached to codex, got %q", got)
	}
	if got := byID[other.ID].AgentID; got != "owner" {
		t.Fatalf("an explicit task-side owner must not be overridden, got %q", got)
	}
	if got := agentTasksOf(t, reloaded, "codex"); len(got) != 1 || got[0] != task.ID {
		t.Fatalf("codex should own exactly the repaired task, got %#v", got)
	}
	if got := agentTasksOf(t, reloaded, "owner"); len(got) != 1 || got[0] != other.ID {
		t.Fatalf("owner should keep its task, got %#v", got)
	}

	// The repair must be idempotent and already persisted.
	again := NewApp()
	again.dataPath = app.dataPath
	if err = again.load(); err != nil {
		t.Fatalf("second load failed: %v", err)
	}
	if got := agentTasksOf(t, again, "codex"); len(got) != 1 || got[0] != task.ID {
		t.Fatalf("repair should be idempotent, got %#v", got)
	}
}

// TestLoadMigratesLegacyProjectKey covers the 015 rename: a save written when the relation was
// still called "projectIds" must keep its links, and the next save must write "taskIds" only.
func TestLoadMigratesLegacyProjectKey(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "写迁移前的存档", "code")
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Codex", Model: "GPT", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	detached := app.ListTasks()
	for i := range detached {
		detached[i].AgentID = ""
	}
	legacy := DataStore{Units: app.units, Tasks: detached, Agents: []AgentApp{{ID: agent.ID, Name: "Codex", Model: "GPT", TaskIDs: []string{task.ID}, Available: true}}}
	raw, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	// Only the key name changes: this is exactly what a pre-015 file looked like.
	raw = []byte(strings.ReplaceAll(string(raw), `"taskIds"`, `"projectIds"`))
	if err = os.WriteFile(app.dataPath, raw, 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	reloaded := NewApp()
	reloaded.dataPath = app.dataPath
	if err = reloaded.load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	byID := map[string]Task{}
	for _, item := range reloaded.ListTasks() {
		byID[item.ID] = item
	}
	if got := byID[task.ID].AgentID; got != agent.ID {
		t.Fatalf("the legacy projectIds link should survive the rename, got %q", got)
	}
	if got := agentTasksOf(t, reloaded, agent.ID); len(got) != 1 || got[0] != task.ID {
		t.Fatalf("the rebuilt task list should hold the legacy link, got %#v", got)
	}
	persisted, err := os.ReadFile(app.dataPath)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if !strings.Contains(string(persisted), `"taskIds"`) {
		t.Fatal("the migrated file should be saved under the taskIds key")
	}
	if strings.Contains(string(persisted), `"projectIds"`) {
		t.Fatal("the deprecated projectIds key should not be written back")
	}
}

// ---------- reset notifications ----------

func TestResetNotifiesWithAgentLabel(t *testing.T) {
	app := newTestApp(t)
	past := time.Now().Add(-time.Minute).Format(agentResetLayout)
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "Codex", Model: "GPT", Available: false, ResetAt: past}); err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	notifications := captureNotifications(app)

	app.checkAgentResets()

	if len(*notifications) != 1 {
		t.Fatalf("expected exactly one notification, got %#v", *notifications)
	}
	if !strings.Contains((*notifications)[0], "Codex-GPT") {
		t.Fatalf("notification should name the agent as 应用名-模型名, got %q", (*notifications)[0])
	}
	listed := app.ListAgentApps()
	if !listed[0].Available || listed[0].ResetAt != "" {
		t.Fatalf("the agent should be restored, got %#v", listed[0])
	}
	// The restore is persisted, so the next check finds nothing left to report.
	app.checkAgentResets()
	if len(*notifications) != 1 {
		t.Fatalf("a restored agent must not notify twice, got %#v", *notifications)
	}
}

func TestResetNotificationWaitsForTheDueTime(t *testing.T) {
	app := newTestApp(t)
	future := time.Now().Add(2 * time.Hour).Format(agentResetLayout)
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "Codex", Model: "GPT", Available: false, ResetAt: future}); err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	notifications := captureNotifications(app)

	app.checkAgentResets()

	if len(*notifications) != 0 {
		t.Fatalf("an agent that is not due yet must stay quiet, got %#v", *notifications)
	}
}

func TestListingAgentsStillNotifiesAfterAReset(t *testing.T) {
	app := newTestApp(t)
	past := time.Now().Add(-time.Minute).Format(agentResetLayout)
	if _, err := app.CreateAgentApp(AgentAppInput{Name: "Codex", Model: "GPT", Available: false, ResetAt: past}); err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	notifications := captureNotifications(app)

	// Opening the list before the watcher ticks must not swallow the notification.
	app.ListAgentApps()

	if len(*notifications) != 1 {
		t.Fatalf("restoring through the list should still notify, got %#v", *notifications)
	}
}

// captureNotifications replaces the notifier so tests never raise a real desktop toast.
func captureNotifications(app *App) *[]string {
	messages := &[]string{}
	app.notifier = func(title, body string) error {
		*messages = append(*messages, title+" | "+body)
		return nil
	}
	return messages
}

func agentTasksOf(t *testing.T, app *App, agentID string) []string {
	t.Helper()
	for _, agent := range app.ListAgentApps() {
		if agent.ID == agentID {
			return agent.TaskIDs
		}
	}
	t.Fatalf("agent %s not found", agentID)
	return nil
}
