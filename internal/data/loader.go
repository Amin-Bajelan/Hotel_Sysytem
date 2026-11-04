package data

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"

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

// LoadSample loads the data embedded in the binary.
func LoadSample() (*Dataset, error) {
	return parse(sampleJSON)
}

// LoadFromFile loads a dataset from a JSON file on disk.
func LoadFromFile(path string) (*Dataset, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(b)
}
