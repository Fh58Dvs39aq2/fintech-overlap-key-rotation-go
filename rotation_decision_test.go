package main

import "testing"

func TestDecide(t *testing.T) {
	base := RotationRequest{Event: PaymentEvent{ID: "payment-42", Kind: "settlement_completed", Risk: "low", Approved: true}, ActiveKeyID: "active-1", TemporaryKeyID: "temporary-1", GraceHours: 24}
	tests := []struct {
		name    string
		change  func(*RotationRequest)
		allowed bool
	}{
		{"approved settlement", func(*RotationRequest) {}, true},
		{"high risk", func(r *RotationRequest) { r.Event.Risk = "high" }, false},
		{"unapproved", func(r *RotationRequest) { r.Event.Approved = false }, false},
		{"wrong event", func(r *RotationRequest) { r.Event.Kind = "payment_pending" }, false},
		{"no overlap", func(r *RotationRequest) { r.GraceHours = 0 }, false},
		{"missing active identity", func(r *RotationRequest) { r.ActiveKeyID = "" }, false},
		{"missing temporary identity", func(r *RotationRequest) { r.TemporaryKeyID = "" }, false},
		{"active key selected", func(r *RotationRequest) { r.TemporaryKeyID = r.ActiveKeyID }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := base
			tt.change(&r)
			if got := decide(r) == nil; got != tt.allowed {
				t.Fatalf("allowed = %v, want %v", got, tt.allowed)
			}
		})
	}
}
