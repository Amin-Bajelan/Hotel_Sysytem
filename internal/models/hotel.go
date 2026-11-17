package models

type Location struct {
	Name  string  `json:"name"`
	Floor int     `json:"floor"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
}

type Hotel struct {
	Floors     int        `json:"floors"`
	Rooms      []Location `json:"rooms"`
	Facilities []Location `json:"facilities"`
}