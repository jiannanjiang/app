package app

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

var (
	// Loggers are the loggers used by the app.
	Loggers []Logger

	driver     Driver
	components Factory
)

func init() {
	components = NewFactory()
	components = ConcurrentFactory(components)

	events := NewEventRegistry(CallOnUIGoroutine)
	events = ConcurrentEventRegistry(events)
	DefaultEventRegistry = events

	actions := NewActionRegistry(events)
	DefaultActionRegistry = actions
}

// Import imports the component into the app.
// Components must be imported in order the be used by the app package.
// This allows components to be created dynamically when they are found into
// markup.
func Import(c Compo) {
	if _, err := components.Register(c); err != nil {
		err = errors.Wrap(err, "import component failed")
		panic(err)
	}
}

// Run runs the app with the given driver as backend.
func Run(d Driver) error {
	driver = d
	return driver.Run(components)
}

// Name returns the application name.
func Name() string {
	return driver.AppName()
}

// Resources returns the given path prefixed by the resources directory
// location.
// Resources should be used only for read only operations.
func Resources(path ...string) string {
	return driver.Resources(path...)
}

// CSSResources returns a list that contains the path of the css files located
// in the resource/css directory.
func CSSResources() []string {
	var css []string

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext := filepath.Ext(path); ext != ".css" {
			return nil
		}

		css = append(css, path)
		return nil
	}

	filepath.Walk(Resources("css"), walker)
	return css
}

// Storage returns the given path prefixed by the storage directory
// location.
func Storage(path ...string) string {
	return driver.Storage(path...)
}

// NewWindow creates and displays the window described by the given
// configuration.
func NewWindow(c WindowConfig) Window {
	w, _ := driver.NewWindow(c)
	return w
}

// NewPage creates the page described by the given configuration.
func NewPage(c PageConfig) {
	driver.NewPage(c)
}

// NewContextMenu creates and displays the context menu described by the
// given configuration.
func NewContextMenu(c MenuConfig) Menu {
	m, _ := driver.NewContextMenu(c)
	return m
}

// Render renders the given component.
// It should be called when the component appearance have to be updated.
func Render(c Compo) {
	CallOnUIGoroutine(func() {
		driver.Render(c)
	})
}

// ElemByCompo returns the element where the given component is mounted.
func ElemByCompo(c Compo) Elem {
	return driver.ElemByCompo(c)
}

// NewFilePanel creates and displays the file panel described by the given
// configuration.
func NewFilePanel(c FilePanelConfig) {
	driver.NewFilePanel(c)
}

// NewSaveFilePanel creates and displays the save file panel described by the
// given configuration.
func NewSaveFilePanel(c SaveFilePanelConfig) {
	driver.NewSaveFilePanel(c)
}

// NewShare creates and display the share pannel to share the given value.
func NewShare(v interface{}) {
	driver.NewShare(v)
}

// NewNotification creates and displays the notification described in the
// given configuration.
func NewNotification(c NotificationConfig) {
	driver.NewNotification(c)
}

// MenuBar returns the menu bar.
func MenuBar() Menu {
	m, _ := driver.MenuBar()
	return m
}

// NewStatusMenu creates and displays the status menu described in the given
// configuration.
func NewStatusMenu(c StatusMenuConfig) StatusMenu {
	s, _ := driver.NewStatusMenu(c)
	return s
}

// Dock returns the dock tile.
func Dock() DockTile {
	d, _ := driver.Dock()
	return d
}

// CallOnUIGoroutine calls a function on the UI goroutine.
// UI goroutine is the running application main thread.
func CallOnUIGoroutine(f func()) {
	driver.CallOnUIGoroutine(f)
}

// Log logs a message according to a format specifier.
// It is a helper function that calls Log() for all the loggers set in
// app.Loggers.
func Log(format string, v ...interface{}) {
	for _, l := range Loggers {
		l.Log(format, v...)
	}
}

// Debug logs a debug message according to a format specifier.
// It is a helper function that calls Debug() for all the loggers set in
// app.Loggers.
func Debug(format string, v ...interface{}) {
	for _, l := range Loggers {
		l.Debug(format, v...)
	}
}

// WhenDebug execute the given function when debug mode is enabled.
// It is a helper function that calls WhenDebug() for all the loggers set in
// app.Loggers.
func WhenDebug(f func()) {
	for _, l := range Loggers {
		l.WhenDebug(f)
	}
}
