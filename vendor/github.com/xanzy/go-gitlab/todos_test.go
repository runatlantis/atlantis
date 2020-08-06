package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListTodos(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/todos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		mustWriteHTTPResponse(t, w, "testdata/list_todos.json")
	})

	opts := &ListTodosOptions{ListOptions: ListOptions{PerPage: 2}}
	todos, _, err := client.Todos.ListTodos(opts)

	require.NoError(t, err)

	want := []*Todo{{ID: 1, State: "pending", Target: TodoTarget{ID: 1, ApprovalsBeforeMerge: 2}}, {ID: 2, State: "pending", Target: TodoTarget{ID: 2}}}
	require.Equal(t, want, todos)
}

func TestMarkAllTodosAsDone(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/todos/mark_as_done", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		w.WriteHeader(http.StatusNoContent)
	})

	_, err := client.Todos.MarkAllTodosAsDone()
	require.NoError(t, err)
}

func TestMarkTodoAsDone(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/todos/1/mark_as_done", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
	})

	_, err := client.Todos.MarkTodoAsDone(1)
	require.NoError(t, err)
}
