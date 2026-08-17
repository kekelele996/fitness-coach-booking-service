package util

import (
	"sort"

	"fitness/internal/model"
)

// FilterBookingsByStatus 返回状态匹配的预约，结果为新切片。
func FilterBookingsByStatus(bookings []*model.Booking, status model.BookingStatus) []*model.Booking {
	out := make([]*model.Booking, 0, len(bookings))
	for _, b := range bookings {
		if b.Status == status {
			out = append(out, b)
		}
	}
	return out
}

// FilterActive 返回仍进行中的预约。
func FilterActive(bookings []*model.Booking) []*model.Booking {
	out := make([]*model.Booking, 0, len(bookings))
	for _, b := range bookings {
		if b.Status == model.StatusPending || b.Status == model.StatusConfirmed {
			out = append(out, b)
		}
	}
	return out
}

// SortByPriority 按优先级降序排序，返回新切片。
func SortByPriority(bookings []*model.Booking) []*model.Booking {
	out := make([]*model.Booking, len(bookings))
	copy(out, bookings)
	sort.SliceStable(out, func(i, j int) bool {
		return model.PriorityRank(out[i].Priority) > model.PriorityRank(out[j].Priority)
	})
	return out
}

// TopByPriority 返回优先级最高的 n 条。
func TopByPriority(bookings []*model.Booking, n int) []*model.Booking {
	if n <= 0 {
		return []*model.Booking{}
	}
	sorted := SortByPriority(bookings)
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	return sorted
}

// SkillCategory 从训练项目推导所需教练技能。
func SkillCategory(program string) string {
	switch program {
	case "strength":
		return "strength"
	case "yoga":
		return "yoga"
	case "hiit":
		return "hiit"
	default:
		return "default"
	}
}
