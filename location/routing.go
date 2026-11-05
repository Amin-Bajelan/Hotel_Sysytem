package location

import "smart-hotel-dispatch/internal/models"

// RoomPoint gives a simulated position for a room number.
// 101-107 = floor 0, 201-207 = floor 1, 301-307 = floor 2.
// On Day 5 this will be replaced by the real hotel map.
func RoomPoint(room int) Point {
	floor := room/100 - 1
	if floor < 0 {
		floor = 0
	}
	n := room % 100
	return Point{Floor: floor, X: 10 + float64(n-1)*12, Y: 40}
}

func StaffPoint(s models.Staff) Point {
	return Point{Floor: s.Floor, X: s.X, Y: s.Y}
}

// TravelTime returns seconds. Speed must be > 0 (Eligible guarantees it).
func TravelTime(distance, speed float64) float64 {
	return distance / speed
}