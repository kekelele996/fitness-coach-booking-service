package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"fitness/internal/model"
	"fitness/internal/service"
)

// Server HTTP 层。
type Server struct {
	svc *service.Service
}

func New(svc *service.Service) *Server {
	return &Server{svc: svc}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /bookings", s.listBookings)
	mux.HandleFunc("GET /bookings/{id}", s.getBooking)
	mux.HandleFunc("POST /bookings", s.createBooking)
	mux.HandleFunc("POST /bookings/{id}/confirm", s.confirmBooking)
	mux.HandleFunc("POST /bookings/{id}/cancel", s.cancelBooking)
	return mux
}

type createRequest struct {
	MemberID string `json:"member_id"`
	CoachID  string `json:"coach_id"`
	Program  string `json:"program"`
	Date     string `json:"date"`
	Slot     string `json:"slot"`
	Sessions int    `json:"sessions"`
	Priority int    `json:"priority"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) listBookings(w http.ResponseWriter, r *http.Request) {
	var status *model.BookingStatus
	if raw := r.URL.Query().Get("status"); raw != "" {
		st := model.BookingStatus(raw)
		status = &st
	}
	bookings, err := s.svc.ListBookings(status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list bookings failed")
		return
	}
	writeJSON(w, http.StatusOK, bookings)
}

func (s *Server) getBooking(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	b, err := s.svc.FindBooking(id)
	if err != nil {
		s.writeBookingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) createBooking(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	b, err := s.svc.CreateBooking(req.MemberID, req.CoachID, req.Program, req.Date, req.Slot, req.Sessions, model.Priority(req.Priority))
	if err != nil {
		if errors.Is(err, service.ErrValidation) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrSlotTaken) || errors.Is(err, service.ErrInsufficientSessions) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "create booking failed")
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) confirmBooking(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	b, err := s.svc.ConfirmBooking(id)
	if err != nil {
		s.writeBookingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) cancelBooking(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	b, err := s.svc.CancelBooking(id)
	if err != nil {
		s.writeBookingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) writeBookingError(w http.ResponseWriter, err error) {
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "booking not found")
		return
	}
	if errors.Is(err, service.ErrInvalidTransition) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "internal error")
}
