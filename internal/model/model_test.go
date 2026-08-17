package model

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct{ from, to BookingStatus; want bool }{
		{StatusPending, StatusConfirmed, true},
		{StatusPending, StatusCompleted, false},
		{StatusConfirmed, StatusCompleted, true},
		{StatusConfirmed, StatusCancelled, true},
		{StatusFailed, StatusRetrying, true},
		{StatusRetrying, StatusConfirmed, true},
		{StatusCompleted, StatusCancelled, false},
	}
	for _, c := range cases {
		if got := CanTransition(c.from, c.to); got != c.want {
			t.Fatalf("CanTransition(%s,%s)=%v want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestActiveStatusesIncludesRetrying(t *testing.T) {
	if !ActiveStatuses[StatusRetrying] {
		t.Fatal("ActiveStatuses must include retrying")
	}
	if ActiveStatuses[StatusCompleted] {
		t.Fatal("ActiveStatuses must not include completed")
	}
}

func TestHasSkill(t *testing.T) {
	c := &Coach{Skills: []string{"strength-cert", "yoga-cert"}}
	if !c.HasSkill("strength-cert") {
		t.Fatal("expected HasSkill(strength-cert) true")
	}
	if c.HasSkill("hiit-cert") {
		t.Fatal("expected HasSkill(hiit-cert) false")
	}
}
