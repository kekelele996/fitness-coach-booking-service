package worker

import (
	"context"
	"testing"

	"fitness/internal/model"
)

// TestTickRespectsRetryLimit 到达重试上限后不再被调度器重试。
func TestTickRespectsRetryLimit(t *testing.T) {
	sch, svc, repo := setup()
	repo.CreateMember(&model.Member{ID: "m1", PackageID: "p-m1"})
	repo.CreatePackage(&model.SessionPackage{ID: "p-m1", MemberID: "m1", TotalSessions: 10, RemainingSessions: 10})
	repo.CreateCoach(&model.Coach{ID: "c1", Skills: []string{"strength-cert"}, Available: true})
	repo.CreateSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"})
	b, _ := svc.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	repo.UpdateBooking(b.ID, func(bb *model.Booking) {
		bb.Status = model.StatusFailed
		bb.Attempts = 3
		bb.MaxAttempts = 3
	})
	_, retried := sch.Tick(context.Background())
	if retried != 0 {
		t.Fatalf("retried=%d, want 0 (already at retry limit)", retried)
	}
}
