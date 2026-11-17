package location

import (
	"strconv"

	"smart-hotel-dispatch/internal/models"
)

const (
	StairsX = 93.0
	StairsY = 25.0
)

// DefaultHotel builds the 3-floor hotel: 21 rooms (101-107, 201-207, 301-307),
// an elevator and staircase on every floor, and the lobby, reception and
// restaurant on floor 0. Room positions come from RoomPoint, so the map and
// the dispatch algorithm always agree.
func DefaultHotel() models.Hotel {
	h := models.Hotel{Floors: 3}
	for f := 0; f < 3; f++ {
		for n := 1; n <= 7; n++ {
			room := (f+1)*100 + n
			p := RoomPoint(room)
			h.Rooms = append(h.Rooms, models.Location{
				Name: strconv.Itoa(room), Floor: f, X: p.X, Y: p.Y, W: 10, H: 16,
			})
		}
		h.Facilities = append(h.Facilities,
			models.Location{Name: "Elevator", Floor: f, X: ElevatorX, Y: ElevatorY, W: 7, H: 8},
			models.Location{Name: "Staircase", Floor: f, X: StairsX, Y: StairsY, W: 7, H: 8},
		)
	}
	h.Facilities = append(h.Facilities,
		models.Location{Name: "Reception", Floor: 0, X: 14, Y: 11, W: 24, H: 14},
		models.Location{Name: "Lobby", Floor: 0, X: 40, Y: 11, W: 24, H: 14},
		models.Location{Name: "Restaurant", Floor: 0, X: 72, Y: 11, W: 36, H: 14},
	)
	return h
}