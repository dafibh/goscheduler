# GoScheduler

GoScheduler is a lightweight, efficient task scheduling library for Go applications. It enables developers to easily schedule recurring tasks with various timing patterns including daily, weekly, monthly, and custom periodic intervals.

## Features

- **Daily scheduling**: Run tasks at specific times every day
- **Weekly scheduling**: Run tasks on specific days of the week
- **Monthly scheduling**: Run tasks on the last working day of the month with holiday awareness
- **Periodic scheduling**: Run tasks at custom intervals with precise timing control
- **Concurrency control**: Limit the maximum number of concurrent task executions
- **Adaptive sleep mechanism**: Efficient CPU usage with dynamic sleep intervals

## Installation

```bash
go get github.com/dafibh/goscheduler
```

## Usage

### Daily Tasks

Schedule a task to run at the same time every day:

```go
package main

import (
    "fmt"
    "github.com/dafibh/goscheduler"
    "time"
)

func main() {
    // Run task every day at 15:30
    goscheduler.ScheduleDailyTask(15, 30, func() {
        fmt.Println("Daily task executed at", time.Now())
        // Your task logic here
    })
    
    // Keep main goroutine alive
    select {}
}
```

### Weekly Tasks

Schedule a task to run on specific days of the week:

```go
// Run task every Monday at 9:00
goscheduler.ScheduleWeeklyTask(time.Monday, 9, 0, func() {
    fmt.Println("Weekly task executed at", time.Now())
    // Your task logic here
})
```

### Monthly Tasks

Schedule a task to run on the last working day of each month, respecting holidays:

```go
// Define holidays
holidays := []time.Time{
    time.Date(2025, time.January, 1, 0, 0, 0, 0, time.Local),  // New Year's Day
    time.Date(2025, time.December, 25, 0, 0, 0, 0, time.Local), // Christmas
}

// Run on the last working day of the month (up to the 28th)
// at 16:30, skipping weekends and holidays
goscheduler.ScheduleMonthlyTask(28, 16, 30, holidays, func() {
    fmt.Println("Monthly task executed at", time.Now())
    // Your monthly reporting or similar task
})
```

### Periodic Tasks

Schedule tasks that run at specific intervals with concurrency control:

```go
// Run every 60 seconds, starting immediately, with max 3 concurrent executions
goscheduler.SchedulePeriodicTask(60, "------------", 3, func() {
    fmt.Println("Periodic task executed at", time.Now())
    time.Sleep(10 * time.Second) // Simulate a long-running task
})

// Run every 5 minutes at the start of the minute, starting at a specific time
// Format: YYMMDDHHmmss with -- for current value
goscheduler.SchedulePeriodicTask(300, "----15--00--", 2, func() {
    fmt.Println("Specific periodic task executed at", time.Now())
})
```

## Start Mask Format

For the `SchedulePeriodicTask` function, the `startMask` parameter uses the following format:

- 12 characters in `YYMMDDHHmmss` format
- Use `--` for any field where you want to use the current time value
- Examples:
  - `------------`: Start immediately
  - `----01000000`: Start at midnight on the 1st day of the current month
  - `--10--120000`: Start at noon on the first day of October of the current year

## How It Works

GoScheduler uses goroutines to run scheduled tasks without blocking the main application. It employs an adaptive sleep mechanism that dynamically adjusts the sleep duration based on the time until the next scheduled task, optimizing CPU usage while ensuring timely task execution.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
