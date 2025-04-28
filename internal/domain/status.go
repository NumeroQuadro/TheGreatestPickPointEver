package domain

import (
	"encoding/json"
	"fmt"
)

type Status string

const (
	Confirmed Status = "confirmed"
	Completed Status = "completed"
	Refunded  Status = "refunded"
)

const (
	ConfirmedString string = "confirmed"
	CompletedString string = "completed"
	RefundedString  string = "refunded"
)

func NewStatusFromString(s string) (Status, error) {
	var state Status
	switch s {
	case ConfirmedString:
		state = Confirmed
	case CompletedString:
		state = Completed
	case RefundedString:
		state = Refunded
	default:
		return "", fmt.Errorf("unknown status string: %s", s)
	}

	return state, nil
}

func (s *Status) String() string {
	return string(*s)
}

func IsStatusValid(statusString string) bool {
	switch statusString {
	case ConfirmedString:
		return true
	case CompletedString:
		return true
	case RefundedString:
		return true
	default:
		return false
	}
}

func GetStatusTypeFromString(s string) (Status, error) {
	var state Status
	switch s {
	case ConfirmedString:
		state = Confirmed
	case CompletedString:
		state = Completed
	case RefundedString:
		state = Refunded
	default:
		return "", fmt.Errorf("unknown status string: %s", s)
	}

	return state, nil
}

func GetStringFromStatus(s Status) string {
	switch s {
	case Confirmed:
		return ConfirmedString
	case Completed:
		return CompletedString
	case Refunded:
		return RefundedString
	default:
		return ""
	}
}

func (s *Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*s))
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status, err := NewStatusFromString(str)
	if err != nil {
		return err
	}

	*s = status

	return nil
}
