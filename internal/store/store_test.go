package store

import (
	"testing"

	"fitness/internal/model"
)

func TestPutGetBooking(t *testing.T) {
	s := New()
	b := &model.Booking{ID: "b1", MemberID: "m1", Status: model.StatusPending}
	if err := s.PutBooking(b); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetBooking("b1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "b1" || got.Status != model.StatusPending {
		t.Fatalf("got=%+v", got)
	}
	if _, err := s.GetBooking("missing"); err != ErrNotFound {
		t.Fatalf("err=%v want ErrNotFound", err)
	}
}

func TestGetBookingReturnsCopy(t *testing.T) {
	s := New()
	if err := s.PutBooking(&model.Booking{ID: "b1", MemberID: "m1"}); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetBooking("b1")
	got.MemberID = "MUTATED"
	again, _ := s.GetBooking("b1")
	if again.MemberID != "m1" {
		t.Fatalf("GetBooking returned internal reference: MemberID=%q", again.MemberID)
	}
}

func TestGetPackageReturnsCopy(t *testing.T) {
	s := New()
	if err := s.PutPackage(&model.SessionPackage{ID: "p1", RemainingSessions: 10}); err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetPackage("p1")
	got.RemainingSessions = 999
	again, _ := s.GetPackage("p1")
	if again.RemainingSessions != 10 {
		t.Fatal("GetPackage returned internal reference")
	}
}

func TestUpdateBooking(t *testing.T) {
	s := New()
	if err := s.PutBooking(&model.Booking{ID: "b1", Status: model.StatusPending}); err != nil {
		t.Fatal(err)
	}
	_, err := s.UpdateBooking("b1", func(b *model.Booking) {
		b.Status = model.StatusConfirmed
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetBooking("b1")
	if got.Status != model.StatusConfirmed {
		t.Fatalf("Status=%q want confirmed", got.Status)
	}
}

func TestUpdatePackageAtomic(t *testing.T) {
	s := New()
	if err := s.PutPackage(&model.SessionPackage{ID: "p1", RemainingSessions: 10}); err != nil {
		t.Fatal(err)
	}
	_, err := s.UpdatePackage("p1", func(p *model.SessionPackage) {
		p.RemainingSessions -= 3
	})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.GetPackage("p1")
	if got.RemainingSessions != 7 {
		t.Fatalf("RemainingSessions=%d want 7", got.RemainingSessions)
	}
}

func TestSchedule(t *testing.T) {
	s := New()
	if err := s.PutSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"}); err != nil {
		t.Fatal(err)
	}
	sc, err := s.GetSchedule("c1", "2026-08-18", "10:00")
	if err != nil {
		t.Fatal(err)
	}
	sc.Booked = true
	again, _ := s.GetSchedule("c1", "2026-08-18", "10:00")
	if again.Booked {
		t.Fatal("GetSchedule returned internal reference")
	}
}
