package main

import "fmt"

type PaymentEvent struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Risk     string `json:"risk"`
	Approved bool   `json:"approved"`
}

type RotationRequest struct {
	Event          PaymentEvent `json:"event"`
	ActiveKeyID    string       `json:"active_key_id"`
	TemporaryKeyID string       `json:"temporary_key_id"`
	GraceHours     int          `json:"grace_hours"`
}

func decide(r RotationRequest) error {
	if r.Event.ID == "" || r.ActiveKeyID == "" || r.TemporaryKeyID == "" {
		return fmt.Errorf("event id, active key id, and temporary key id are required")
	}
	if r.TemporaryKeyID == r.ActiveKeyID {
		return fmt.Errorf("active key must stay untouched")
	}
	if r.Event.Kind != "settlement_completed" || !r.Event.Approved || r.Event.Risk != "low" {
		return fmt.Errorf("payment event requires compliance review")
	}
	if r.GraceHours < 1 || r.GraceHours > 72 {
		return fmt.Errorf("grace_hours must be between 1 and 72")
	}
	return nil
}
