package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/G-Research/armada/internal/armada/api"
)

func TestWatchContext_ProcessEvent(t *testing.T) {
	watchContext := NewWatchContext()

	expected := JobInfo{Status: Pending}

	watchContext.ProcessEvent(&api.JobPendingEvent{JobId: "1"})
	result := watchContext.GetJobInfo("1")

	assert.Equal(t, result, expected)
}

func TestWatchContext_ProcessEvent_UpdatesExisting(t *testing.T) {
	watchContext := NewWatchContext()

	expected := JobInfo{Status: Running}

	watchContext.ProcessEvent(&api.JobPendingEvent{JobId: "1"})
	watchContext.ProcessEvent(&api.JobRunningEvent{JobId: "1"})
	result := watchContext.GetJobInfo("1")

	assert.Equal(t, result, expected)
}

func TestWatchContext_ProcessEvent_SubmittedEventAddsJobToJobInfo(t *testing.T) {
	watchContext := NewWatchContext()

	job := api.Job{
		Id:       "1",
		JobSetId: "job-set-1",
		Queue:    "queue1",
		PodSpec: &v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "Container1",
				},
			},
		},
	}

	expected := JobInfo{
		Status: Submitted,
		Job:    &job,
	}

	watchContext.ProcessEvent(&api.JobSubmittedEvent{JobId: "1", Job: job})
	result := watchContext.GetJobInfo("1")

	assert.Equal(t, result, expected)
}

func TestWatchContext_GetCurrentState(t *testing.T) {
	watchContext := NewWatchContext()

	watchContext.ProcessEvent(&api.JobQueuedEvent{JobId: "1"})
	watchContext.ProcessEvent(&api.JobPendingEvent{JobId: "2"})
	watchContext.ProcessEvent(&api.JobRunningEvent{JobId: "3"})

	expected := map[string]*JobInfo{
		"1": {Status: Queued},
		"2": {Status: Pending},
		"3": {Status: Running},
	}

	result := watchContext.GetCurrentState()

	assert.Equal(t, expected, result)
}

func TestWatchContext_GetCurrentStateSummary(t *testing.T) {
	watchContext := NewWatchContext()

	watchContext.ProcessEvent(&api.JobQueuedEvent{JobId: "1"})
	watchContext.ProcessEvent(&api.JobPendingEvent{JobId: "2"})
	watchContext.ProcessEvent(&api.JobRunningEvent{JobId: "3"})

	expected := "Queued:   1, Leased:   0, Pending:   1, Running:   1, Succeeded:   0, Failed:   0, Cancelled:   0"
	result := watchContext.GetCurrentStateSummary()

	assert.Equal(t, result, expected)
}

func TestWatchContext_GetCurrentStateSummary_IsCorrectlyAlteredOnUpdateToExistingJob(t *testing.T) {
	watchContext := NewWatchContext()

	watchContext.ProcessEvent(&api.JobQueuedEvent{JobId: "1"})
	watchContext.ProcessEvent(&api.JobPendingEvent{JobId: "1"})
	expected := "Queued:   0, Leased:   0, Pending:   1, Running:   0, Succeeded:   0, Failed:   0, Cancelled:   0"
	result := watchContext.GetCurrentStateSummary()
	assert.Equal(t, result, expected)
}

func TestWatchContext_GetNumberOfJobsInStates(t *testing.T) {
	watchContext := NewWatchContext()

	watchContext.ProcessEvent(&api.JobQueuedEvent{JobId: "1"})
	result := watchContext.GetNumberOfJobsInStates([]JobStatus{Queued})

	assert.Equal(t, result, 1)
}

func TestWatchContext_GetNumberOfJobsInStates_IsCorrectlyUpdatedOnUpdateToExistingJobState(t *testing.T) {
	watchContext := NewWatchContext()

	watchContext.ProcessEvent(&api.JobQueuedEvent{JobId: "1"})
	assert.Equal(t, watchContext.GetNumberOfJobsInStates([]JobStatus{Queued}), 1)

	watchContext.ProcessEvent(&api.JobPendingEvent{JobId: "1"})
	assert.Equal(t, watchContext.GetNumberOfJobsInStates([]JobStatus{Queued}), 0)
	assert.Equal(t, watchContext.GetNumberOfJobsInStates([]JobStatus{Pending}), 1)
}
