package application

import "time"

type CreateShipmentDTO struct {
	ID              string
	ReferenceNumber string
	Origin          string
	Destination     string
	DriverName      string
	VehicleID       string
	Amount          float64
	DriverRevenue   float64
}

type AddShipmentEventDTO struct {
	EventID    string
	ShipmentID string
	Status     int32
	Note       string
	OccurredAt time.Time
}

type ShipmentDTO struct {
	ID            string
	ReferenceNum  string
	Origin        string
	Destination   string
	DriverName    string
	VehicleID     string
	Amount        float64
	DriverRevenue float64
	CurrentStatus int32
	Events        []*ShipmentEventDTO
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ShipmentEventDTO struct {
	ID         string
	ShipmentID string
	Status     int32
	Note       string
	OccurredAt time.Time
}
