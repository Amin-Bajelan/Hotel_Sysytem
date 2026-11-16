package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"smart-hotel-dispatch/internal/models"
)

//go:embed sample.json
var sampleJSON []byte

type Dataset struct {
	Staff []models.Staff `json:"staff"`
	Tasks []models.Task  `json:"tasks"`
}

func parse(b []byte) (*Dataset, error) {
	var ds Dataset
	if err := json.Unmarshal(b, &ds); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &ds, nil
}

// rebase shifts task times so the newest task was created "now".
func rebase(ds *Dataset) {
	var latest time.Time
	for _, t := range ds.Tasks {
		if t.CreatedAt.After(latest) {
			latest = t.CreatedAt
		}
	}
	if latest.IsZero() {
		return
	}
	shift := time.Since(latest)
	for i := range ds.Tasks {
		ds.Tasks[i].CreatedAt = ds.Tasks[i].CreatedAt.Add(shift)
	}
}

// LoadSample loads the data embedded in the binary.
func LoadSample() (*Dataset, error) {
	ds, err := parse(sampleJSON)
	if err != nil {
		return nil, err
	}
	rebase(ds)
	return ds, nil
}

// LoadFromFile loads a dataset from a JSON file on disk.
func LoadFromFile(path string) (*Dataset, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(b)
}