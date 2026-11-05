package ui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/container"
	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/models"
)

var taskHeaders = []string{"ID", "Room", "Type", "Priority", "Status", "Created", "Assigned To", "Est. (min)"}

func taskCell(t models.Task, col int) string {
	switch col {
	case 0:
		return t.ID
	case 1:
		return strconv.Itoa(t.RoomNumber)
	case 2:
		return string(t.Type)
	case 3:
		return string(t.Priority)
	case 4:
		return string(t.Status)
	case 5:
		return t.CreatedAt.Format("15:04")
	case 6:
		if t.AssignedStaffID == "" {
			return "-"
		}
		return t.AssignedStaffID
	case 7:
		return strconv.Itoa(t.EstimatedDuration)
	}
	return ""
}

func NewTasksTab(store *data.Store, win fyne.Window) fyne.CanvasObject {
	table := widget.NewTable(
		func() (int, int) { return len(store.Tasks()) + 1, len(taskHeaders) },
		func() fyne.CanvasObject { return widget.NewLabel("placeholder") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText(taskHeaders[id.Col])
				return
			}
			l.TextStyle = fyne.TextStyle{}
			tasks := store.Tasks()
			if id.Row-1 >= len(tasks) {
				l.SetText("")
				return
			}
			l.SetText(taskCell(tasks[id.Row-1], id.Col))
		},
	)
	widths := []float32{50, 60, 140, 80, 100, 80, 110, 90}
	for i, w := range widths {
		table.SetColumnWidth(i, w)
	}
	addBtn := widget.NewButton("Add Task", func() {
		showAddTaskDialog(store, win, table.Refresh)
	})
	return container.NewBorder(container.NewHBox(addBtn), nil, nil, nil, table)
}