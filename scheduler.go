package goscheduler

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// ScheduleDailyTask schedules a task that runs daily at the specified hour and minute
func ScheduleDailyTask(hour, minute int, task func()) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			// Poll every few seconds until the desired time is reached
			for time.Now().Before(next) {
				adaptiveSleep(next)
			}

			task()
		}
	}()
}

// ScheduleWeeklyTask schedules a task that runs weekly on the specified day, hour and minute
func ScheduleWeeklyTask(dayOfWeek time.Weekday, hour, minute int, task func()) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

			// Find the next correct weekday
			for next.Weekday() != dayOfWeek || now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			// Polling to ensure we don't miss the exact time
			for time.Now().Before(next) {
				adaptiveSleep(next)
			}

			task()
		}
	}()
}

// ScheduleMonthlyTask schedules a task that runs monthly on the specified day, hour and minute
func ScheduleMonthlyTask(maxDay, hour, minute int, holidays []time.Time, task func()) {
	isHoliday := func(date time.Time) bool {
		for _, holiday := range holidays {
			if date.Year() == holiday.Year() && date.YearDay() == holiday.YearDay() {
				return true
			}
		}
		return false
	}

	findLastWorkingDay := func(year int, month time.Month) time.Time {
		// Only consider dates up to and including the maxDay
		for day := maxDay; day >= 1; day-- {
			date := time.Date(year, month, day, hour, minute, 0, 0, time.Local)
			if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday && !isHoliday(date) {
				return date
			}
		}
		// Fallback in case every date is blocked (e.g., entire 1-maxDay are holidays/weekends)
		return time.Date(year, month, 1, hour, minute, 0, 0, time.Local)
	}

	go func() {
		for {
			now := time.Now()
			year, month := now.Year(), now.Month()

			lastWorkingDay := findLastWorkingDay(year, month)
			if now.After(lastWorkingDay) {
				// Move to next month if we've passed this month's window
				if month == 12 {
					year++
					month = 1
				} else {
					month++
				}
				lastWorkingDay = findLastWorkingDay(year, month)
			}

			// Polling to ensure we don't miss the exact time
			for time.Now().Before(lastWorkingDay) {
				adaptiveSleep(lastWorkingDay)
			}

			task()
		}
	}()
}

// SchedulePeriodicTask schedules a periodic task with custom interval and start mask
func SchedulePeriodicTask(intervalSeconds int, startMask string, maxWorkers int, task func()) {
	if intervalSeconds <= 0 {
		log.Println("Invalid interval, must be greater than 0")
		return
	}

	next, err := parseStartMask(startMask)
	if err != nil {
		log.Println(err)
		return
	}

	taskQueue := make(chan struct{}, maxWorkers)

	go func() {
		// Precise polling with AdaptiveSleep
		for time.Now().Before(next) {
			adaptiveSleep(next)
		}

		// Align the ticker start time with 'next'
		for {
			if len(taskQueue) < maxWorkers {
				taskQueue <- struct{}{}
				go func() {
					task()
					<-taskQueue
				}()
			} else {
				log.Println("Skipping task: Too many running tasks")
			}

			// Calculate exact next interval
			next = next.Add(time.Duration(intervalSeconds) * time.Second)
			for time.Now().Before(next) {
				adaptiveSleep(next)
			}
		}
	}()
}

// [Helper] to parse a start mask in the format YYMMDDHHmmss with optional '--' for missing fields
func parseStartMask(startMask string) (time.Time, error) {
	if len(startMask) != 12 {
		return time.Time{}, fmt.Errorf("invalid start mask format, expected YYMMDDHHmmss with optional '--'")
	}
	now := time.Now()

	// If the mask is entirely "------------", run immediately
	if startMask == "------------" {
		return now, nil
	}

	// Helper to parse a two-character field or use the default.
	parse := func(part string, def, min, max int, name string) (int, error) {
		if part == "--" {
			return def, nil
		}
		val, err := strconv.Atoi(part)
		if err != nil || val < min || val > max {
			return 0, fmt.Errorf("invalid %s in start mask", name)
		}
		return val, nil
	}

	year := now.Year()
	if startMask[0:2] != "--" {
		y, err := strconv.Atoi(startMask[0:2])
		if err != nil || y < 0 {
			return time.Time{}, fmt.Errorf("invalid year in start mask")
		}
		year = 2000 + y
	}

	month, err := parse(startMask[2:4], int(now.Month()), 1, 12, "month")
	if err != nil {
		return time.Time{}, err
	}
	day, err := parse(startMask[4:6], now.Day(), 1, 31, "day")
	if err != nil {
		return time.Time{}, err
	}
	hour, err := parse(startMask[6:8], now.Hour(), 0, 23, "hour")
	if err != nil {
		return time.Time{}, err
	}
	minute, err := parse(startMask[8:10], now.Minute(), 0, 59, "minute")
	if err != nil {
		return time.Time{}, err
	}
	second, err := parse(startMask[10:12], now.Second(), 0, 59, "second")
	if err != nil {
		return time.Time{}, err
	}

	next := time.Date(year, time.Month(month), day, hour, minute, second, 0, now.Location())
	if now.After(next) {
		switch {
		case startMask[2:4] != "--":
			next = next.AddDate(1, 0, 0)
		case startMask[4:6] != "--":
			next = next.AddDate(0, 1, 0)
		case startMask[6:8] != "--":
			next = next.AddDate(0, 0, 1)
		case startMask[8:10] != "--":
			next = next.Add(time.Hour)
		default:
			next = next.Add(time.Minute)
		}
	}
	return next, nil
}

// [Helper] to sleep until the target time, with adaptive intervals
func adaptiveSleep(target time.Time) {
	diff := time.Until(target)

	if diff <= 0 {
		return
	}

	switch {
	case diff > 48*time.Hour:
		time.Sleep(12 * time.Hour)
	case diff > 12*time.Hour:
		time.Sleep(3 * time.Hour)
	case diff > 3*time.Hour:
		time.Sleep(1 * time.Hour)
	case diff > time.Hour:
		time.Sleep(15 * time.Minute)
	case diff > 10*time.Minute:
		time.Sleep(5 * time.Minute)
	case diff > time.Minute:
		time.Sleep(30 * time.Second)
	default:
		time.Sleep(1 * time.Second)
	}
}
