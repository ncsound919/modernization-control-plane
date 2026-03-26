package cobol

import (
	"fmt"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/sidecar-gateway/internal/models"
)

// JobResult holds the outcome of a COBOL batch job execution.
type JobResult struct {
	JobID       string            `json:"job_id"`
	ProgramName string            `json:"program_name"`
	Status      string            `json:"status"`
	ReturnCode  int               `json:"return_code"`
	Output      string            `json:"output"`
	ExecutedAt  time.Time         `json:"executed_at"`
	Parameters  map[string]string `json:"parameters,omitempty"`
	DurationMs  int64             `json:"duration_ms"`
}

// Adapter is the stub COBOL/MQ adapter.
type Adapter struct{}

// New returns a new COBOL Adapter.
func New() *Adapter { return &Adapter{} }

// Execute simulates submitting a COBOL batch job via JCL/MQ and returns a
// mock result. In production this would marshal the JCL template, enqueue the
// job on IBM MQ and poll for completion.
func (a *Adapter) Execute(job models.COBOLJob) (*JobResult, error) {
	if job.ProgramName == "" {
		return nil, fmt.Errorf("program_name is required")
	}

	start := time.Now()

	// Simulate a short processing delay.
	time.Sleep(10 * time.Millisecond)

	result := &JobResult{
		JobID:       fmt.Sprintf("JOB%06d", time.Now().UnixMilli()%1_000_000),
		ProgramName: job.ProgramName,
		Status:      "COMPLETED",
		ReturnCode:  0,
		Output: fmt.Sprintf(
			"//STEP001  EXEC PGM=%s\n//SYSOUT   DD  SYSOUT=*\nIEF142I %s - STEP WAS EXECUTED - COND CODE 0000\nIEF373I STEP/STEP001/START 2024.001 00:00:00.000\nIEF374I STEP/STEP001/STOP  2024.001 00:00:00.010  CPU 0MIN 00.01SEC",
			job.ProgramName, job.ProgramName,
		),
		ExecutedAt: start,
		Parameters: job.Parameters,
		DurationMs: time.Since(start).Milliseconds(),
	}

	return result, nil
}
