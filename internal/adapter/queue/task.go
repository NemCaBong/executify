package queue

import "encoding/json"

const (
	TypeSubmissionSubmit = "submission:submit"
	TypeSubmissionRun    = "submission:run"
)

const (
	QueueSubmit = "submit"
	QueueRun    = "run"
)

type SubmissionPayload struct {
	SubmissionID int `json:"submission_id"`
	// LogCommand, when true, makes the worker log every isolate command line
	// for this submission. Toggled per request via the X-Enable-Log-Command header.
	EnableCommandLog bool `json:"log_command,omitempty"`
}

func (p SubmissionPayload) Marshal() ([]byte, error) {
	return json.Marshal(p)
}

func UnmarshalSubmissionPayload(data []byte) (SubmissionPayload, error) {
	var p SubmissionPayload
	if err := json.Unmarshal(data, &p); err != nil {
		return SubmissionPayload{}, err
	}
	return p, nil
}
