package ui

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/location"
	"smart-hotel-dispatch/internal/models"
	"smart-hotel-dispatch/internal/simulation"
)

const (
	worldW = 100.0
	worldH = 52.0
)

var (
	colMapBg     = color.NRGBA{R: 238, G: 240, B: 244, A: 255}
	colCorridor  = color.NRGBA{R: 208, G: 213, B: 222, A: 255}
	colRoom      = color.NRGBA{R: 252, G: 252, B: 252, A: 255}
	colBorder    = color.NRGBA{R: 120, G: 125, B: 135, A: 255}
	colText      = color.NRGBA{R: 30, G: 30, B: 35, A: 255}
	colWhite     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colAvailable = color.NRGBA{R: 46, G: 160, B: 67, A: 255}
	colBusy      = color.NRGBA{R: 30, G: 100, B: 220, A: 255}
	colPending   = color.NRGBA{R: 240, G: 150, B: 20, A: 255}
	colAssigned  = color.NRGBA{R: 215, G: 40, B: 40, A: 255}
	colPath      = color.NRGBA{R: 30, G: 100, B: 220, A: 200}
)

var floorNames = []string{"Floor 0 (rooms 101-107)", "Floor 1 (rooms 201-207)", "Floor 2 (rooms 301-307)"}

func facilityColor(name string) color.Color {
	switch name {
	case "Lobby":
		return color.NRGBA{R: 200, G: 225, B: 255, A: 255}
	case "Reception":
		return color.NRGBA{R: 255, G: 235, B: 200, A: 255}
	case "Restaurant":
		return color.NRGBA{R: 220, G: 245, B: 215, A: 255}
	case "Elevator":
		return color.NRGBA{R: 185, G: 185, B: 210, A: 255}
	}
	return color.NRGBA{R: 205, G: 195, B: 175, A: 255} // Staircase
}

func centeredText(s string, size float32, c color.Color, center fyne.Position) *canvas.Text {
	t := canvas.NewText(s, c)
	t.TextSize = size
	ms := t.MinSize()
	t.Move(fyne.NewPos(center.X-ms.Width/2, center.Y-ms.Height/2))
	t.Resize(ms)
	return t
}

// ---- the map widget ----

type mapView struct {
	widget.BaseWidget
	store     *data.Store
	engine    *simulation.Engine
	hotel     models.Hotel
	floor     int
	showPaths bool
}

func (m *mapView) CreateRenderer() fyne.WidgetRenderer {
	return &mapRenderer{m: m}
}

type mapRenderer struct {
	m       *mapView
	size    fyne.Size
	objects []fyne.CanvasObject
}

func (r *mapRenderer) Layout(size fyne.Size) {
	r.size = size
	r.rebuild()
}

func (r *mapRenderer) MinSize() fyne.Size { return fyne.NewSize(500, 280) }

func (r *mapRenderer) Refresh() {
	r.rebuild()
	canvas.Refresh(r.m)
}

func (r *mapRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *mapRenderer) Destroy() {}

// rebuild redraws the whole floor from the current data.
func (r *mapRenderer) rebuild() {
	m := r.m
	if r.size.Width <= 0 || r.size.Height <= 0 {
		return
	}
	scale := float32(math.Min(float64(r.size.Width)/worldW, float64(r.size.Height)/worldH))
	offX := (r.size.Width - worldW*scale) / 2
	offY := (r.size.Height - worldH*scale) / 2

	pos := func(x, y float64) fyne.Position {
		return fyne.NewPos(offX+float32(x)*scale, offY+float32(y)*scale)
	}
	textSize := float32(math.Max(8, math.Min(14, float64(scale)*1.2)))

	var objs []fyne.CanvasObject
	box := func(cx, cy, w, h float64, fill color.Color) {
		rect := canvas.NewRectangle(fill)
		rect.StrokeColor = colBorder
		rect.StrokeWidth = 1
		rect.Move(pos(cx-w/2, cy-h/2))
		rect.Resize(fyne.NewSize(float32(w)*scale, float32(h)*scale))
		objs = append(objs, rect)
	}
	label := func(s string, x, y float64) {
		objs = append(objs, centeredText(s, textSize, colText, pos(x, y)))
	}

	// building and corridor
	box(worldW/2, worldH/2, worldW, worldH, colMapBg)
	box(50, location.CorridorY, worldW, 10, colCorridor)

	// facilities and rooms of the selected floor
	for _, f := range m.hotel.Facilities {
		if f.Floor != m.floor {
			continue
		}
		box(f.X, f.Y, f.W, f.H, facilityColor(f.Name))
		label(f.Name, f.X, f.Y)
	}
	for _, rm := range m.hotel.Rooms {
		if rm.Floor != m.floor {
			continue
		}
		box(rm.X, rm.Y, rm.W, rm.H, colRoom)
		label(rm.Name, rm.X, rm.Y-rm.H/2+2.5)
	}

	staff := m.store.Staff()
	tasks := m.store.Tasks()

	// remaining walking paths (only the part that lies on this floor)
	if m.showPaths && m.engine != nil {
		for _, s :=range staff {
			route := m.engine.RemainingRoute(s.ID)
			for i := 1; i < len(route); i++ {
				p, q := route[i-1], route[i]
				if p.Floor != q.Floor || p.Floor != m.floor {
					continue
				}
				line := canvas.NewLine(colPath)
				line.StrokeWidth = 2.5
				line.Position1 = pos(p.X, p.Y)
				line.Position2 = pos(q.X, q.Y)
				objs = append(objs, line)
			}
		}
	}

	// task destinations
	for _, t := range tasks {
		if t.Status == models.StatusCompleted || t.Status == models.StatusCancelled {
			continue
		}
		rp := location.RoomPoint(t.RoomNumber)
		if rp.Floor != m.floor {
			continue
		}
		col := colPending
		if t.Status != models.StatusPending {
			col = colAssigned
		}
		d := 2.8 * scale
		sq := canvas.NewRectangle(col)
		sq.StrokeColor = colWhite
		sq.StrokeWidth = 1
		p := pos(rp.X, rp.Y)
		sq.Move(fyne.NewPos(p.X-d/2, p.Y-d/2))
		sq.Resize(fyne.NewSize(d, d))
		objs = append(objs, sq, centeredText(t.ID, textSize, colText, fyne.NewPos(p.X, p.Y-d/2-7)))
	}

	// employees
	for _, s := range staff {
		if s.Floor != m.floor {
			continue
		}
		col := colAvailable
		if s.Status == models.StatusBusy {
			col = colBusy
		}
		d := 3.2 * scale
		c := canvas.NewCircle(col)
		c.StrokeColor = colWhite
		c.StrokeWidth = 1.5
		p := pos(s.X, s.Y)
		c.Move(fyne.NewPos(p.X-d/2, p.Y-d/2))
		c.Resize(fyne.NewSize(d, d))
		objs = append(objs, c, centeredText(s.Name, textSize, colText, fyne.NewPos(p.X, p.Y+d/2+7)))
	}

	r.objects = objs
}

// ---- the tab ----

func NewMapTab(store *data.Store, engine *simulation.Engine) fyne.CanvasObject {
	mv := &mapView{store: store, engine: engine, hotel: location.DefaultHotel(), showPaths: true}
	mv.ExtendBaseWidget(mv)

	floorSel := widget.NewSelect(floorNames, func(sel string) {
		for i, n := range floorNames {
			if n == sel {
				mv.floor = i
			}
		}
		mv.Refresh()
	})
	floorSel.SetSelected(floorNames[0])

	paths := widget.NewCheck("Show paths", func(v bool) {
		mv.showPaths = v
		mv.Refresh()
	})
	paths.SetChecked(true)

	store.Subscribe(func() { fyne.Do(mv.Refresh) })

	legend := widget.NewLabel("Green: available employee | Blue: busy employee | Orange square: pending task | Red square: assigned task | Blue line: remaining path")
	legend.Wrapping = fyne.TextWrapWord

	top := container.NewHBox(widget.NewLabel("Floor:"), floorSel, paths)
	return container.NewBorder(top, legend, nil, nil, mv)
}