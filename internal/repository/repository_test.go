package repository

import (
	"errors"
	"testing"

	"fitness/internal/model"
	"fitness/internal/store"
)

func newRepo() *Repository {
	return New(store.New())
}

func TestFindBookingWrapsNotFound(t *testing.T) {
	r := newRepo()
	_, err := r.FindBooking("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound)=false, err=%v", err)
	}
}

func TestFindCoachWrapsNotFound(t *testing.T) {
	r := newRepo()
	_, err := r.FindCoach("missing")
	if !errors.Is(err, ErrCoachNotFound) {
		t.Fatalf("errors.Is(err, ErrCoachNotFound)=false, err=%v", err)
	}
}

func TestCreateDuplicate(t *testing.T) {
	r := newRepo()
	if _, err := r.CreateMember(&model.Member{ID: "m1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := r.CreateMember(&model.Member{ID: "m1"}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("dup err=%v want ErrAlreadyExists", err)
	}
}
