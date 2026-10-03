// Package alexamodels provides dependency API request and response models.
package alexamodels

// Endpoint exposes identity details needed to build behavior operations.
type Endpoint interface {
	GetDeviceType() string
	GetDeviceSerialNumber() string
	GetLocale() string
	GetEndpointId() string
	GetDeviceFamily() string
	GetDeviceAccountId() string
}
