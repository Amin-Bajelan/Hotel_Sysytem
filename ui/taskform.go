package ui

import (
	"errors"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/models"
)

func showAddTaskDialog(store *data.Store, win fyne.Window, onDone func()) {
	room := widget.NewEntry()
	room.SetPlaceHolder("e.g. 101")

	types := make([]string, len(models.AllTaskTypes))
	for i, t := range models.AllTaskTypes {
		types[i] = string(t)
	}
	typeSel := widget.NewSelect(types, nil)
	typeSel.SetSelected(types[0])

	prio := widget.NewSelect([]string{"Low", "Normal", "High", "Urgent"}, nil)
	prio.SetSelected("Normal")

	dur := widget.NewEntry()
	dur.SetText("15")

	items := []*widget.FormItem{
		widget.NewFormItem("Room", room),
		widget.NewFormItem("Type", typeSel),
		widget.NewFormItem("Priority", prio),
		widget.NewFormItem("Duration (min)", dur),
	}

	dialog.ShowForm("Add Task", "Add", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		r, err1 := strconv.Atoi(room.Text)
		m, err2 := strconv.Atoi(dur.Text)
		if err1 != nil || err2 != nil {
			dialog.ShowError(errors.New("room and duration must be whole numbers"), win)
			return
		}
		_, err := store.AddTask(r, models.TaskType(typeSel.Selected), models.Priority(prio.Selected), m)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		onDone()
	}, win)
}