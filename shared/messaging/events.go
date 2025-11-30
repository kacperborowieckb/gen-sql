package messaging

const (
	DataGenerationQueue = "data_generation_queue"
)

// ProjectCreatedEvent defines the payload for a project.created event
type ProjectCreatedEvent struct {
	ProjectID      string  `json:"projectId"`
	DdlSchema      string  `json:"ddlSchema"`
	Instructions   string  `json:"instructions"`
	RowsToGenerate int32   `json:"rowsToGenerate"`
	Temperature    float32 `json:"temperature"`
}
