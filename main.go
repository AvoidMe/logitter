// TODO: parse flags from cli
// TODO: fix TODOs

// kill -s SIGUSR1 $(cat ~/.logitter.pid)
package main

import (
	"image/color"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	LOGITTER_WIDTH  = 650
	LOGITTER_HEIGHT = 30

	FONT_SIZE = 30
)

type Theme struct{}

func (m Theme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff} // Black
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x00} // White
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m Theme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m Theme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m Theme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return FONT_SIZE
	}
	return theme.DefaultTheme().Size(name)
}

var _ fyne.Theme = (*Theme)(nil)

type Entry struct {
	widget.Entry
	window fyne.Window
	db     *LogitterDB
}

func NewEntry(window fyne.Window, db *LogitterDB) *Entry {
	entry := &Entry{window: window, db: db}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (self *Entry) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeyReturn:
		OnSave(self.db, self, self.window)
	case fyne.KeyEscape:
		OnHide(self, self.window)
	default:
		self.Entry.TypedKey(event)
	}
}

func OnSave(db *LogitterDB, entry *Entry, window fyne.Window) {
	text := strings.Trim(entry.Text, " \n")
	if len(text) > 0 {
		db.InsertRecord(entry.Text)
	}
	OnHide(entry, window)
}

func OnHide(entry *Entry, window fyne.Window) {
	entry.SetText("")
	window.Hide()
}

func SignalListener(sigs chan os.Signal, w fyne.Window, entry *Entry) {
	for {
		<-sigs
		w.Show()
		w.Canvas().Focus(entry)
	}
}

func main() {
	// Checking if another instance of logitter is running
	// Exit immediately if it is
	if PIDExists() {
		log.Println("Loggitter already running, exiting...")
		return
	}
	WritePID()

	// Setup signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)

	// Init db
	db := NewDB()
	defer db.Close()

	// Start web-server
	go ServeFrontend(db)

	<-sigs // Waiting until waked up by signal

	// Create new window
	a := app.New()
	drv := a.Driver().(desktop.Driver)
	w := drv.CreateSplashWindow()
	w.SetTitle("Logitter")
	w.Resize(fyne.NewSize(LOGITTER_WIDTH, LOGITTER_HEIGHT))
	w.CenterOnScreen()

	// Set theme
	a.Settings().SetTheme(&Theme{})

	// Create main entry
	entry := NewEntry(w, db)
	w.SetContent(container.NewVBox(
		entry,
	))

	// Listen for next wake-up call
	go SignalListener(sigs, w, entry)

	// Setup events on enter/escape
	canvas := w.Canvas().(desktop.Canvas)
	canvas.SetOnKeyDown(entry.TypedKey)
	w.Canvas().Focus(entry)

	// Run program
	w.ShowAndRun()
}
