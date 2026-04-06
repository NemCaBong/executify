package queue

import "encoding/json"

type SubmissionMessage struct {
	SubmissionID int `json:"submission_id"`
}

func (m SubmissionMessage) String() string {
	data, _ := json.Marshal(m)
	return string(data)
}

func (m *SubmissionMessage) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), m)
}
