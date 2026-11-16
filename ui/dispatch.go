package ui

import (
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/dispatch"
	"smart-hotel-dispatch/internal/models"
)

func NewDispatchTab(store *data.Store) fyne.CanvasObject {
	strategies := map[string]dispatch.Strategy{
		"Weighted Scoring": dispatch.Weighted{W: dispatch.DefaultWeights()},
		"First Available":  dispatch.FirstAvailable{},
	}
	choice := widget.NewSelect([]string{"Weighted Scoring", "First Available"}, nil)
	choice.SetSelected("Weighted Scoring")

	output := widget.NewLabel("Press Preview to see assignment decisions.")
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

	apply := widget.NewButton("Apply Assignments", func() {
		decisions := dispatch.Dispatch(strategies[choice.Selected], store.Tasks(), store.Staff())
		now := time.Now()
		applied, left := 0, 0
		for _, d := range decisions {
			if d.Chosen == nil {
				left++
				continue
			}
			if err := store.AssignTask(d.TaskID, d.Chosen.Staff.ID, d.Chosen.Distance, now); err != nil {
				left++
				continue
			}
			applied++
		}
		output.SetText(fmt.Sprintf("Applied %d assignment(s); %d task(s) left pending.", applied, left))
	})

	complete := widget.NewButton("Complete All Assigned", func() {
		now := time.Now()
		n := 0
		for _, t := range store.Tasks() {
			if t.Status == models.StatusAssigned || t.Status == models.StatusInProgress {
				if store.CompleteTask(t.ID, now) == nil {
					n++
				}
			}
		}
		output.SetText(fmt.Sprintf("Completed %d task(s).", n))
	})

	reset := widget.NewButton("Reset Data", func() {
		ds, err := data.LoadSample()
		if err != nil {
			output.SetText("Reset failed: " + err.Error())
			return
		}
		store.Reset(ds)
		output.SetText("Data reset to sample.")
	})

	top := container.NewVBox(
		container.NewHBox(widget.NewLabel("Strategy:"), choice),
		container.NewHBox(preview, apply, complete, reset),
	)
	return container.NewBorder(top, nil, nil, nil, container.NewVScroll(output))
}