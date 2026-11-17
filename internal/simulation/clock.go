package simulation

// BaseTimeScale is how many simulated seconds pass per real second at 1x speed.
const BaseTimeScale = 6.0

// Speeds are the multipliers offered in the GUI.
var Speeds = []float64{1, 2, 5, 10}

// maxRealStep caps the real time of one tick, so a stalled window
// does not make the simulation jump forward.
const maxRealStep = 0.5