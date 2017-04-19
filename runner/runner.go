package runner

import "time"

type Summary struct {
	Commands       map[int]Command `json:"commands"`
	Duration       time.Duration   `json:"duration"`
	SuccessCounter int             `json:"success_counter"`
	ErrorCounter   int             `json:"error_counter"`
	EachRun        []RunStats      `json:"each_run"`
}

type RunStats struct {
	Command   int           `json:"command"`
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	Failed    bool          `json:"failed"`
}

type Command struct {
	Exec     string `json:"exec"`
	RunCount int    `json:"run_count"`
}

type Runner interface {
	Run(concurrency int) (Summary, error)
}
