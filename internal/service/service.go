package service

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"fitness/internal/config"
	"fitness/internal/model"
	"fitness/internal/repository"
	"fitness/internal/util"
)

var (
	ErrNotFound        = errors.New("booking not found")
	ErrCoachNotFound   = errors.New("coach not found")
	ErrValidation      = errors.New("invalid booking request")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrNoCoach         = errors.New("no available coach")
	ErrInsufficientSessions = errors.New("insufficient sessions")
	ErrSlotTaken       = errors.New("schedule slot already booked")
)

var idCounter atomic.Int64

func generateID() string {
	return fmt.Sprintf("bk-%d", idCounter.Add(1))
}

// Dispatcher 根据训练项目给出所需技能标签。
type Dispatcher interface {
	SkillFor(program string) string
}

// SkillDispatcher 默认实现。
type SkillDispatcher struct {
	routes   map[string]string
	fallback string
}

func NewSkillDispatcher(routes map[string]string) *SkillDispatcher {
	if routes == nil {
		routes = map[string]string{}
	}
	return &SkillDispatcher{routes: routes, fallback: "general"}
}

func (d *SkillDispatcher) SkillFor(program string) string {
	cat := util.SkillCategory(program)
	if skill, ok := d.routes[cat]; ok {
		return skill
	}
	d.routes[cat] = d.fallback
	return d.fallback
}

// Service 业务逻辑层。
type Service struct {
	repo       *repository.Repository
	cfg        *config.Config
	dispatcher Dispatcher
}

func New(repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		repo:       repo,
		cfg:        cfg,
		dispatcher: NewSkillDispatcher(cfg.SkillRoutes),
	}
}

// CreateMember 创建会员并绑定课时包。
func (s *Service) CreateMember(name string, totalSessions int) (*model.Member, error) {
	if name == "" || totalSessions <= 0 {
		return nil, ErrValidation
	}
	id := "m-" + generateID()
	m := &model.Member{ID: id, Name: name, Level: 1, PackageID: "p-" + id}
	if _, err := s.repo.CreateMember(m); err != nil {
		return nil, err
	}
	p := &model.SessionPackage{
		ID:               "p-" + id,
		MemberID:         id,
		TotalSessions:    totalSessions,
		RemainingSessions: totalSessions,
		Status:           model.PackageActive,
	}
	if _, err := s.repo.CreatePackage(p); err != nil {
		return nil, err
	}
	return m, nil
}

// CreateBooking 创建预约：校验会员/教练/档期/课时，扣减课时，生成预约。
func (s *Service) CreateBooking(memberID, coachID, program, date, slot string, sessions int, priority model.Priority) (*model.Booking, error) {
	if memberID == "" || coachID == "" || sessions <= 0 {
		return nil, ErrValidation
	}
	if priority < model.PriorityLow || priority > model.PriorityUrgent {
		return nil, ErrValidation
	}
	member, err := s.repo.FindMember(memberID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("member %s: %w", memberID, ErrNotFound)
		}
		return nil, err
	}
	coach, err := s.repo.FindCoach(coachID)
	if err != nil {
		if errors.Is(err, repository.ErrCoachNotFound) {
			return nil, fmt.Errorf("coach %s: %w", coachID, ErrCoachNotFound)
		}
		return nil, err
	}
	if !coach.Available {
		return nil, ErrNoCoach
	}
	required := s.dispatcher.SkillFor(program)
	if !coach.HasSkill(required) {
		return nil, ErrNoCoach
	}
	sc, err := s.repo.FindSchedule(coachID, date, slot)
	if err != nil {
		return nil, fmt.Errorf("schedule %s/%s/%s: %w", coachID, date, slot, ErrSlotTaken)
	}
	if sc.Booked {
		return nil, ErrSlotTaken
	}
	// 扣减课时（原子）。
	pkg, err := s.repo.FindPackage(member.PackageID)
	if err != nil {
		return nil, err
	}
	if pkg.RemainingSessions < sessions {
		return nil, ErrInsufficientSessions
	}
	if _, err := s.repo.UpdatePackage(pkg.ID, func(p *model.SessionPackage) {
		p.RemainingSessions -= sessions
		if p.RemainingSessions == 0 {
			p.Status = model.PackageExhausted
		}
	}); err != nil {
		return nil, err
	}
	if _, err := s.repo.UpdateSchedule(coachID, date, slot, func(s *model.Schedule) {
		s.Booked = true
	}); err != nil {
		return nil, err
	}
	now := time.Now()
	b := &model.Booking{
		ID:          generateID(),
		MemberID:    memberID,
		CoachID:     coachID,
		PackageID:   pkg.ID,
		ScheduledAt: now.Add(24 * time.Hour),
		Sessions:    sessions,
		Priority:    priority,
		Status:      model.StatusPending,
		Attempts:    0,
		MaxAttempts: s.cfg.RetryLimit,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repo.CreateBooking(b)
}

// ConfirmBooking 确认预约。
func (s *Service) ConfirmBooking(bookingID string) (*model.Booking, error) {
	b, err := s.repo.FindBooking(bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", bookingID, ErrNotFound)
		}
		return nil, err
	}
	if !model.CanTransition(b.Status, model.StatusConfirmed) {
		return nil, fmt.Errorf("booking %s from %s: %w", bookingID, b.Status, ErrInvalidTransition)
	}
	return s.repo.UpdateBooking(bookingID, func(bb *model.Booking) {
		bb.Status = model.StatusConfirmed
		bb.UpdatedAt = time.Now()
	})
}

// CompleteBooking 完成预约。
func (s *Service) CompleteBooking(bookingID string) (*model.Booking, error) {
	b, err := s.repo.FindBooking(bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", bookingID, ErrNotFound)
		}
		return nil, err
	}
	if !model.CanTransition(b.Status, model.StatusCompleted) {
		return nil, fmt.Errorf("booking %s from %s: %w", bookingID, b.Status, ErrInvalidTransition)
	}
	return s.repo.UpdateBooking(bookingID, func(bb *model.Booking) {
		bb.Status = model.StatusCompleted
		bb.UpdatedAt = time.Now()
	})
}

// CancelBooking 取消预约并退还课时。
func (s *Service) CancelBooking(bookingID string) (*model.Booking, error) {
	b, err := s.repo.FindBooking(bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", bookingID, ErrNotFound)
		}
		return nil, err
	}
	if !model.CanTransition(b.Status, model.StatusCancelled) {
		return nil, fmt.Errorf("booking %s from %s: %w", bookingID, b.Status, ErrInvalidTransition)
	}
	if _, err := s.repo.UpdatePackage(b.PackageID, func(p *model.SessionPackage) {
		p.RemainingSessions += b.Sessions
		p.Status = model.PackageActive
	}); err != nil {
		return nil, err
	}
	return s.repo.UpdateBooking(bookingID, func(bb *model.Booking) {
		bb.Status = model.StatusCancelled
		bb.UpdatedAt = time.Now()
	})
}

// RetryBooking 把失败预约置为 retrying。
func (s *Service) RetryBooking(bookingID string) (*model.Booking, error) {
	b, err := s.repo.FindBooking(bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", bookingID, ErrNotFound)
		}
		return nil, err
	}
	if !model.CanTransition(b.Status, model.StatusRetrying) {
		return nil, fmt.Errorf("booking %s from %s: %w", bookingID, b.Status, ErrInvalidTransition)
	}
	return s.repo.UpdateBooking(bookingID, func(bb *model.Booking) {
		bb.Status = model.StatusRetrying
		bb.UpdatedAt = time.Now()
	})
}

// ListBookings 查询预约（可选按状态过滤），按优先级排序。
func (s *Service) ListBookings(status *model.BookingStatus) ([]*model.Booking, error) {
	bookings, err := s.repo.ListBookings()
	if err != nil {
		return nil, err
	}
	if status == nil {
		return util.SortByPriority(bookings), nil
	}
	filtered := util.FilterBookingsByStatus(bookings, *status)
	return util.SortByPriority(filtered), nil
}

// Stats 统计各状态预约数量。
func (s *Service) Stats() map[model.BookingStatus]int {
	bookings, _ := s.repo.ListBookings()
	result := map[model.BookingStatus]int{}
	for _, st := range []model.BookingStatus{
		model.StatusPending,
		model.StatusConfirmed,
		model.StatusCompleted,
		model.StatusCancelled,
		model.StatusExpired,
		model.StatusFailed,
		model.StatusRetrying,
	} {
		result[st] = len(util.FilterBookingsByStatus(bookings, st))
	}
	return result
}

// ActiveCount 并发统计进行中的预约数。
func (s *Service) ActiveCount() int {
	bookings, _ := s.repo.ListBookings()
	if len(bookings) == 0 {
		return 0
	}
	var wg sync.WaitGroup
	var count atomic.Int64
	step := (len(bookings) + 3) / 4
	if step < 1 {
		step = 1
	}
	for i := 0; i < len(bookings); i += step {
		end := i + step
		if end > len(bookings) {
			end = len(bookings)
		}
		go func(chunk []*model.Booking) {
			wg.Add(1)
			defer wg.Done()
			for _, b := range chunk {
				if model.ActiveStatuses[b.Status] {
					count.Add(1)
				}
			}
		}(bookings[i:end])
	}
	wg.Wait()
	return int(count.Load())
}

// FindBooking 查询单条预约。
func (s *Service) FindBooking(bookingID string) (*model.Booking, error) {
	b, err := s.repo.FindBooking(bookingID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("booking %s: %w", bookingID, ErrNotFound)
		}
		return nil, fmt.Errorf("lookup booking %s: %w", bookingID, err)
	}
	return b, nil
}
