package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/reports"
)

type statCard struct {
	text *canvas.Text
	obj  fyne.CanvasObject
}

func newStatCard(title string) statCard {
	t := canvas.NewText("0", theme.Color(theme.ColorNameForeground))
	t.TextSize = 30
	t.TextStyle = fyne.TextStyle{Bold: true}
	t.Alignment = fyne.TextAlignCenter
	return statCard{text: t, obj: widget.NewCard(title, "", container.NewCenter(t))}
}

func (c statCard) set(v string) {
	c.text.Text = v
	c.text.Refresh()
}

func formatDuration(sec float64) string {
	if sec <= 0 {
		return "-"
	}
	total := int(sec + 0.5)
	if total < 60 {
		return fmt.Sprintf("%ds", total)
	}
	return fmt.Sprintf("%dm %02ds", total/60, total%60)
}

func NewDashboardTab(store *data.Store) fyne.CanvasObject {
	totalStaff := newStatCard("Total Employees")
	available := newStatCard("Available Employees")
	busy := newStatCard("Busy Employees")
	totalTasks := newStatCard("Total Tasks")
	pending := newStatCard("Pending Tasks")
	assigned := newStatCard("Assigned Tasks")
	completed := newStatCard("Completed Tasks")
	avgResp := newStatCard("Avg Response Time")
	distance := newStatCard("Total Distance")

	statusList := widget.NewList(
		func() int { return len(store.Staff()) },
		func() fyne.CanvasObject { return widget.NewLabel("placeholder") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			staff := store.Staff()
			if i >= len(staff) {
				return
			}
			s := staff[i]
			task := "-"
			if s.CurrentTaskID != "" {
				task = s.CurrentTaskID
			}
			o.(*widget.Label).SetText(fmt.Sprintf("%s (%s)  |  %s  |  task: %s", s.Name, s.Role, s.Status, task))
		},
	)

	refresh := func() {
		st := reports.Compute(store.Staff(), store.Tasks())
		totalStaff.set(fmt.Sprint(st.TotalStaff))
		available.set(fmt.Sprint(st.AvailableStaff))
		busy.set(fmt.Sprint(st.BusyStaff))
		totalTasks.set(fmt.Sprint(st.TotalTasks))
		pending.set(fmt.Sprint(st.Pending))
		assigned.set(fmt.Sprint(st.Assigned))
		completed.set(fmt.Sprint(st.Completed))
		avgResp.set(formatDuration(st.AvgResponseSec))
		distance.set(fmt.Sprintf("%.1f", st.TotalDistance))
		statusList.Refresh()
	}
	refresh()
	store.Subscribe(func() { fyne.Do(refresh) })

	grid := container.NewGridWithColumns(3,
		totalStaff.obj, available.obj, busy.obj,
		totalTasks.obj, pending.obj, assigned.obj,
		completed.obj, avgResp.obj, distance.obj,
	)
	header := widget.NewLabelWithStyle("Employee Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewBorder(container.NewVBox(grid, header), nil, nil, nil, statusList)
}