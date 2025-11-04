package messaging

const (
	DataGenerationQueue = "data_generation_queue"
)

// ProjectCreatedEvent defines the payload for a project.created event
type ProjectCreatedEvent struct {
	ProjectID              string `json:"projectId"`
	DdlSchema              string `json:"ddlSchema"`
	GenerationInstructions string `json:"generationInstructions"`
	MaxRows                int32  `json:"maxRows"`
}
