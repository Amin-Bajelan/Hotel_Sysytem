package ui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/models"
)

var staffHeaders = []string{"ID", "Name", "Role", "Skills", "Floor", "X", "Y", "Speed", "Status", "Task"}

func staffCell(s models.Staff, col int) string {
	switch col {
	case 0:
		return s.ID
	case 1:
		return s.Name
	case 2:
		return s.Role
	case 3:
		names := make([]string, len(s.Skills))
		for i, sk := range s.Skills {
			names[i] = string(sk)
		}
		return strings.Join(names, ", ")
	case 4:
		return strconv.Itoa(s.Floor)
	case 5:
		return fmt.Sprintf("%.0f", s.X)
	case 6:
		return fmt.Sprintf("%.0f", s.Y)
	case 7:
		return fmt.Sprintf("%.1f", s.Speed)
	case 8:
		return string(s.Status)
	case 9:
		if s.CurrentTaskID == "" {
			return "-"
		}
		return s.CurrentTaskID
	}
	return ""
}

func NewStaffTab(store *data.Store, win fyne.Window) fyne.CanvasObject {
	table := widget.NewTable(
		func() (int, int) { return len(store.Staff()) + 1, len(staffHeaders) },
		func() fyne.CanvasObject { return widget.NewLabel("placeholder") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			l := o.(*widget.Label)
			if id.Row == 0 {
				l.TextStyle = fyne.TextStyle{Bold: true}
				l.SetText(staffHeaders[id.Col])
				return
			}
			l.TextStyle = fyne.TextStyle{}
			staff := store.Staff()
			if id.Row-1 >= len(staff) {
				l.SetText("")
				return
			}
			l.SetText(staffCell(staff[id.Row-1], id.Col))
		},
	)
	widths := []float32{50, 110, 120, 280, 60, 60, 60, 60, 100, 70}
	for i, w := range widths {
		table.SetColumnWidth(i, w)
	}

	addBtn := widget.NewButton("Add Staff", func() {
		showAddStaffDialog(store, win, table.Refresh)
	})
	return container.NewBorder(container.NewHBox(addBtn), nil, nil, nil, table)
}

func showAddStaffDialog(store *data.Store, win fyne.Window, onDone func()) {
	name := widget.NewEntry()
	role := widget.NewEntry()

	options := make([]string, len(models.AllTaskTypes))
	for i, t := range models.AllTaskTypes {
		options[i] = string(t)
	}
	skills := widget.NewCheckGroup(options, nil)

	floor := widget.NewSelect([]string{"0", "1", "2"}, nil)
	floor.SetSelected("0")

	x := widget.NewEntry()
	x.SetText("0")
	y := widget.NewEntry()
	y.SetText("0")
	speed := widget.NewEntry()
	speed.SetText("1.5")

	items := []*widget.FormItem{
		widget.NewFormItem("Name", name),
		widget.NewFormItem("Role", role),
		widget.NewFormItem("Skills", skills),
		widget.NewFormItem("Floor", floor),
		widget.NewFormItem("X", x),
		widget.NewFormItem("Y", y),
		widget.NewFormItem("Speed", speed),
	}

	dialog.ShowForm("Add Staff", "Add", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		xv, err1 := strconv.ParseFloat(x.Text, 64)
		yv, err2 := strconv.ParseFloat(y.Text, 64)
		sp, err3 := strconv.ParseFloat(speed.Text, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			dialog.ShowError(errors.New("X, Y and Speed must be numbers"), win)
			return
		}
		fl, _ := strconv.Atoi(floor.Selected)
		sk := make([]models.TaskType, len(skills.Selected))
		for i, s := range skills.Selected {
			sk[i] = models.TaskType(s)
		}
		_, err := store.AddStaff(models.Staff{
			Name: name.Text, Role: role.Text, Skills: sk,
			Floor: fl, X: xv, Y: yv, Speed: sp,
		})
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		onDone()
	}, win)
}