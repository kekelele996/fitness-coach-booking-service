package worker

import (
	"context"
	"errors"
	"time"

	"fitness/internal/model"
	"fitness/internal/repository"
	"fitness/internal/service"
)

// Executor 调度器依赖的能力。
type Executor interface {
	RetryBooking(bookingID string) (*model.Booking, error)
	ConfirmBooking(bookingID string) (*model.Booking, error)
}

// Scheduler 后台调度器：过期未确认的预约置为 expired，失败预约重试。
type Scheduler struct {
	repo         *repository.Repository
	exec         Executor
	pollInterval time.Duration
}

func New(repo *repository.Repository, exec Executor, pollInterval time.Duration) *Scheduler {
	if pollInterval <= 0 {
		pollInterval = 500 * time.Millisecond
	}
	return &Scheduler{repo: repo, exec: exec, pollInterval: pollInterval}
}

// Tick 执行一轮调度。
func (sch *Scheduler) Tick(ctx context.Context) (expired, retried int) {
	if ctx.Err() != nil {
		return 0, 0
	}
	bookings, err := sch.repo.ListBookings()
	if err != nil {
		return 0, 0
	}
	now := time.Now()
	for _, b := range bookings {
		if b.Status == model.StatusPending && now.After(b.ScheduledAt) {
			if _, err := sch.repo.UpdateBooking(b.ID, func(bb *model.Booking) {
				bb.Status = model.StatusExpired
				bb.UpdatedAt = now
			}); err == nil {
				expired++
			} else if errors.Is(err, service.ErrCoachNotFound) {
				continue
			}
		}
	}

	bookings, err = sch.repo.ListBookings()
	if err != nil {
		return expired, retried
	}
	for _, b := range bookings {
		if b.Status == model.StatusFailed && b.Attempts < b.MaxAttempts {
			if _, err := sch.exec.RetryBooking(b.ID); err == nil {
				retried++
			} else if errors.Is(err, service.ErrCoachNotFound) {
				continue
			}
		}
	}
	return expired, retried
}

// Run 周期执行 Tick。
func (sch *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(sch.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			sch.Tick(ctx)
		}
	}
}
