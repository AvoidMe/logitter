// TODO: write text to db
// TODO: show text in web-UI
// TODO: currently backspace incorrectly removes unicode (non-ascii) runes
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func hideWindow() {
	if !rl.IsWindowState(rl.FlagWindowHidden) {
		rl.SetWindowState(rl.FlagWindowHidden)
	}
}

func showWindow() {
	rl.ClearWindowState(rl.FlagWindowHidden)
}

func signalHandler(sigs chan os.Signal) {
	for {
		<-sigs
		showWindow()
	}
}

func main() {
	// Setup signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go signalHandler(sigs)

	// Setup window
	rl.InitWindow(650, 70, "")
	rl.SetTargetFPS(150)
	rl.SetWindowState(rl.FlagWindowUndecorated)

	// NOTE: Textures/Fonts MUST be loaded after Window initialization (OpenGL context is required)
	font := rl.LoadFontEx(
		"",
		32,
		nil,
		4096,
	)

	var lastBackspace time.Time
	var text string

	for !rl.WindowShouldClose() {
		// Write text in db + hide window on enter
		if rl.IsKeyPressed(rl.KeyEnter) {
			hideWindow()
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
