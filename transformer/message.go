package transformer

// Payload represents JSON Message payload.
type Payload map[string]interface{}

// Message represents a JSON messages.
type Message struct {
	Project   string  `json:"project,omitempty" db:"project" bson:"project"`
	Created   int64   `json:"created,omitempty" db:"created" bson:"created"`
	Subtopic  string  `json:"subtopic,omitempty" db:"subtopic" bson:"subtopic,omitempty"`
	Publisher string  `json:"publisher,omitempty" db:"publisher" bson:"publisher"`
	Protocol  string  `json:"protocol,omitempty" db:"protocol" bson:"protocol"`
	Payload   Payload `json:"payload,omitempty" db:"payload" bson:"payload,omitempty"`
}

// Messages represents a list of JSON messages.
type Messages struct {
	Data   []Message
	Format string
}
