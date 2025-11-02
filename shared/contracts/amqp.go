package contracts

type AmqpMessage struct {
	OwnerId string `json:"ownerId"`
	Data    []byte `json:"data"`
}

const (
	ProjectCreatedRoutingKey = "project.created"
)
