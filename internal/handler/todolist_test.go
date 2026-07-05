package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"oat431/tiny-mchawa-api/internal/model"
	"oat431/tiny-mchawa-api/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTodolist(t *testing.T) {
	app, _, _ := testutil.NewTestApp()

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"valid", map[string]string{"title": "Grocery", "sourceService": "todolist"}, 201},
		{"missing title", map[string]string{"sourceService": "todolist"}, 400},
		{"missing sourceService", map[string]string{"title": "Grocery"}, 400},
		{"empty body", map[string]string{}, 400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/todolists", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestGetTodolist(t *testing.T) {
	app, todoStore, _ := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Test", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{"existing", tl.ID, 200},
		{"not found", "nonexistent", 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/todolists/"+tt.id, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestGetTodolist_Response(t *testing.T) {
	app, todoStore, _ := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Grocery", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	req := httptest.NewRequest("GET", "/api/v1/todolists/"+tl.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	b, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(b, &result)
	assert.Nil(t, result["error"])

	data := result["data"].(map[string]any)
	assert.Equal(t, "Grocery", data["title"])
	assert.Equal(t, "pending", data["status"])
	assert.Equal(t, testutil.TestUserID, data["ownedBy"])
}

func TestListTodolists(t *testing.T) {
	app, todoStore, _ := testutil.NewTestApp()

	todoStore.Create(&model.Todolist{Title: "List 1", OwnedBy: testutil.TestUserID, SourceService: "todolist"})
	todoStore.Create(&model.Todolist{Title: "List 2", OwnedBy: testutil.TestUserID, SourceService: "blog"})

	req := httptest.NewRequest("GET", "/api/v1/todolists", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	b, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(b, &result)
	assert.Nil(t, result["error"])
	assert.NotNil(t, result["meta"])

	data := result["data"].([]any)
	assert.Len(t, data, 2)
}

func TestUpdateTodolist(t *testing.T) {
	app, todoStore, _ := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Original", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	body, _ := json.Marshal(map[string]string{"title": "Updated"})
	req := httptest.NewRequest("PUT", "/api/v1/todolists/"+tl.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	b, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(b, &result)
	data := result["data"].(map[string]any)
	assert.Equal(t, "Updated", data["title"])
}

func TestDeleteTodolist(t *testing.T) {
	app, todoStore, _ := testutil.NewTestApp()

	tl := &model.Todolist{Title: "To Delete", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	req := httptest.NewRequest("DELETE", "/api/v1/todolists/"+tl.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestCreateTask(t *testing.T) {
	app, todoStore, _ := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Parent", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	tests := []struct {
		name       string
		todolistID string
		body       map[string]string
		wantStatus int
	}{
		{"valid", tl.ID, map[string]string{"title": "Buy milk"}, 201},
		{"missing title", tl.ID, map[string]string{}, 400},
		{"todolist not found", "nonexistent", map[string]string{"title": "Task"}, 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/todolists/"+tt.todolistID+"/tasks", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestListTasks(t *testing.T) {
	app, todoStore, taskStore := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Parent", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	taskStore.Create(&model.Task{TodolistID: tl.ID, Title: "Task 1", Status: "pending"})
	taskStore.Create(&model.Task{TodolistID: tl.ID, Title: "Task 2", Status: "done"})

	req := httptest.NewRequest("GET", "/api/v1/todolists/"+tl.ID+"/tasks", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	b, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(b, &result)
	assert.Nil(t, result["error"])
	data := result["data"].([]any)
	assert.Len(t, data, 2)
}

func TestUpdateTask(t *testing.T) {
	app, todoStore, taskStore := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Parent", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	task := &model.Task{TodolistID: tl.ID, Title: "Task", Status: "pending"}
	taskStore.Create(task)

	body, _ := json.Marshal(map[string]string{"title": "Updated", "status": "done"})
	req := httptest.NewRequest("PUT", "/api/v1/tasks/"+task.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDeleteTask(t *testing.T) {
	app, todoStore, taskStore := testutil.NewTestApp()

	tl := &model.Todolist{Title: "Parent", OwnedBy: testutil.TestUserID, SourceService: "todolist"}
	todoStore.Create(tl)

	task := &model.Task{TodolistID: tl.ID, Title: "To Delete", Status: "pending"}
	taskStore.Create(task)

	req := httptest.NewRequest("DELETE", "/api/v1/tasks/"+task.ID, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
}

func TestHealthEndpoint(t *testing.T) {
	app, _, _ := testutil.NewTestApp()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestNotFoundRoute(t *testing.T) {
	app, _, _ := testutil.NewTestApp()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}
