package domain

type Status int32

const (
	StatusUnspecified Status = 0
	StatusPending     Status = 1
	StatusPickedUp    Status = 2
	StatusInTransit   Status = 3
	StatusDelivered   Status = 4
	StatusCancelled   Status = 5
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusPickedUp:
		return "PICKED_UP"
	case StatusInTransit:
		return "IN_TRANSIT"
	case StatusDelivered:
		return "DELIVERED"
	case StatusCancelled:
		return "CANCELLED"
	default:
		return "UNSPECIFIED"
	}
}

func (s Status) IsTerminal() bool {
	return s == StatusDelivered || s == StatusCancelled
}

func (s Status) CanTransitionTo(next Status) bool {
	if s.IsTerminal() {
		return false
	}

	switch s {
	case StatusPending:
		return next == StatusPickedUp || next == StatusCancelled
	case StatusPickedUp:
		return next == StatusInTransit || next == StatusCancelled
	case StatusInTransit:
		return next == StatusDelivered || next == StatusCancelled
	default:
		return false
	}
}
