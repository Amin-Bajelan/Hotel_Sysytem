package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/dispatch"
	"smart-hotel-dispatch/internal/models"
)

// NewDispatchTab only previews what a strategy WOULD decide right now.
// Real assignments are made by the simulation engine.
func NewDispatchTab(store *data.Store) fyne.CanvasObject {
	strategies := map[string]dispatch.Strategy{
		"Weighted Scoring": dispatch.Weighted{W: dispatch.DefaultWeights()},
		"First Available":  dispatch.FirstAvailable{},
	}
	choice := widget.NewSelect([]string{"Weighted Scoring", "First Available"}, nil)
	choice.SetSelected("Weighted Scoring")

	output := widget.NewLabel("Press Preview to see what the strategy would decide for the pending tasks.")
	output.Wrapping = fyne.TextWrapWord

	preview := widget.NewButton("Preview", func() {
		tasks := store.Tasks()
		byID := map[string]models.Task{}
		for _, t := range tasks {
			byID[t.ID] = t
		}
		decisions := dispatch.Dispatch(strategies[choice.Selected], tasks, store.Staff())
		if len(decisions) == 0 {
			output.SetText("No pending tasks.")
			return
		}
		var sb strings.Builder
		for _, d := range decisions {
			t := byID[d.TaskID]
			fmt.Fprintf(&sb, "%s | %s | %s | room %d\n", t.ID, t.Priority, t.Type, t.RoomNumber)
			if d.Chosen == nil {
				fmt.Fprintf(&sb, "   UNASSIGNED: %s\n\n", d.Reason)
				continue
			}
			fmt.Fprintf(&sb, "   -> %s (%s)\n   %s\n\n", d.Chosen.Staff.Name, d.Chosen.Staff.ID, d.Reason)
		}
		output.SetText(sb.String())
	})

	top := container.NewHBox(widget.NewLabel("Strategy:"), choice, preview)
	return container.NewBorder(top, nil, nil, nil, container.NewVScroll(output))
}