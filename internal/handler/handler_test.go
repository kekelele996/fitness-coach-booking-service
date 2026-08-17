package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fitness/internal/config"
	"fitness/internal/model"
	"fitness/internal/repository"
	"fitness/internal/service"
	"fitness/internal/store"
)

func newServer() http.Handler {
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, config.Load())
	return New(svc).Routes()
}

func seedBase(s *service.Service, repo *repository.Repository) {
	repo.CreateMember(&model.Member{ID: "m1", PackageID: "p-m1"})
	repo.CreatePackage(&model.SessionPackage{ID: "p-m1", MemberID: "m1", TotalSessions: 10, RemainingSessions: 10, Status: model.PackageActive})
	repo.CreateCoach(&model.Coach{ID: "c1", Skills: []string{"strength-cert"}, Available: true})
	repo.CreateSchedule(&model.Schedule{CoachID: "c1", Date: "2026-08-18", TimeSlot: "10:00"})
}

func TestGetMissingBookingReturns404(t *testing.T) {
	h := newServer()
	req := httptest.NewRequest(http.MethodGet, "/bookings/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status=%d want 404", rec.Code)
	}
}

func TestCreateBooking(t *testing.T) {
	st := store.New()
	repo := repository.New(st)
	svc := service.New(repo, config.Load())
	seedBase(svc, repo)
	h := New(svc).Routes()
	body := `{"member_id":"m1","coach_id":"c1","program":"strength","date":"2026-08-18","slot":"10:00","sessions":1,"priority":3}`
	req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status=%d want 201, body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id := created["ID"].(string)
	getReq := httptest.NewRequest(http.MethodGet, "/bookings/"+id, nil)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status=%d want 200", getRec.Code)
	}
}
