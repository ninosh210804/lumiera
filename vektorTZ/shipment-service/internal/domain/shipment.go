package domain

import "time"

type Shipment struct {
	ID            string           `json:"id"`
	ReferenceNum  string           `json:"reference_number"`
	Origin        string           `json:"origin"`
	Destination   string           `json:"destination"`
	DriverName    string           `json:"driver_name"`
	VehicleID     string           `json:"vehicle_id"`
	Amount        float64          `json:"amount"`
	DriverRevenue float64          `json:"driver_revenue"`
	CurrentStatus Status           `json:"current_status"`
	Events        []*ShipmentEvent `json:"events"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func NewShipment(id, referenceNum, origin, destination, driverName, vehicleID string, amount, driverRevenue float64) *Shipment {
	now := time.Now()
	shipment := &Shipment{
		ID:            id,
		ReferenceNum:  referenceNum,
		Origin:        origin,
		Destination:   destination,
		DriverName:    driverName,
		VehicleID:     vehicleID,
		Amount:        amount,
		DriverRevenue: driverRevenue,
		CurrentStatus: StatusPending,
		Events:        make([]*ShipmentEvent, 0),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	initialEvent := NewShipmentEvent(id+"-event-0", id, StatusPending, "Shipment created", now)
	shipment.Events = append(shipment.Events, initialEvent)

	return shipment
}

func (s *Shipment) AddEvent(event *ShipmentEvent) error {
	if s.CurrentStatus.IsTerminal() {
		return ErrShipmentTerminal
	}

	if event.Status == s.CurrentStatus {
		return ErrDuplicateStatus
	}

	if !s.CurrentStatus.CanTransitionTo(event.Status) {
		return ErrInvalidTransition
	}

	s.Events = append(s.Events, event)
	s.CurrentStatus = event.Status
	s.UpdatedAt = event.OccurredAt

	return nil
}

func (s *Shipment) GetLatestEvent() *ShipmentEvent {
	if len(s.Events) == 0 {
		return nil
	}
	return s.Events[len(s.Events)-1]
}
