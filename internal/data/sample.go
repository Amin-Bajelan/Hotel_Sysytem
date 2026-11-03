package models

type Location struct {
	Name  string  `json:"name"`
	Floor int     `json:"floor"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type Hotel struct {
	Floors     int        `json:"floors"`
	Rooms      []Location `json:"rooms"`
	Facilities []Location `json:"facilities"`
}