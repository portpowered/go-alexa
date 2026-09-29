// Package alexamodels provides dependency API request and response models.
package alexamodels

import "github.com/portpowered/go-alexa/pkg/alexaapimodels"

// Device represents an Alexa device.
type Device struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Capabilities []CapabilityInterface  `json:"capabilities,omitempty"`
	State        map[string]interface{} `json:"state,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Capability represents a device capability.
type Capability struct {
	Type       string                 `json:"type"`
	Interface  string                 `json:"interface,omitempty"`
	Version    string                 `json:"version,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Command represents a device control command.
type Command struct {
	Type      string `json:"type"`
	DeviceID  string `json:"device_id,omitempty"`
	Payload   any    `json:"payload,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// AnnouncementContent represents content for an announcement.
type AnnouncementContent struct {
	Locale  string              `json:"locale"`
	Display AnnouncementDisplay `json:"display"`
	Speak   AnnouncementSpeak   `json:"speak"`
}

// AnnouncementDisplay represents display content.
type AnnouncementDisplay struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// AnnouncementSpeak represents speak content.
type AnnouncementSpeak struct {
	Type  string `json:"type"` // "text"
	Value string `json:"value"`
}

// MediaCommand represents a media control command.
type MediaCommand struct {
	Type    string `json:"type"`
	Shuffle *bool  `json:"shuffle,omitempty"` // For ShuffleCommand: true or false
	Repeat  *bool  `json:"repeat,omitempty"`  // For RepeatCommand: true or false
}

// Endpoint represents a device endpoint that can be used for behavior operations
// This interface provides unified access to endpoint information from different sources.
type Endpoint interface {
	GetDeviceType() string
	GetDeviceSerialNumber() string
	GetLocale() string
	GetEndpointId() string
	GetDeviceFamily() string
	GetDeviceAccountId() string
}

// Request structs for behavior APIs

// StopPlaybackRequest represents a request to stop playback.
type StopPlaybackRequest struct {
	Endpoint   Endpoint `json:"-"` // Not serialized, used to extract device info
	CustomerID string   `json:"customerId,omitempty"`
	AllDevices bool     `json:"allDevices,omitempty"`
}

// MediaControlRequest represents a request for media control (pause, resume, next, previous).
type MediaControlRequest struct {
	Endpoint Endpoint `json:"-"` // Not serialized, used to extract device info
}

// PlayerStateRequest represents a request to get player state.
type PlayerStateRequest struct {
	Endpoint Endpoint `json:"-"` // Not serialized, used to extract device info
}

// PlayerStateResponse represents the response from getting player state.
type PlayerStateResponse struct {
	PlayerInfo *PlayerInfo `json:"playerInfo,omitempty"`
}

// PlayerInfo represents player information.
type PlayerInfo struct {
	Hint             *string     `json:"hint,omitempty"`
	InfoText         *InfoText   `json:"infoText,omitempty"`
	IsPlayingInLemur bool        `json:"isPlayingInLemur,omitempty"`
	LemurVolume      *int        `json:"lemurVolume,omitempty"`
	Lyrics           *string     `json:"lyrics,omitempty"`
	MainArt          *Art        `json:"mainArt,omitempty"`
	MediaId          string      `json:"mediaId,omitempty"`
	MiniArt          *Art        `json:"miniArt,omitempty"`
	MiniInfoText     *InfoText   `json:"miniInfoText,omitempty"`
	PlaybackSource   *string     `json:"playbackSource,omitempty"`
	PlayingInLemurId *string     `json:"playingInLemurId,omitempty"`
	Progress         *Progress   `json:"progress,omitempty"`
	Provider         *Provider   `json:"provider,omitempty"`
	Quality          *string     `json:"quality,omitempty"`
	QueueId          string      `json:"queueId,omitempty"`
	State            string      `json:"state,omitempty"` // e.g., "PLAYING", "PAUSED", "STOPPED"
	Template         *Template   `json:"template,omitempty"`
	Transport        *Transport  `json:"transport,omitempty"`
	UpNextItems      interface{} `json:"upNextItems,omitempty"`
	Volume           *Volume     `json:"volume,omitempty"`
}

// InfoText represents text information displayed to the user.
type InfoText struct {
	Header         *string `json:"header,omitempty"`
	HeaderSubtext1 *string `json:"headerSubtext1,omitempty"`
	MultiLineMode  bool    `json:"multiLineMode,omitempty"`
	SubText1       string  `json:"subText1,omitempty"`
	SubText2       string  `json:"subText2,omitempty"`
	Title          string  `json:"title,omitempty"`
}

// Art represents artwork (album art, icons, etc.)
type Art struct {
	AltText     string  `json:"altText,omitempty"`
	ArtType     string  `json:"artType,omitempty"` // e.g., "UrlArtSource", "IconArtSource"
	ContentType *string `json:"contentType,omitempty"`
	URL         *string `json:"url,omitempty"`
	IconId      *string `json:"iconId,omitempty"`
	IconStyles  *string `json:"iconStyles,omitempty"`
}

// Progress represents playback progress information.
type Progress struct {
	AllowScrubbing bool    `json:"allowScrubbing,omitempty"`
	LocationInfo   *string `json:"locationInfo,omitempty"`
	MediaLength    int     `json:"mediaLength,omitempty"`   // Duration in seconds
	MediaProgress  int     `json:"mediaProgress,omitempty"` // Position in seconds
	ShowTiming     bool    `json:"showTiming,omitempty"`
	Visible        bool    `json:"visible,omitempty"`
}

// Provider represents the media provider information.
type Provider struct {
	ArtOverlay          *Art    `json:"artOverlay,omitempty"`
	FallbackMainArt     *Art    `json:"fallbackMainArt,omitempty"`
	ProviderDisplayName *string `json:"providerDisplayName,omitempty"`
	ProviderLogo        *Art    `json:"providerLogo,omitempty"`
	ProviderName        string  `json:"providerName,omitempty"`
}

// Template represents the display template.
type Template struct {
	Art                *Art   `json:"art,omitempty"`
	BackgroundImageURL string `json:"backgroundImageUrl,omitempty"`
	TemplateType       string `json:"templateType,omitempty"`
}

// Transport represents transport controls state.
type Transport struct {
	ClosedCaptions    *string            `json:"closedCaptions,omitempty"`
	LayoutType        string             `json:"layoutType,omitempty"`
	Lyrics            string             `json:"lyrics,omitempty"` // e.g., "ENABLED", "DISABLED"
	Next              string             `json:"next,omitempty"`   // e.g., "ENABLED", "DISABLED"
	PlayPause         string             `json:"playPause,omitempty"`
	Previous          string             `json:"previous,omitempty"`
	RateContentAction *RateContentAction `json:"rateContentAction,omitempty"`
	ThumbsDown        string             `json:"thumbsDown,omitempty"`
	ThumbsUp          string             `json:"thumbsUp,omitempty"`
}

// RateContentAction represents rating action information.
type RateContentAction struct {
	MediaOwnerCustomerId string  `json:"mediaOwnerCustomerId,omitempty"`
	Rating               *string `json:"rating,omitempty"`
	Type                 string  `json:"type,omitempty"`
}

// Volume represents volume information.
type Volume struct {
	Muted  bool `json:"muted,omitempty"`
	Volume int  `json:"volume,omitempty"` // Volume level (0-100)
}

// SendNotificationRequest represents a request to send a notification.
type SendNotificationRequest struct {
	Endpoint   Endpoint `json:"-"` // Not serialized, used to extract device info
	Message    string   `json:"message"`
	Title      string   `json:"title"`
	CustomerID string   `json:"customerId,omitempty"`
}

// SendAnnouncementRequest represents a request to send an announcement.
type SendAnnouncementRequest struct {
	Endpoint      Endpoint       `json:"-"` // Not serialized, used to extract device info
	Message       string         `json:"message"`
	Method        string         `json:"method"` // "speak", "show", or "all"
	Title         string         `json:"title,omitempty"`
	Locale        string         `json:"locale,omitempty"`
	CustomerID    string         `json:"customerId,omitempty"`
	TargetDevices []DeviceTarget `json:"targetDevices,omitempty"`
}

// SendTTSRequest represents a request to send text-to-speech.
type SendTTSRequest struct {
	Endpoint      Endpoint       `json:"-"` // Not serialized, used to extract device info
	Message       string         `json:"message"`
	CustomerID    string         `json:"customerId,omitempty"`
	TargetDevices []DeviceTarget `json:"targetDevices,omitempty"`
}

// PlayMusicRequest represents a request to play music.
type PlayMusicRequest struct {
	Endpoint     Endpoint `json:"-"` // Not serialized, used to extract device info
	ProviderID   string   `json:"providerId"`
	SearchPhrase string   `json:"searchPhrase"`
	CustomerID   string   `json:"customerId,omitempty"`
	TimerSeconds *int     `json:"timerSeconds,omitempty"`
}

// PlayVideoRequest represents a request to play video.
type PlayVideoRequest struct {
	Endpoint        Endpoint `json:"-"` // Not serialized, used to extract device info
	VideoProviderID string   `json:"videoProviderId"`
	SearchPhrase    string   `json:"searchPhrase"`
	CustomerID      string   `json:"customerId,omitempty"`
	TimerSeconds    *int     `json:"timerSeconds,omitempty"`
}

// PlayAudioURIRequest represents a request to play audio from a public HTTPS URI.
type PlayAudioURIRequest struct {
	Endpoint   Endpoint `json:"-"` // Not serialized, used to extract device info
	URI        string   `json:"uri"`
	CustomerID string   `json:"customerId,omitempty"`
}

// VideoProviderID identifies a video provider supported by Fire TV operations.
type VideoProviderID string

const (
	// VideoProviderIDAmazon identifies Prime Video.
	VideoProviderIDAmazon VideoProviderID = "PRIME VIDEO"
	// VideoProviderIDYouTube identifies YouTube.
	VideoProviderIDYouTube VideoProviderID = "YOUTUBE"
	// VideoProviderIDNetflix identifies Netflix.
	VideoProviderIDNetflix VideoProviderID = "NETFLIX"
	// VideoProviderIDDailymotion identifies Dailymotion.
	VideoProviderIDDailymotion VideoProviderID = "DAILYMOTION"
)

// FireTVRequest selects an endpoint for a Fire TV operation.
type FireTVRequest struct {
	Endpoint Endpoint `json:"-"` // Not serialized, used to extract device info
}

// FireTVOperationRequest represents a request for Fire TV operations.
type FireTVOperationRequest struct {
	Endpoint Endpoint `json:"-"` // Not serialized, used to extract device info
}

// PlaybackControlRequest represents a unified request for playback control
// Works for both media playback and FireTV playback.
type PlaybackControlRequest struct {
	Endpoint   Endpoint `json:"-"` // Not serialized, used to extract device info
	CustomerID string   `json:"customerId,omitempty"`
	Operation  string   `json:"operation"` // "play", "pause", "resume", "stop", "next", "previous"
}

// InterfaceMessageRequest represents a direct interface message to an endpoint feature/operation.
type InterfaceMessageRequest struct {
	Endpoint      Endpoint    `json:"-"`                 // Not serialized, used to extract endpoint information
	FeatureName   string      `json:"featureName"`       // Name of the feature, e.g. "playback"
	OperationName string      `json:"operationName"`     // Name of the operation on the feature, e.g. "play"
	Payload       interface{} `json:"payload,omitempty"` // Optional payload passed through to the API
}

// PowerControlRequest represents a unified request for power control
// Works for both regular power and FireTV power.
type PowerControlRequest struct {
	Endpoint Endpoint                  `json:"-"`     // Not serialized, used to extract device info
	State    alexaapimodels.PowerState `json:"state"` // "ON" or "OFF"
}

// VolumeControlRequest represents a unified request for volume control
// Works for both REST API volume control and GraphQL speaker feature control.
type VolumeControlRequest struct {
	Endpoint   Endpoint `json:"-"`                    // Not serialized, used to extract device info
	Volume     *int     `json:"volume,omitempty"`     // Set volume (0-100)
	Delta      *int     `json:"delta,omitempty"`      // Adjust volume by delta
	SetVolume  bool     `json:"setVolume,omitempty"`  // If true, use volume; if false, use delta
	CustomerID string   `json:"customerId,omitempty"` // For REST API
}

// Message represents a message received from the Alexa API via HTTP/2 connection.
type Message struct {
	Data map[string]interface{} `json:"data,omitempty"`
}
