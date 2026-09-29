// Package alexamodels provides dependency API request and response models.
package alexamodels

import "time"

// Event represents an event received from Alexa.
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	DeviceID  string                 `json:"device_id,omitempty"`
}
