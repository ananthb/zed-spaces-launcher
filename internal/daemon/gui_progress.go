package daemon

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// progressScreen runs an on-theme rocket-launch animation while a long-running
// flow (codespace boot, SSH connect, editor launch) is in progress.
//
// The animation has two phases:
//
//  1. launch: a one-shot lift-off sequence that plays once on first show.
//  2. cruise: a looping flight-through-stars sequence that plays until the
//     screen is replaced or stop() is called.
type progressScreen struct {
	rocket *widget.Label
	status *widget.Label
	canvas fyne.CanvasObject

	stopOnce sync.Once
	done     chan struct{}
}

func newProgressScreen(message string) *progressScreen {
	rocket := widget.NewLabelWithStyle(
		launchFrames[0],
		fyne.TextAlignLeading,
		fyne.TextStyle{Monospace: true},
	)
	status := widget.NewLabel(message)
	status.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(rocket),
		status,
		layout.NewSpacer(),
	)

	p := &progressScreen{
		rocket: rocket,
		status: status,
		canvas: container.NewPadded(content),
		done:   make(chan struct{}),
	}
	go p.animate()
	return p
}

func (p *progressScreen) setStatus(msg string) {
	p.status.SetText(msg)
}

// stop halts the animation goroutine. Safe to call multiple times.
func (p *progressScreen) stop() {
	p.stopOnce.Do(func() { close(p.done) })
}

func (p *progressScreen) animate() {
	const frameDelay = 220 * time.Millisecond
	ticker := time.NewTicker(frameDelay)
	defer ticker.Stop()

	for i := 1; i < len(launchFrames); i++ {
		select {
		case <-p.done:
			return
		case <-ticker.C:
		}
		frame := launchFrames[i]
		fyne.Do(func() { p.rocket.SetText(frame) })
	}

	for i := 0; ; i = (i + 1) % len(cruiseFrames) {
		select {
		case <-p.done:
			return
		case <-ticker.C:
		}
		frame := cruiseFrames[i]
		fyne.Do(func() { p.rocket.SetText(frame) })
	}
}

// launchFrames play once: rocket on the pad, ignites, lifts off, climbs out
// of view. Each frame is exactly 12 rows tall and 15 columns wide so the
// canvas does not jump between frames.
var launchFrames = []string{
	// 0: standing on the pad
	`.      *      .
   *
      .
            *
  .
       *      .


      /\
     /||\
    /====\
  /==[||]==\   `,
	// 1: ignition, flame appears under the rocket
	`.      *      .
   *
      .
            *
  .
       *      .

      /\
     /||\
    /====\
    \V||V/
  /==[||]==\   `,
	// 2: lift-off, rocket clears the pad, exhaust trail forms
	`.      *      .
   *
      .
            *
  .
       *      .
      /\
     /||\
    /====\
    \V||V/
      ::
  /========\   `,
	// 3: climbing
	`.      *      .
   *
      .
            *
  .
      /\
     /||\
    /====\
    \V||V/
      ::
      ::
  /========\   `,
	// 4
	`.      *      .
   *
      .
            *
      /\
     /||\
    /====\
    \V||V/
      ::
      ::
      ::
  /========\   `,
	// 5
	`.      *      .
   *
      .
      /\
     /||\
    /====\
    \V||V/
      ::
      ::
      ::
      ::
  /========\   `,
	// 6
	`.      *      .
   *
      /\
     /||\
    /====\
    \V||V/
      ::
      ::
      ::
      ::
      ::
  /========\   `,
	// 7: nearing the top of the canvas
	`.      *      .
      /\
     /||\
    /====\
    \V||V/
      ::
      ::
      ::
      ::
      ::
      ::
  /========\   `,
}

// cruiseFrames loop while the rocket is in space. The rocket sits at the top
// of the canvas, the exhaust nozzle flickers, and the starfield scrolls
// downward to suggest continued upward flight.
var cruiseFrames = []string{
	`      /\
     /||\
    /====\
    \V||V/
.
       *
   .
       .
*
   .       *
.        .
        .      `,
	`      /\
     /||\
    /====\
    \v||v/
   .       .
.
       *
   .
       .
*
   .       *
.        .     `,
	`      /\
     /||\
    /====\
    \V||V/
.        .
   .       .
.
       *
   .
       .
*
   .       *   `,
	`      /\
     /||\
    /====\
    \v||v/
   .       *
.        .
   .       .
.
       *
   .
       .
*              `,
}
