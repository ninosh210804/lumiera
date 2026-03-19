package domain

import (
	"errors"

	"github.com/google/uuid"
)

func GenerateID() string {
	return uuid.New().String()
}

var (
	ErrInvalidTransition     = errors.New("invalid status transition")
	ErrShipmentTerminal      = errors.New("shipment has reached terminal status")
	ErrDuplicateStatus       = errors.New("cannot add event with same status as current")
	ErrShipmentNotFound      = errors.New("shipment not found")
	ErrReferenceNumberExists = errors.New("reference number already exists")
)
