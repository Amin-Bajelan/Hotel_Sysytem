package models

import "time"

type TaskType string
type Priority string
type TaskStatus string

const (
	RoomCleaning    TaskType = "Room Cleaning"
	TowelRequest    TaskType = "Towel Request"
	FoodDelivery    TaskType = "Food Delivery"
	Maintenance     TaskType = "Maintenance"
	GuestAssistance TaskType = "Guest Assistance"
)

const (
	PriorityLow    Priority = "Low"
	PriorityNormal Priority = "Normal"
	PriorityHigh   Priority = "High"
	PriorityUrgent Priority = "Urgent"
)

const (
	StatusPending    TaskStatus = "Pending"
	StatusAssigned   TaskStatus = "Assigned"
	StatusInProgress TaskStatus = "In Progress"
	StatusCompleted  TaskStatus = "Completed"
	StatusCancelled  TaskStatus = "Cancelled"
)

type Task struct {
	ID                string     `json:"id"`
	RoomNumber        int        `json:"room_number"`
	Type              TaskType   `json:"type"`
	Priority          Priority   `json:"priority"`
	Status            TaskStatus `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	AssignedStaffID   string     `json:"assigned_staff_id"`
	EstimatedDuration int        `json:"estimated_duration_min"`
}
