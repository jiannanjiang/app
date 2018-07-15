package app_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/murlokswarm/app"
	"github.com/murlokswarm/app/drivers/test"
	"github.com/murlokswarm/app/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImport(t *testing.T) {
	app.Import(&tests.Foo{})

	defer func() { recover() }()
	app.Import(tests.NoPointerCompo{})
}

func TestApp(t *testing.T) {
	var d app.Driver
	var newPage func(c app.PageConfig) app.Page

	output := &bytes.Buffer{}
	app.Loggers = []app.Logger{app.NewLogger(output, output, true, true)}

	app.Import(&tests.Foo{})
	app.Import(&tests.Bar{})

	ctx, cancel := context.WithCancel(context.Background())

	onRun := func() {
		rd := app.RunningDriver()
		require.NotNil(t, rd, "driver not set")
		assert.NotEmpty(t, app.Name())
		assert.Equal(t, filepath.Join("resources", "hello", "world"), app.Resources("hello", "world"))
		assert.Equal(t, filepath.Join("storage", "hello", "world"), app.Storage("hello", "world"))

		// Window:
		win := app.NewWindow(app.WindowConfig{
			DefaultURL: "tests.foo",
		})

		compo := win.Compo()
		require.NotNil(t, compo)
		app.Render(compo)

		// Page:
		page := newPage(app.PageConfig{
			DefaultURL: "tests.foo",
		})

		compo = page.Compo()
		require.NotNil(t, compo)
		app.Render(compo)

		// Menu:
		menu := app.NewContextMenu(app.MenuConfig{
			DefaultURL: "tests.bar",
		})

		compo = menu.Compo()
		require.NotNil(t, compo)
		app.Render(compo)

		elem := app.ElemByCompo(compo)
		assert.Equal(t, menu.ID(), elem.ID())

		app.NewFilePanel(app.FilePanelConfig{})
		app.NewSaveFilePanel(app.SaveFilePanelConfig{})
		app.NewShare("Hello world")
		app.NewNotification(app.NotificationConfig{})
		app.MenuBar()

		// Status menu:
		statusMenu := app.NewStatusMenu(app.StatusMenuConfig{})

		err := statusMenu.Load("tests.bar")
		require.NoError(t, err)

		compo = statusMenu.Compo()
		require.NotNil(t, compo)
		app.Render(compo)

		elem = app.ElemByCompo(compo)
		assert.Equal(t, statusMenu.ID(), elem.ID())

		err = statusMenu.SetText("test")
		assert.NoError(t, err)

		err = statusMenu.SetIcon(filepath.Join("internal", "tests", "resources", "logo.png"))
		assert.NoError(t, err)

		err = statusMenu.Close()
		assert.NoError(t, err)

		// Dock:
		dockTile := app.Dock()

		err = dockTile.Load("tests.bar")
		require.NoError(t, err)

		compo = dockTile.Compo()
		require.NotNil(t, compo)
		app.Render(compo)

		elem = app.ElemByCompo(compo)
		assert.Equal(t, dockTile.ID(), elem.ID())

		err = dockTile.SetBadge("42")
		assert.NoError(t, err)

		err = dockTile.SetIcon(filepath.Join("internal", "tests", "resources", "logo.png"))
		assert.NoError(t, err)

		// CSS resources:
		assert.Len(t, app.CSSResources(), 0)

		os.MkdirAll(app.Resources("css", "sub"), 0777)
		os.Create(app.Resources("css", "test.css"))
		os.Create(app.Resources("css", "test.scss"))
		os.Create(app.Resources("css", "sub", "sub.css"))
		defer os.RemoveAll(app.Resources())

		assert.Contains(t, app.CSSResources(), app.Resources("css", "test.css"))
		assert.NotContains(t, app.CSSResources(), app.Resources("css", "test.scss"))
		assert.Contains(t, app.CSSResources(), app.Resources("css", "sub", "sub.css"))

		app.CallOnUIGoroutine(func() {
			t.Log("CallOnUIGoroutine")
		})

		cancel()
	}

	dtest := &test.Driver{
		Ctx:   ctx,
		OnRun: onRun,
	}
	d = dtest

	newPage = func(c app.PageConfig) app.Page {
		app.NewPage(c)
		return dtest.Page
	}

	err := app.Run(d)
	require.NoError(t, err)

	t.Log(output.String())
}

func TestAppError(t *testing.T) {
	var d app.Driver

	output := &bytes.Buffer{}
	app.Loggers = []app.Logger{app.NewLogger(output, output, true, true)}

	ctx, cancel := context.WithCancel(context.Background())

	onRun := func() {
		defer cancel()

		app.Render(nil)

		// Window:
		win, err := d.NewWindow(app.WindowConfig{})
		require.NoError(t, err)

		err = win.Load("")
		assert.Error(t, err)

		err = win.Render(nil)
		assert.Error(t, err)

		err = win.Reload()
		assert.Error(t, err)

		err = win.Previous()
		assert.Error(t, err)

		err = win.Next()
		assert.Error(t, err)

		err = win.Close()
		assert.Error(t, err)

		err = win.Move(0, 0)
		assert.Error(t, err)

		err = win.Center()
		assert.Error(t, err)

		err = win.Resize(0, 0)
		assert.Error(t, err)

		err = win.Focus()
		assert.Error(t, err)

		err = win.ToggleFullScreen()
		assert.Error(t, err)

		err = win.ToggleMinimize()
		assert.Error(t, err)

		// Menu:
		var menu app.Menu
		menu, err = d.NewContextMenu(app.MenuConfig{})
		require.NoError(t, err)

		err = menu.Load("")
		assert.Error(t, err)

		err = menu.Render(nil)
		assert.Error(t, err)

		// Status Bar:
		statusMenu := app.NewStatusMenu(app.StatusMenuConfig{
			Text: "test",
		})

		err = statusMenu.Load("")
		assert.Error(t, err)

		err = statusMenu.Render(nil)
		assert.Error(t, err)

		err = statusMenu.SetText("")
		assert.Error(t, err)

		err = statusMenu.SetIcon("")
		assert.Error(t, err)

		err = statusMenu.Close()
		assert.Error(t, err)

		// Dock tile:
		dockTile := app.Dock()

		err = dockTile.Load("")
		assert.Error(t, err)

		err = dockTile.Render(nil)
		assert.Error(t, err)

		err = dockTile.SetIcon("")
		assert.Error(t, err)

		err = dockTile.SetBadge("")
		assert.Error(t, err)
	}

	dtest := &test.Driver{
		Ctx:             ctx,
		SimulateErr:     true,
		SimulateElemErr: true,
		OnRun:           onRun,
	}
	d = app.Logs()(dtest)

	err := app.Run(d)
	assert.Error(t, err)

	app.NewWindow(app.WindowConfig{})
	app.NewPage(app.PageConfig{})
	app.NewContextMenu(app.MenuConfig{})
	app.NewFilePanel(app.FilePanelConfig{})
	app.NewSaveFilePanel(app.SaveFilePanelConfig{})
	app.NewShare(nil)
	app.NewNotification(app.NotificationConfig{})
	app.MenuBar()
	app.NewStatusMenu(app.StatusMenuConfig{})
	app.Dock()

	dtest.SimulateErr = false
	err = app.Run(d)
	require.NoError(t, err)

	t.Log(output.String())
}
