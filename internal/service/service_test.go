package service

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"fitness/internal/config"
	"fitness/internal/model"
	"fitness/internal/repository"
	"fitness/internal/store"
)

func newService() *Service {
	return New(repository.New(store.New()), config.Load())
}

func seedCoach(s *Service, id string, skills []string, available bool) {
	s.repo.CreateCoach(&model.Coach{ID: id, Name: id, Skills: skills, Available: available})
}

func seedSchedule(s *Service, coachID, date, slot string) {
	s.repo.CreateSchedule(&model.Schedule{CoachID: coachID, Date: date, TimeSlot: slot})
}

func seedMember(s *Service, id string, sessions int) *model.Member {
	m := &model.Member{ID: id, Name: id, PackageID: "p-" + id}
	s.repo.CreateMember(m)
	p := &model.SessionPackage{ID: "p-" + id, MemberID: id, TotalSessions: sessions, RemainingSessions: sessions, Status: model.PackageActive}
	s.repo.CreatePackage(p)
	m.PackageID = p.ID
	return m
}

func TestCreateAndFind(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	b, err := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 2, model.PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != model.StatusPending {
		t.Fatalf("status=%q want pending", b.Status)
	}
	pkg, _ := s.repo.FindPackage("p-m1")
	if pkg.RemainingSessions != 8 {
		t.Fatalf("RemainingSessions=%d want 8", pkg.RemainingSessions)
	}
}

func TestFindMissingWrapsNotFound(t *testing.T) {
	s := newService()
	_, err := s.FindBooking("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound)=false, err=%v", err)
	}
}

func TestDispatchRejectsWrongSkill(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"yoga-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	_, err := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	if !errors.Is(err, ErrNoCoach) {
		t.Fatalf("err=%v want ErrNoCoach", err)
	}
}

func TestRejectSlotTaken(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedMember(s, "m2", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	_, err := s.CreateBooking("m2", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	if !errors.Is(err, ErrSlotTaken) {
		t.Fatalf("err=%v want ErrSlotTaken", err)
	}
}

func TestRejectInsufficientSessions(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 2)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	_, err := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 3, model.PriorityHigh)
	if !errors.Is(err, ErrInsufficientSessions) {
		t.Fatalf("err=%v want ErrInsufficientSessions", err)
	}
}

func TestLifecycle(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	b, _ := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 2, model.PriorityHigh)
	if _, err := s.ConfirmBooking(b.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CompleteBooking(b.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := s.FindBooking(b.ID)
	if got.Status != model.StatusCompleted {
		t.Fatalf("status=%q want completed", got.Status)
	}
}

func TestCancelRefundsSessions(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	b, _ := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 2, model.PriorityHigh)
	s.ConfirmBooking(b.ID)
	if _, err := s.CancelBooking(b.ID); err != nil {
		t.Fatal(err)
	}
	pkg, _ := s.repo.FindPackage("p-m1")
	if pkg.RemainingSessions != 10 {
		t.Fatalf("RemainingSessions=%d want 10 (refunded)", pkg.RemainingSessions)
	}
}

func TestListBookingsByStatus(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	seedSchedule(s, "c1", "2026-08-18", "11:00")
	b1, _ := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityLow)
	b2, _ := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "11:00", 1, model.PriorityUrgent)
	status := model.StatusPending
	got, _ := s.ListBookings(&status)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if got[0].ID != b2.ID || got[1].ID != b1.ID {
		t.Fatalf("order wrong: %v", got)
	}
}

func TestStats(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	b1, _ := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	s.ConfirmBooking(b1.ID)
	stats := s.Stats()
	if stats[model.StatusPending] != 0 {
		t.Fatalf("pending=%d want 0", stats[model.StatusPending])
	}
	if stats[model.StatusConfirmed] != 1 {
		t.Fatalf("confirmed=%d want 1", stats[model.StatusConfirmed])
	}
}

func TestActiveCount(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 10)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	seedSchedule(s, "c1", "2026-08-18", "10:00")
	b1, _ := s.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	s.ConfirmBooking(b1.ID)
	if got := s.ActiveCount(); got != 1 {
		t.Fatalf("ActiveCount=%d want 1", got)
	}
}

func TestConcurrentSessionDeduction(t *testing.T) {
	s := newService()
	seedMember(s, "m1", 100)
	seedCoach(s, "c1", []string{"strength-cert"}, true)
	for i := 0; i < 50; i++ {
		seedSchedule(s, "c1", "2026-08-18", fmt.Sprintf("%02d:00", i))
	}
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = s.CreateBooking("m1", "c1", "strength", "2026-08-18", fmt.Sprintf("%02d:00", i), 1, model.PriorityHigh)
		}(i)
	}
	wg.Wait()
	pkg, _ := s.repo.FindPackage("p-m1")
	if pkg.RemainingSessions != 50 {
		t.Fatalf("RemainingSessions=%d want 50", pkg.RemainingSessions)
	}
}

func TestCustomSkillRoutes(t *testing.T) {
	t.Setenv("FITNESS_SKILL_ROUTES", "strength:special")
	st := store.New()
	repo := repository.New(st)
	svc := New(repo, config.Load())
	m := &model.Member{ID: "m1", Name: "m1", PackageID: "p-m1"}
	repo.CreateMember(m)
	p := &model.SessionPackage{ID: "p-m1", MemberID: "m1", TotalSessions: 10, RemainingSessions: 10, Status: model.PackageActive}
	repo.CreatePackage(p)
	repo.CreateCoach(&model.Coach{ID: "c1", Name: "c1", Skills: []string{"special"}, Available: true})
	repo.CreateSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"})
	b, err := svc.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != model.StatusPending {
		t.Fatalf("status=%q", b.Status)
	}
}

func TestSkillRouteFallbackDispatch(t *testing.T) {
	t.Setenv("FITNESS_SKILL_ROUTES", "invalid")
	st := store.New()
	repo := repository.New(st)
	svc := New(repo, config.Load())
	repo.CreateMember(&model.Member{ID: "m1", PackageID: "p-m1"})
	repo.CreatePackage(&model.SessionPackage{ID: "p-m1", MemberID: "m1", TotalSessions: 10, RemainingSessions: 10, Status: model.PackageActive})
	repo.CreateCoach(&model.Coach{ID: "c1", Skills: []string{"strength-cert"}, Available: true})
	repo.CreateSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"})
	b, err := svc.CreateBooking("m1", "c1", "strength", "2026-08-18", "10:00", 1, model.PriorityHigh)
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != model.StatusPending {
		t.Fatalf("status=%q", b.Status)
	}
}
