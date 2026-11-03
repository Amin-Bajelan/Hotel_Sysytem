package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/models"
)

func main() {
	a := app.New()
	w := a.NewWindow("Smart Hotel Staff Dispatch")

	demo := models.Staff{ID: "S1", Name: "Ali", Status: models.StatusAvailable}
	label := widget.NewLabel(fmt.Sprintf("Setup OK. Demo staff: %s (%s)", demo.Name, demo.Status))

	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("Smart Hotel Staff Dispatch", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		label,
	))
	w.Resize(fyne.NewSize(1000, 650))
	w.ShowAndRun()
}