package domain

import "time"

type ShipmentEvent struct {
	ID         string    `json:"id"`
	ShipmentID string    `json:"shipment_id"`
	Status     Status    `json:"status"`
	Note       string    `json:"note"`
	OccurredAt time.Time `json:"occurred_at"`
}

func NewShipmentEvent(id string, shipmentID string, status Status, note string, occurredAt time.Time) *ShipmentEvent {
	return &ShipmentEvent{
		ID:         id,
		ShipmentID: shipmentID,
		Status:     status,
		Note:       note,
		OccurredAt: occurredAt,
	}
}
