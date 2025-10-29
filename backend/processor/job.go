package processor

import (
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

type JobPriority int

const (
	PriorityLow JobPriority = iota
	PriorityMedium
	PriorityHigh
)

type Job struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Priority    JobPriority    `json:"priority"`
	Status      JobStatus      `json:"status"`
	FilePath    string         `json:"file_path"`
	Bucket      string         `json:"bucket"`
	ObjectName  string         `json:"object_name"`
	CreatedAt   time.Time      `json:"created_at"`
	StartedAt   *time.Time     `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Error       string         `json:"error,omitempty"`
	Result      any            `json:"result,omitempty"`
	Progress    float64        `json:"progress"`
	Metadata    map[string]any `json:"metadata"`
}

type JobResult struct {
	Success        bool           `json:"success"`
	ExtractedFiles []string       `json:"extracted_files,omitempty"`
	FileInfo       map[string]any `json:"file_info,omitempty"`
	ProcessingTime time.Duration  `json:"processing_time"`
	Message        string         `json:"message"`
}

func NewJob(jobType, filePath, bucket, objectName string, priority JobPriority) *Job {
	return &Job{
		ID:         uuid.New().String(),
		Type:       jobType,
		Priority:   priority,
		Status:     JobStatusPending,
		FilePath:   filePath,
		Bucket:     bucket,
		ObjectName: objectName,
		CreatedAt:  time.Now(),
		Progress:   0.0,
		Metadata:   make(map[string]any),
	}
}

func (j *Job) Start() {
	now := time.Now()
	j.Status = JobStatusProcessing
	j.StartedAt = &now
}

func (j *Job) Complete(result JobResult) {
	now := time.Now()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
	j.Result = result
	j.Progress = 100.0
}

func (j *Job) Fail(err error) {
	now := time.Now()
	j.Status = JobStatusFailed
	j.CompletedAt = &now
	j.Error = err.Error()
}

func (j *Job) Cancel() {
	now := time.Now()
	j.Status = JobStatusCancelled
	j.CompletedAt = &now
}

func (j *Job) UpdateProgress(progress float64) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	j.Progress = progress
}

func (j *Job) GetDuration() time.Duration {
	if j.StartedAt == nil {
		return 0
	}
	end := time.Now()
	if j.CompletedAt != nil {
		end = *j.CompletedAt
	}
	return end.Sub(*j.StartedAt)
}

func (p JobPriority) String() string {
	switch p {
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

func ParsePriority(s string) JobPriority {
	switch s {
	case "high":
		return PriorityHigh
	case "medium":
		return PriorityMedium
	case "low":
		return PriorityLow
	default:
		return PriorityMedium
	}
}
