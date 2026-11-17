package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/dispatch"
	"smart-hotel-dispatch/internal/simulation"
)

func setEnabled(b *widget.Button, on bool) {
	if on {
		b.Enable()
	} else {
		b.Disable()
	}
}

func NewControlBar(engine *simulation.Engine, win fyne.Window) fyne.CanvasObject {
	started := false // only touched on the GUI thread

	clock := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true, Monospace: true})
	startBtn := widget.NewButton("Start", nil)
	pauseBtn := widget.NewButton("Pause", nil)
	resumeBtn := widget.NewButton("Resume", nil)
	resetBtn := widget.NewButton("Reset", nil)

	update := func() {
		running := engine.Running()
		state := "Idle"
		switch {
		case running:
			state = fmt.Sprintf("Running %gx", engine.Speed())
		case started:
			state = "Paused"
		}
		clock.SetText(fmt.Sprintf("Sim time %s  |  %s", engine.SimNow().Format("15:04:05"), state))
		setEnabled(startBtn, !started)
		setEnabled(pauseBtn, running)
		setEnabled(resumeBtn, started && !running)
	}

	startBtn.OnTapped = func() {
		engine.Start()
		started = true
		update()
	}
	pauseBtn.OnTapped = func() {
		engine.Pause()
		update()
	}
	resumeBtn.OnTapped = func() {
		engine.Resume()
		update()
	}
	resetBtn.OnTapped = func() {
		ds, err := data.LoadSample()
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		engine.Reset(ds)
		started = false
		update()
	}

	speedOpts := make([]string, len(simulation.Speeds))
	for i, s := range simulation.Speeds {
		speedOpts[i] = fmt.Sprintf("%gx", s)
	}
	speed := widget.NewSelect(speedOpts, func(sel string) {
		for i, o := range speedOpts {
			if o == sel {
				engine.SetSpeed(simulation.Speeds[i])
			}
		}
		update()
	})
	speed.SetSelected(speedOpts[0])

	strategies := map[string]dispatch.Strategy{
		"Weighted Scoring": dispatch.Weighted{W: dispatch.DefaultWeights()},
		"First Available":  dispatch.FirstAvailable{},
	}
	strat := widget.NewSelect([]string{"Weighted Scoring", "First Available"}, func(sel string) {
		engine.SetStrategy(strategies[sel])
	})
	strat.SetSelected("Weighted Scoring")

	update()
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for range t.C {
			fyne.Do(update)
		}
	}()

	row := container.NewHBox(
		startBtn, pauseBtn, resumeBtn, resetBtn,
		widget.NewLabel("Speed:"), speed,
		widget.NewLabel("Strategy:"), strat,
	)
	return container.NewVBox(row, clock, widget.NewSeparator())
}

func NewLogTab(engine *simulation.Engine, store *data.Store) fyne.CanvasObject {
	out := widget.NewLabel("")
	out.Wrapping = fyne.TextWrapWord

	refresh := func() {
		ev := engine.Events()
		if len(ev) == 0 {
			out.SetText("No events yet. Press Start.")
			return
		}
		var sb strings.Builder
		for i := len(ev) - 1; i >= 0; i-- { // newest first
			fmt.Fprintf(&sb, "[%s] %s\n", ev[i].Time.Format("15:04:05"), ev[i].Text)
		}
		out.SetText(sb.String())
	}
	refresh()
	store.Subscribe(func() { fyne.Do(refresh) })
	return container.NewVScroll(out)
}