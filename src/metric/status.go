package metric

type Status = string

const (
	StatusCancelled Status = "cancelled"
	StatusCompleted Status = "completed"
	StatusError     Status = "error"
	StatusSuccess   Status = "success"
)
