package runner

import "time"

type Summary struct {
	Duration       time.Duration `json:"duration"`
	SuccessCounter int           `json:"success_counter"`
	ErrorCounter   int           `json:"error_counter"`
	EachRun        []RunStats    `json:"each_run"`
}

type RunStats struct {
	Duration  time.Duration `json:"duration"`
	StartTime time.Time     `json:"start_time"`
	Failed    bool          `json:"failed"`
}

type Runner interface {
	Run(concurrency int) (Summary, error)
}
