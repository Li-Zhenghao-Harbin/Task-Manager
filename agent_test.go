package main

import (
	"path/filepath"
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

func TestAgentProjectsMustBeCodeTasks(t *testing.T) {
	app := newTestApp(t)
	code := seedCodeTask(t, app, "代码任务", "code")
	doc := seedCodeTask(t, app, "文档任务", "document")

	if _, err := app.CreateAgentApp(AgentAppInput{Name: "A", Model: "m", ProjectIDs: []string{doc.ID}, Available: true}); err == nil {
		t.Fatal("expected a non-code task to be rejected as an associated project")
	}

	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", ProjectIDs: []string{code.ID, code.ID, ""}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if len(agent.ProjectIDs) != 1 || agent.ProjectIDs[0] != code.ID {
		t.Fatalf("project ids should be de-duplicated, got %#v", agent.ProjectIDs)
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

func TestDeletingTaskDetachesItFromAgentProjects(t *testing.T) {
	app := newTestApp(t)
	task := seedCodeTask(t, app, "代码任务", "code")
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Claude Code", Model: "Opus", ProjectIDs: []string{task.ID}, Available: true})
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
	if len(listed[0].ProjectIDs) != 0 {
		t.Fatalf("deleted task should be removed from the project list, got %#v", listed[0].ProjectIDs)
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
	if listed[0].ProjectIDs == nil {
		t.Fatal("projectIds should serialise as an empty array, not null")
	}
}
