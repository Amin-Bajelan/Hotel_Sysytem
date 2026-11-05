package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/ui"
)

func main() {
	ds, err := data.LoadSample()
	if err != nil {
		log.Fatalf("failed to load sample data: %v", err)
	}
	store := data.NewStore(ds)

	a := app.New()
	w := a.NewWindow("Smart Hotel Staff Dispatch")

	tabs := container.NewAppTabs(
		container.NewTabItem("Staff", ui.NewStaffTab(store, w)),
		container.NewTabItem("Tasks", ui.NewTasksTab(store, w)),
		container.NewTabItem("Dispatch", ui.NewDispatchTab(store)),
	)
	w.SetContent(tabs)
	w.Resize(fyne.NewSize(1000, 650))
	w.ShowAndRun()
}
