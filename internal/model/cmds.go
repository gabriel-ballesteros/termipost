package model

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/httpclient"
	"github.com/gabriel-ballesteros/termipost/internal/runner"
)

// sendResultMsg carries the outcome of an HTTP send.
type sendResultMsg struct {
	resp       *httpclient.Response
	unresolved []string
	err        error
}

// reqRunMsg carries the outcome of running a single request as a test.
type reqRunMsg struct{ result domain.RunResult }

// collRunMsg carries the outcome of running a whole collection.
type collRunMsg struct {
	collectionName string
	result         domain.CollectionRunResult
}

// sendCmd performs an HTTP send off the UI thread.
func sendCmd(app *App, req domain.Request) tea.Cmd {
	return func() tea.Msg {
		resp, unresolved, err := httpclient.Send(context.Background(), req, app.Resolver())
		return sendResultMsg{resp: resp, unresolved: unresolved, err: err}
	}
}

// runRequestCmd runs a single request's assertions off the UI thread.
func runRequestCmd(app *App, req domain.Request) tea.Cmd {
	return func() tea.Msg {
		return reqRunMsg{result: runner.RunRequest(context.Background(), req, app.Resolver())}
	}
}

// runCollectionCmd runs every request-with-assertions in a collection.
func runCollectionCmd(app *App, col domain.Collection) tea.Cmd {
	return func() tea.Msg {
		return collRunMsg{
			collectionName: col.Name,
			result:         runner.RunCollection(context.Background(), col, app.Resolver()),
		}
	}
}
