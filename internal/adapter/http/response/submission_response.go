package response

type SubmitResponse struct {
	Message string `json:"message"`
	Data    struct {
		ID int `json:"id"`
	} `json:"data"`
}

func NewSubmitResponse(id int) *SubmitResponse {
	return &SubmitResponse{
		Message: "Submission created successfully",
		Data: struct {
			ID int `json:"id"`
		}{
			ID: id,
		},
	}
}
