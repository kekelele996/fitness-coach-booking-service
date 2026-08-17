package store

import (
	"errors"
	"sync"

	"fitness/internal/model"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Store 内存存储，读写锁保护。
type Store struct {
	mu         sync.RWMutex
	members    map[string]*model.Member
	memberIDs  []string
	coaches    map[string]*model.Coach
	coachIDs   []string
	packages   map[string]*model.SessionPackage
	packageIDs []string
	bookings   map[string]*model.Booking
	bookingIDs []string
	schedules  map[string]*model.Schedule
}

func New() *Store {
	return &Store{
		members:    make(map[string]*model.Member),
		coaches:    make(map[string]*model.Coach),
		packages:   make(map[string]*model.SessionPackage),
		bookings:   make(map[string]*model.Booking),
		schedules:  make(map[string]*model.Schedule),
	}
}

func cloneMember(m *model.Member) *model.Member {
	if m == nil {
		return nil
	}
	c := *m
	return &c
}

func cloneCoach(c *model.Coach) *model.Coach {
	if c == nil {
		return nil
	}
	cc := *c
	cc.Skills = append([]string(nil), c.Skills...)
	return &cc
}

func clonePackage(p *model.SessionPackage) *model.SessionPackage {
	if p == nil {
		return nil
	}
	c := *p
	return &c
}

func cloneBooking(b *model.Booking) *model.Booking {
	if b == nil {
		return nil
	}
	c := *b
	return &c
}

func scheduleKey(coachID, date, slot string) string {
	return coachID + "|" + date + "|" + slot
}

func (s *Store) PutMember(m *model.Member) error {
	if m == nil || m.ID == "" {
		return errors.New("invalid member")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.members[m.ID]; ok {
		return ErrAlreadyExists
	}
	s.members[m.ID] = cloneMember(m)
	s.memberIDs = append(s.memberIDs, m.ID)
	return nil
}

func (s *Store) GetMember(id string) (*model.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneMember(m), nil
}

func (s *Store) ListMembers() []*model.Member {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Member, 0, len(s.memberIDs))
	for _, id := range s.memberIDs {
		out = append(out, cloneMember(s.members[id]))
	}
	return out
}

func (s *Store) PutCoach(c *model.Coach) error {
	if c == nil || c.ID == "" {
		return errors.New("invalid coach")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.coaches[c.ID]; ok {
		return ErrAlreadyExists
	}
	s.coaches[c.ID] = cloneCoach(c)
	s.coachIDs = append(s.coachIDs, c.ID)
	return nil
}

func (s *Store) GetCoach(id string) (*model.Coach, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.coaches[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneCoach(c), nil
}

func (s *Store) ListCoaches() []*model.Coach {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Coach, 0, len(s.coachIDs))
	for _, id := range s.coachIDs {
		out = append(out, cloneCoach(s.coaches[id]))
	}
	return out
}

func (s *Store) PutPackage(p *model.SessionPackage) error {
	if p == nil || p.ID == "" {
		return errors.New("invalid package")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.packages[p.ID]; ok {
		return ErrAlreadyExists
	}
	s.packages[p.ID] = clonePackage(p)
	s.packageIDs = append(s.packageIDs, p.ID)
	return nil
}

func (s *Store) GetPackage(id string) (*model.SessionPackage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.packages[id]
	if !ok {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *Store) ListPackages() []*model.SessionPackage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.SessionPackage, 0, len(s.packageIDs))
	for _, id := range s.packageIDs {
		out = append(out, clonePackage(s.packages[id]))
	}
	return out
}

func (s *Store) UpdatePackage(id string, fn func(*model.SessionPackage)) (*model.SessionPackage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.packages[id]
	if !ok {
		return nil, ErrNotFound
	}
	fn(p)
	return p, nil
}

func (s *Store) PutBooking(b *model.Booking) error {
	if b == nil || b.ID == "" {
		return errors.New("invalid booking")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.bookings[b.ID]; ok {
		return ErrAlreadyExists
	}
	s.bookings[b.ID] = cloneBooking(b)
	s.bookingIDs = append(s.bookingIDs, b.ID)
	return nil
}

func (s *Store) GetBooking(id string) (*model.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bookings[id]
	if !ok {
		return nil, ErrNotFound
	}
	return b, nil
}

func (s *Store) ListBookings() []*model.Booking {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Booking, 0, len(s.bookingIDs))
	for _, id := range s.bookingIDs {
		out = append(out, s.bookings[id])
	}
	return out
}

func (s *Store) UpdateBooking(id string, fn func(*model.Booking)) (*model.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	if !ok {
		return nil, ErrNotFound
	}
	fn(b)
	return b, nil
}

func (s *Store) PutSchedule(sc *model.Schedule) error {
	if sc == nil || sc.CoachID == "" {
		return errors.New("invalid schedule")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := scheduleKey(sc.CoachID, sc.Date, sc.TimeSlot)
	if _, ok := s.schedules[key]; ok {
		return ErrAlreadyExists
	}
	cc := *sc
	s.schedules[key] = &cc
	return nil
}

func (s *Store) GetSchedule(coachID, date, slot string) (*model.Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.schedules[scheduleKey(coachID, date, slot)]
	if !ok {
		return nil, ErrNotFound
	}
	c := *sc
	return &c, nil
}

func (s *Store) UpdateSchedule(coachID, date, slot string, fn func(*model.Schedule)) (*model.Schedule, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sc, ok := s.schedules[scheduleKey(coachID, date, slot)]
	if !ok {
		return nil, ErrNotFound
	}
	fn(sc)
	c := *sc
	return &c, nil
}

func (s *Store) ListSchedules() []*model.Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*model.Schedule, 0, len(s.schedules))
	for _, sc := range s.schedules {
		c := *sc
		out = append(out, &c)
	}
	return out
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.bookingIDs)
}
