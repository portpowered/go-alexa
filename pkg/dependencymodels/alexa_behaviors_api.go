package alexamodels

// Behavior models for Alexa behaviors API

// BehaviorPreviewRequest represents a request to preview/run a behavior
type BehaviorPreviewRequest struct {
	BehaviorID   string `json:"behaviorId"`   // Typically "PREVIEW"
	SequenceJSON string `json:"sequenceJson"` // JSON string of the sequence
	Status       string `json:"status"`       // Typically "ENABLED"
}

// Sequence represents a behavior sequence
type Sequence struct {
	Type      string      `json:"@type"` // "com.amazon.alexa.behaviors.model.Sequence"
	StartNode interface{} `json:"startNode"`
}

// SerialNode represents a serial node in a sequence
type SerialNode struct {
	Type           string        `json:"@type"` // "com.amazon.alexa.behaviors.model.SerialNode"
	NodesToExecute []interface{} `json:"nodesToExecute"`
}

// ParallelNode represents a parallel node in a sequence
type ParallelNode struct {
	Type           string        `json:"@type"` // "com.amazon.alexa.behaviors.model.ParallelNode"
	NodesToExecute []interface{} `json:"nodesToExecute"`
}

// OpaquePayloadOperationNode represents an operation node with a payload
type OpaquePayloadOperationNode struct {
	Type             string                 `json:"@type"` // "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode"
	OperationType    string                 `json:"type"`  // e.g., "Alexa.Music.PlaySearchPhrase"
	OperationPayload map[string]interface{} `json:"operationPayload"`
}

// DeviceTarget represents a target device for operations
type DeviceTarget struct {
	DeviceType         string `json:"deviceType"`
	DeviceSerialNumber string `json:"deviceSerialNumber"`
}

// Target represents a target for operations (can include multiple devices)
type Target struct {
	CustomerID string         `json:"customerId,omitempty"`
	Devices    []DeviceTarget `json:"devices,omitempty"`
}
