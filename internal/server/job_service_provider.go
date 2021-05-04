package server

import (
    "app/internal/server/job"
    "app/pkg/queue"
)

var JobSP JobServiceProvider

type JobServiceProvider struct {
}

func (*JobServiceProvider) Register() {
    go queue.NewConsumer("my_job_demo", &job.MyJobDemo{}, 0, 0)
}
