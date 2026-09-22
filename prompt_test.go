package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// seedAgentWithTask creates an available Agent with one code task attached to it.
func seedAgentWithTask(t *testing.T, app *App) (AgentApp, Task) {
	t.Helper()
	task := seedCodeTask(t, app, "Claude 适配", "code")
	agent, err := app.CreateAgentApp(AgentAppInput{Name: "Codex", Model: "GPT-5", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("CreateAgentApp failed: %v", err)
	}
	if len(agent.TaskIDs) != 1 {
		t.Fatalf("the task should be attached on creation, got %#v", agent.TaskIDs)
	}
	return agent, task
}

// reloadApp re-reads the same data file, standing in for a restart.
func reloadApp(t *testing.T, app *App) *App {
	t.Helper()
	reloaded := NewApp()
	reloaded.dataPath = app.dataPath
	if err := reloaded.load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	return reloaded
}

// rawStore reads the data file as-is so tests can assert on what actually got persisted.
func rawStore(t *testing.T, app *App) DataStore {
	t.Helper()
	raw, err := os.ReadFile(app.dataPath)
	if err != nil {
		t.Fatalf("reading the data file failed: %v", err)
	}
	var store DataStore
	if err := json.Unmarshal(raw, &store); err != nil {
		t.Fatalf("the data file is not valid JSON: %v", err)
	}
	return store
}

func TestPromptsRoundTripOnAgentTaskPair(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)

	saved, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{
		{Text: "先读一遍现有实现"},
		{ID: "fixed", Text: "再改接口"},
	})
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if len(saved) != 2 {
		t.Fatalf("expected 2 prompts, got %#v", saved)
	}
	if saved[0].ID == "" {
		t.Fatal("a prompt submitted without an id should be given one")
	}
	if saved[1].ID != "fixed" {
		t.Fatalf("a submitted id should be preserved, got %q", saved[1].ID)
	}

	reloaded := reloadApp(t, app)
	listed, err := reloaded.ListAgentPrompts(agent.ID, task.ID)
	if err != nil {
		t.Fatalf("ListAgentPrompts failed: %v", err)
	}
	if len(listed) != 2 || listed[0].Text != "先读一遍现有实现" || listed[1].Text != "再改接口" {
		t.Fatalf("prompts should survive a reload in order, got %#v", listed)
	}

	// The panel saves the whole list, so a shorter list means "the rest were deleted".
	trimmed, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{saved[1]})
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if len(trimmed) != 1 || trimmed[0].ID != "fixed" {
		t.Fatalf("dropping an entry should delete it, got %#v", trimmed)
	}
}

func TestBlankPromptsNeverReachTheDataFile(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)

	saved, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{
		{Text: "   "},
		{Text: "\n\t"},
		{Text: "有效内容"},
	})
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if len(saved) != 1 || saved[0].Text != "有效内容" {
		t.Fatalf("blank prompts should be dropped, got %#v", saved)
	}

	// Clearing the text box is how a user deletes a prompt, so the key must go away too.
	emptied, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{{ID: saved[0].ID, Text: "  "}})
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if len(emptied) != 0 {
		t.Fatalf("clearing the only prompt should empty the list, got %#v", emptied)
	}
	store := rawStore(t, app)
	if len(store.Agents) != 1 {
		t.Fatalf("expected 1 agent in the file, got %d", len(store.Agents))
	}
	if _, ok := store.Agents[0].Prompts[task.ID]; ok {
		t.Fatalf("an emptied prompt list should not leave a key behind: %#v", store.Agents[0].Prompts)
	}
	if store.Agents[0].Prompts == nil {
		t.Fatal("prompts should serialise as an object, not null")
	}
}

func TestPromptIDsAreGeneratedAndDeduplicated(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)

	saved, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{
		{ID: "same", Text: "第一条"},
		{ID: "same", Text: "第二条"},
	})
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if len(saved) != 2 {
		t.Fatalf("expected 2 prompts, got %#v", saved)
	}
	if saved[0].ID == saved[1].ID {
		t.Fatalf("duplicate ids would break deletion, got %#v", saved)
	}
}

func TestPromptsRejectTaskThatIsNotAttached(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)
	stranger := seedCodeTask(t, app, "别人的任务", "code")
	doc := seedCodeTask(t, app, "文档任务", "document")

	cases := []struct {
		name            string
		agentID, taskID string
	}{
		{"未关联的代码任务", agent.ID, stranger.ID},
		{"非代码任务", agent.ID, doc.ID},
		{"不存在的任务", agent.ID, "no-such-task"},
		{"没有指定任务", agent.ID, "   "},
		{"不存在的应用", "no-such-agent", task.ID},
	}
	for _, item := range cases {
		if _, err := app.UpdateAgentPrompts(item.agentID, item.taskID, []AgentPrompt{{Text: "x"}}); err == nil {
			t.Fatalf("%s：应当被拒绝", item.name)
		}
		if _, err := app.ListAgentPrompts(item.agentID, item.taskID); err == nil {
			t.Fatalf("%s：读取也应当被拒绝", item.name)
		}
	}
}

func TestDeletingTaskClearsItsPrompts(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)
	if _, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{{Text: "留着"}}); err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}

	if err := app.DeleteTask(task.ID); err != nil {
		t.Fatalf("DeleteTask failed: %v", err)
	}

	store := rawStore(t, app)
	for _, item := range store.Agents {
		if len(item.Prompts) != 0 {
			t.Fatalf("prompts must not outlive their task, got %#v", item.Prompts)
		}
	}
	// A leftover key would also be pruned on the next load; this proves the write path does it.
	if store.Tasks == nil {
		t.Fatal("tasks should serialise as an array")
	}
}

func TestUnlinkingATaskKeepsItsPrompts(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)
	if _, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{{Text: "保留我"}}); err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}

	// Unchecking the task in the Agent panel detaches it without deleting anything.
	if _, err := app.UpdateAgentApp(agent.ID, AgentAppInput{Name: agent.Name, Model: agent.Model, Available: true}); err != nil {
		t.Fatalf("UpdateAgentApp failed: %v", err)
	}
	detached := app.ListAgentApps()
	if len(detached[0].TaskIDs) != 0 {
		t.Fatalf("the task should be detached, got %#v", detached[0].TaskIDs)
	}
	if len(detached[0].Prompts[task.ID]) != 1 {
		t.Fatalf("prompts should survive unlinking, got %#v", detached[0].Prompts)
	}
	if _, err := app.ListAgentPrompts(agent.ID, task.ID); err == nil {
		t.Fatal("a detached task must not be a valid prompt target")
	}

	// Re-attaching it brings the prompts back.
	if _, err := app.UpdateAgentApp(agent.ID, AgentAppInput{Name: agent.Name, Model: agent.Model, TaskIDs: []string{task.ID}, Available: true}); err != nil {
		t.Fatalf("UpdateAgentApp failed: %v", err)
	}
	restored, err := app.ListAgentPrompts(agent.ID, task.ID)
	if err != nil {
		t.Fatalf("ListAgentPrompts failed after re-attaching: %v", err)
	}
	if len(restored) != 1 || restored[0].Text != "保留我" {
		t.Fatalf("re-attaching should bring the prompts back, got %#v", restored)
	}
}

func TestUpdateAgentAppLeavesPromptsAlone(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)
	if _, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{{Text: "别弄丢我"}}); err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}

	updated, err := app.UpdateAgentApp(agent.ID, AgentAppInput{Name: "改名了", Model: "GPT-5.1", TaskIDs: []string{task.ID}, Available: true})
	if err != nil {
		t.Fatalf("UpdateAgentApp failed: %v", err)
	}
	if updated.Name != "改名了" {
		t.Fatalf("the rename should apply, got %q", updated.Name)
	}
	if len(updated.Prompts[task.ID]) != 1 || updated.Prompts[task.ID][0].Text != "别弄丢我" {
		t.Fatalf("UpdateAgentApp must not touch prompts, got %#v", updated.Prompts)
	}
}

func TestPromptLimitsAreEnforced(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)

	many := make([]AgentPrompt, 0, maxPromptsPerPair+10)
	for i := 0; i < maxPromptsPerPair+10; i++ {
		many = append(many, AgentPrompt{Text: fmt.Sprintf("prompt %d", i)})
	}
	saved, err := app.UpdateAgentPrompts(agent.ID, task.ID, many)
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if len(saved) != maxPromptsPerPair {
		t.Fatalf("expected the list to be capped at %d, got %d", maxPromptsPerPair, len(saved))
	}

	long, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{{Text: strings.Repeat("阿", maxPromptLength+50)}})
	if err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}
	if got := len([]rune(long[0].Text)); got != maxPromptLength {
		t.Fatalf("a long prompt should be truncated to %d runes, got %d", maxPromptLength, got)
	}
}

func TestLegacyAgentsWithoutPromptsLoadCleanly(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)

	// Strip the field entirely, as a file written before 005 would look.
	raw, err := os.ReadFile(app.dataPath)
	if err != nil {
		t.Fatalf("reading the data file failed: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshalling the data file failed: %v", err)
	}
	agents, ok := doc["agents"].([]any)
	if !ok || len(agents) != 1 {
		t.Fatalf("expected one agent in the file, got %#v", doc["agents"])
	}
	first, ok := agents[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected agent shape: %#v", agents[0])
	}
	if _, ok := first["prompts"]; !ok {
		t.Fatal("expected the prompts field to exist before stripping it")
	}
	delete(first, "prompts")
	legacy, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("re-marshalling the legacy file failed: %v", err)
	}
	if err := os.WriteFile(app.dataPath, legacy, 0644); err != nil {
		t.Fatalf("writing the legacy file failed: %v", err)
	}

	reloaded := reloadApp(t, app)
	listed := reloaded.ListAgentApps()
	if len(listed) != 1 || listed[0].Name != agent.Name {
		t.Fatalf("the agent should load, got %#v", listed)
	}
	if listed[0].Prompts == nil {
		t.Fatal("a missing prompts field should become an empty map")
	}
	prompts, err := reloaded.ListAgentPrompts(agent.ID, task.ID)
	if err != nil {
		t.Fatalf("ListAgentPrompts failed: %v", err)
	}
	if len(prompts) != 0 {
		t.Fatalf("expected no prompts, got %#v", prompts)
	}
	// And the normalised empty map should be persisted back, not null.
	store := rawStore(t, reloaded)
	if store.Agents[0].Prompts == nil {
		t.Fatal("the normalised map should be written back as an object")
	}
}

func TestPromptsOfDeletedTaskArePrunedOnLoad(t *testing.T) {
	app := newTestApp(t)
	agent, task := seedAgentWithTask(t, app)
	if _, err := app.UpdateAgentPrompts(agent.ID, task.ID, []AgentPrompt{{Text: "孤儿"}}); err != nil {
		t.Fatalf("UpdateAgentPrompts failed: %v", err)
	}

	// Hand-craft a file whose prompt key points at a task that is not in it.
	store := rawStore(t, app)
	store.Tasks = nil
	raw, err := json.Marshal(store)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if err := os.WriteFile(app.dataPath, raw, 0644); err != nil {
		t.Fatalf("writing the crafted file failed: %v", err)
	}

	reloaded := reloadApp(t, app)
	listed := reloaded.ListAgentApps()
	if len(listed) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(listed))
	}
	if len(listed[0].Prompts) != 0 {
		t.Fatalf("a prompt keyed to a missing task should be pruned, got %#v", listed[0].Prompts)
	}
}
