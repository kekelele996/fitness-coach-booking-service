package worker

import (
	"context"
	"testing"
	"time"

	"fitness/internal/config"
	"fitness/internal/model"
	"fitness/internal/repository"
	"fitness/internal/service"
	"fitness/internal/store"
)

func setup() (*Scheduler, *service.Service, *repository.Repository) {
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, config.Load())
	sch := New(repo, svc, 500*time.Millisecond)
	return sch, svc, repo
}

func TestTickExpiresPending(t *testing.T) {
	sch, svc, repo := setup()
	repo.CreateMember(&model.Member{ID: "m1", PackageID: "p-m1"})
	repo.CreatePackage(&model.SessionPackage{ID: "p-m1", MemberID: "m1", TotalSessions: 10, RemainingSessions: 10})
	repo.CreateCoach(&model.Coach{ID: "c1", Skills: []string{"strength-cert"}, Available: true})
	repo.CreateSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"})
	b, _ := svc.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	// 把 ScheduledAt 拨到过去
	repo.UpdateBooking(b.ID, func(bb *model.Booking) {
		bb.ScheduledAt = time.Now().Add(-1 * time.Hour)
	})
	expired, _ := sch.Tick(context.Background())
	if expired != 1 {
		t.Fatalf("expired=%d want 1", expired)
	}
	got, _ := svc.FindBooking(b.ID)
	if got.Status != model.StatusExpired {
		t.Fatalf("status=%q want expired", got.Status)
	}
}

func TestTickRetriesFailed(t *testing.T) {
	sch, svc, repo := setup()
	repo.CreateMember(&model.Member{ID: "m1", PackageID: "p-m1"})
	repo.CreatePackage(&model.SessionPackage{ID: "p-m1", MemberID: "m1", TotalSessions: 10, RemainingSessions: 10})
	repo.CreateCoach(&model.Coach{ID: "c1", Skills: []string{"strength-cert"}, Available: true})
	repo.CreateSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"})
	b, _ := svc.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	repo.UpdateBooking(b.ID, func(bb *model.Booking) {
		bb.Status = model.StatusFailed
		bb.Attempts = 1
	})
	_, retried := sch.Tick(context.Background())
	if retried != 1 {
		t.Fatalf("retried=%d want 1", retried)
	}
	got, _ := svc.FindBooking(b.ID)
	if got.Status != model.StatusRetrying {
		t.Fatalf("status=%q want retrying", got.Status)
	}
}

func TestTickStopsWhenCancelled(t *testing.T) {
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, config.Load())
	sch := New(repo, svc, 500*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	expired, retried := sch.Tick(ctx)
	if expired != 0 || retried != 0 {
		t.Fatalf("expired=%d retried=%d want 0/0", expired, retried)
	}
}
