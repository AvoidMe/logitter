// TODO: show text in web-UI
// TODO: currently backspace incorrectly removes unicode (non-ascii) runes
// TODO: fix panics
// TODO: fix backspace speed
// TODO: font searcher

// kill -s SIGUSR1 $(cat ~/.logitter.pid)
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	ACTIVE_FPS = 60

	LOGITTER_WIDTH  = 650
	LOGITTER_HEIGHT = 70
)

var (
	hidden = false

	SCREEN_WIDTH  = 0
	SCREEN_HEIGHT = 0
)

func hideWindow() {
	if !rl.IsWindowState(rl.FlagWindowHidden) {
		rl.SetWindowState(rl.FlagWindowHidden)
		hidden = true
	}
}

func showWindow() {
	rl.SetWindowPosition(
		SCREEN_WIDTH/2-LOGITTER_WIDTH/2,
		SCREEN_HEIGHT/2-LOGITTER_HEIGHT*2,
	)
	rl.ClearWindowState(rl.FlagWindowHidden)
	hidden = false
}

func ServeFrontend() {
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
	go ServeFrontend()

	// Setup window
	rl.InitWindow(650, 70, "")
	rl.SetTargetFPS(ACTIVE_FPS)
	rl.SetWindowState(rl.FlagWindowUndecorated)
	rl.SetExitKey(0) // Disable exit on escape by default

	monitor := rl.GetCurrentMonitor()
	SCREEN_HEIGHT = rl.GetMonitorHeight(monitor)
	SCREEN_WIDTH = rl.GetMonitorWidth(monitor)

	// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
	font := rl.LoadFontEx(
		"",
		32,
		nil,
		4096,
	)

	var lastBackspace time.Time
	var text string

	// Hack: hide window and show it at the same time, to have window position centered
	//			 without noticeable delay
	hideWindow()
	showWindow()

	for !rl.WindowShouldClose() {
		if hidden {
			<-sigs // Waiting until waked up by signal
			showWindow()
		}
		// Write text in db + hide window on enter
		if rl.IsKeyPressed(rl.KeyEnter) {
			hideWindow()
			db.InsertRecord(text)
			text = ""
		}
		// Just hide window on escape and clear text
		if rl.IsKeyPressed(rl.KeyEscape) {
			hideWindow()
			text = ""
		}
		// Get next char from queue
		char := rl.GetCharPressed()
		if char != 0 {
			text += string(rune(char))
		}
		// Remove char if backspace is pressed
		if rl.IsKeyDown(rl.KeyBackspace) {
			// Try to remove chars every 200 milliseconds
			t := time.Now()
			if lastBackspace.Add(200*time.Millisecond).Compare(t) <= 0 {
				lastBackspace = t
				if len(text) > 0 {
					text = text[:len(text)-1]
				}
			}
		}
		if rl.IsKeyUp(rl.KeyBackspace) {
			lastBackspace = time.Time{}
		}
		// Main Draw loop
		rl.BeginDrawing()
		{
			rl.ClearBackground(rl.RayWhite)
			rl.DrawTextEx(font, text, rl.Vector2{X: 15, Y: 20}, 32, 0, rl.Black)
		}
		rl.EndDrawing()
	}

	rl.CloseWindow()
}
