package jobs

import "time"

// Schedule returns the next run time after now.
type Schedule func(now time.Time) time.Time

func Every(interval time.Duration) Schedule {
	return func(now time.Time) time.Time { return now.Add(interval) }
}

func Weekly(day time.Weekday, hour int) Schedule {
	return func(now time.Time) time.Time {
		now = now.Local()
		days := (int(day) - int(now.Weekday()) + 7) % 7
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local).AddDate(0, 0, days)
		if !next.After(now) {
			next = next.AddDate(0, 0, 7)
		}
		return next
	}
}
