package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/dispatch"
	"smart-hotel-dispatch/internal/simulation"
	"smart-hotel-dispatch/ui"
)

func main() {
	ds, err := data.LoadSample()
	if err != nil {
		log.Fatalf("failed to load sample data: %v", err)
	}
	store := data.NewStore(ds)
	engine := simulation.NewEngine(store, dispatch.Weighted{W: dispatch.DefaultWeights()})

	a := app.New()
	w := a.NewWindow("Smart Hotel Staff Dispatch")

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", ui.NewDashboardTab(store)),
		container.NewTabItem("Map", ui.NewMapTab(store, engine)),
		container.NewTabItem("Staff", ui.NewStaffTab(store, w)),
		container.NewTabItem("Tasks", ui.NewTasksTab(store, w)),
		container.NewTabItem("Dispatch", ui.NewDispatchTab(store)),
		container.NewTabItem("Log", ui.NewLogTab(engine, store)),
	)
	w.SetContent(container.NewBorder(ui.NewControlBar(engine, w), nil, nil, nil, tabs))
	w.Resize(fyne.NewSize(1100, 700))
	w.ShowAndRun()
}