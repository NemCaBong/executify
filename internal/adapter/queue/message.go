package queue

import "encoding/json"

type SubmissionMessage struct {
	SubmissionID int `json:"submission_id"`
}

func (m SubmissionMessage) ToBytes() []byte {
	data, _ := json.Marshal(m)
	return data
}

func (m *SubmissionMessage) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), m)
}
