package model

import "time"

// Priority 数值越大越紧急（用于教练排期优先级）。
type Priority int

const (
	PriorityLow Priority = iota + 1
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

// BookingStatus 预约状态。
type BookingStatus string

const (
	StatusPending    BookingStatus = "pending"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusCompleted  BookingStatus = "completed"
	StatusCancelled  BookingStatus = "cancelled"
	StatusExpired    BookingStatus = "expired"
	StatusFailed     BookingStatus = "failed"
	StatusRetrying   BookingStatus = "retrying"
)

// ActiveStatuses 仍在进行中的状态集合。
var ActiveStatuses = map[BookingStatus]bool{
	StatusPending:   true,
	StatusConfirmed: true,
	StatusRetrying:  true,
}

// PackageStatus 课时包状态。
type PackageStatus string

const (
	PackageActive   PackageStatus = "active"
	PackageExhausted PackageStatus = "exhausted"
	PackageFrozen    PackageStatus = "frozen"
)

// Member 会员。
type Member struct {
	ID       string
	Name     string
	Level    int
	PackageID string
}

// Coach 教练。
type Coach struct {
	ID        string
	Name      string
	Skills    []string
	Available bool
}

// SessionPackage 课时包。
type SessionPackage struct {
	ID               string
	MemberID         string
	TotalSessions    int
	RemainingSessions int
	Status           PackageStatus
}

// Booking 预约单。
type Booking struct {
	ID          string
	MemberID    string
	CoachID     string
	PackageID   string
	ScheduledAt time.Time
	Sessions    int
	Priority    Priority
	Status      BookingStatus
	Attempts    int
	MaxAttempts int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Schedule 教练排期槽位。
type Schedule struct {
	CoachID   string
	Date      string
	TimeSlot  string
	Booked    bool
}

// transitions 合法状态迁移表。
var transitions = map[BookingStatus][]BookingStatus{
	StatusPending:   {StatusConfirmed, StatusExpired, StatusFailed},
	StatusConfirmed: {StatusCompleted, StatusCancelled, StatusFailed},
	StatusFailed:    {StatusRetrying},
	StatusRetrying:  {StatusConfirmed, StatusFailed, StatusCancelled},
	StatusCompleted: {},
	StatusCancelled: {},
	StatusExpired:   {},
}

// CanTransition 判断 from 能否直接迁移到 to。
func CanTransition(from, to BookingStatus) bool {
	for _, t := range transitions[from] {
		if t == to {
			return true
		}
	}
	return false
}

// PriorityRank 返回优先级整数权重。
func PriorityRank(p Priority) int {
	return int(p)
}

// HasSkill 判断教练是否具备技能。
func (c *Coach) HasSkill(skill string) bool {
	for _, s := range c.Skills {
		if s == skill {
			return true
		}
	}
	return false
}
