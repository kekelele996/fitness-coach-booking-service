package repository

import (
	"errors"
	"fmt"

	"fitness/internal/model"
	"fitness/internal/store"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrCoachNotFound = errors.New("coach not found")
)

// Repository 数据访问层。
type Repository struct {
	store *store.Store
}

func New(s *store.Store) *Repository {
	return &Repository{store: s}
}

func (r *Repository) CreateMember(m *model.Member) (*model.Member, error) {
	if err := r.store.PutMember(m); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, fmt.Errorf("member %s: %w", m.ID, ErrAlreadyExists)
		}
		return nil, fmt.Errorf("member %s: %w", m.ID, err)
	}
	return m, nil
}

func (r *Repository) FindMember(id string) (*model.Member, error) {
	m, err := r.store.GetMember(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("member %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("member %s: %w", id, err)
	}
	return m, nil
}

func (r *Repository) CreateCoach(c *model.Coach) (*model.Coach, error) {
	if err := r.store.PutCoach(c); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, fmt.Errorf("coach %s: %w", c.ID, ErrAlreadyExists)
		}
		return nil, fmt.Errorf("coach %s: %w", c.ID, err)
	}
	return c, nil
}

func (r *Repository) FindCoach(id string) (*model.Coach, error) {
	c, err := r.store.GetCoach(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("coach %s: %w", id, ErrCoachNotFound)
		}
		return nil, fmt.Errorf("coach %s: %w", id, err)
	}
	return c, nil
}

func (r *Repository) CreatePackage(p *model.SessionPackage) (*model.SessionPackage, error) {
	if err := r.store.PutPackage(p); err != nil {
		return nil, fmt.Errorf("package %s: %w", p.ID, err)
	}
	return p, nil
}

func (r *Repository) FindPackage(id string) (*model.SessionPackage, error) {
	p, err := r.store.GetPackage(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("package %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("package %s: %w", id, err)
	}
	return p, nil
}

func (r *Repository) ListPackages() ([]*model.SessionPackage, error) {
	return r.store.ListPackages(), nil
}

func (r *Repository) UpdatePackage(id string, fn func(*model.SessionPackage)) (*model.SessionPackage, error) {
	p, err := r.store.UpdatePackage(id, fn)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("package %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("package %s: %w", id, err)
	}
	return p, nil
}

func (r *Repository) CreateBooking(b *model.Booking) (*model.Booking, error) {
	if err := r.store.PutBooking(b); err != nil {
		return nil, fmt.Errorf("booking %s: %w", b.ID, err)
	}
	return b, nil
}

func (r *Repository) FindBooking(id string) (*model.Booking, error) {
	b, err := r.store.GetBooking(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("booking %s: %w", id, err)
	}
	return b, nil
}

func (r *Repository) ListBookings() ([]*model.Booking, error) {
	bookings := r.store.ListBookings()
	return append([]*model.Booking(nil), bookings...), nil
}

func (r *Repository) UpdateBooking(id string, fn func(*model.Booking)) (*model.Booking, error) {
	b, err := r.store.UpdateBooking(id, fn)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("booking %s: %w", id, err)
	}
	return b, nil
}

func (r *Repository) CreateSchedule(sc *model.Schedule) (*model.Schedule, error) {
	if err := r.store.PutSchedule(sc); err != nil {
		return nil, fmt.Errorf("schedule %s/%s/%s: %w", sc.CoachID, sc.Date, sc.TimeSlot, err)
	}
	return sc, nil
}

func (r *Repository) FindSchedule(coachID, date, slot string) (*model.Schedule, error) {
	sc, err := r.store.GetSchedule(coachID, date, slot)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("schedule %s/%s/%s: %w", coachID, date, slot, ErrNotFound)
		}
		return nil, fmt.Errorf("schedule %s/%s/%s: %w", coachID, date, slot, err)
	}
	return sc, nil
}

func (r *Repository) UpdateSchedule(coachID, date, slot string, fn func(*model.Schedule)) (*model.Schedule, error) {
	sc, err := r.store.UpdateSchedule(coachID, date, slot, fn)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, fmt.Errorf("schedule %s/%s/%s: %w", coachID, date, slot, ErrNotFound)
		}
		return nil, fmt.Errorf("schedule %s/%s/%s: %w", coachID, date, slot, err)
	}
	return sc, nil
}

func (r *Repository) ListSchedules() ([]*model.Schedule, error) {
	return r.store.ListSchedules(), nil
}
