package util

import (
	"testing"

	"fitness/internal/model"
)

func mk(id string, st model.BookingStatus, p model.Priority) *model.Booking {
	return &model.Booking{ID: id, Status: st, Priority: p}
}

func TestFilterByStatusNoAliasing(t *testing.T) {
	bookings := []*model.Booking{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusCompleted, model.PriorityLow),
		mk("c", model.StatusPending, model.PriorityLow),
	}
	got := FilterBookingsByStatus(bookings, model.StatusPending)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
	if bookings[0].ID != "a" || bookings[1].ID != "b" || bookings[2].ID != "c" {
		t.Fatalf("FilterBookingsByStatus corrupted input: %v %v %v", bookings[0].ID, bookings[1].ID, bookings[2].ID)
	}
}

func TestFilterActive(t *testing.T) {
	bookings := []*model.Booking{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusCompleted, model.PriorityLow),
		mk("c", model.StatusRetrying, model.PriorityLow),
	}
	got := FilterActive(bookings)
	if len(got) != 2 {
		t.Fatalf("len=%d want 2", len(got))
	}
}

func TestSortByPriority(t *testing.T) {
	bookings := []*model.Booking{
		mk("a", model.StatusPending, model.PriorityLow),
		mk("b", model.StatusPending, model.PriorityUrgent),
		mk("c", model.StatusPending, model.PriorityHigh),
	}
	got := SortByPriority(bookings)
	want := []string{"b", "c", "a"}
	for i := range want {
		if got[i].ID != want[i] {
			t.Fatalf("got[%d]=%s want %s", i, got[i].ID, want[i])
		}
	}
}

func TestSkillCategory(t *testing.T) {
	cases := map[string]string{
		"strength": "strength",
		"yoga":     "yoga",
		"hiit":     "hiit",
		"other":    "default",
	}
	for in, want := range cases {
		if got := SkillCategory(in); got != want {
			t.Fatalf("SkillCategory(%q)=%q want %q", in, got, want)
		}
	}
}
