package data

import (
	"testing"

	"smart-hotel-dispatch/internal/models"
)

func TestLoadSample(t *testing.T) {
	ds, err := LoadSample()
	if err != nil {
		t.Fatal(err)
	}
	if len(ds.Staff) != 4 {
		t.Errorf("staff: got %d, want 4", len(ds.Staff))
	}
	if len(ds.Tasks) != 3 {
		t.Errorf("tasks: got %d, want 3", len(ds.Tasks))
	}
}

func TestAddStaff(t *testing.T) {
	ds, _ := LoadSample()
	s := NewStore(ds)
	got, err := s.AddStaff(models.Staff{
		Name:   "Test",
		Role:   "Cleaner",
		Skills: []models.TaskType{models.RoomCleaning},
		Speed:  1.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "S5" {
		t.Errorf("ID: got %s, want S5", got.ID)
	}
	if len(s.Staff()) != 5 {
		t.Errorf("staff count: got %d, want 5", len(s.Staff()))
	}
}

func TestAddStaffValidation(t *testing.T) {
	s := NewStore(&Dataset{})
	if _, err := s.AddStaff(models.Staff{}); err == nil {
		t.Error("expected an error for an empty employee")
	}
}
