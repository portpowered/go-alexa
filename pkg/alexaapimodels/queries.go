package alexaapimodels

type EndpointQuery struct {
	IncludeFields *EndpointIncludeFields `json:"includeFields,omitempty"`
}

type EndpointIncludeFields struct {
	// Include the properties of the endpoints (power on, volume, etc) in the response object.
	Properties bool `json:"states,omitempty"`
	// Include the capabilities (metadata of what functions the endpoint supports) in the response object.
	Features bool `json:"capabilities,omitempty"`
}

// PlayerStateRequest represents a request to get player state
type PlayerStateRequest struct {
	// Target is the endpoint to get player state for
	Target EndpointInterface `json:"-"` // Not serialized, used to extract device info
}

// PlayerStateResponse represents the response from getting player state
type PlayerStateResponse struct {
	PlayerInfo *PlayerInfo `json:"playerInfo,omitempty"`
}

// PlayerInfo represents player information
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

// InfoText represents text information displayed to the user
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

// Progress represents playback progress information
type Progress struct {
	AllowScrubbing bool    `json:"allowScrubbing,omitempty"`
	LocationInfo   *string `json:"locationInfo,omitempty"`
	MediaLength    int     `json:"mediaLength,omitempty"`   // Duration in seconds
	MediaProgress  int     `json:"mediaProgress,omitempty"` // Position in seconds
	ShowTiming     bool    `json:"showTiming,omitempty"`
	Visible        bool    `json:"visible,omitempty"`
}

// Provider represents the media provider information
type Provider struct {
	ArtOverlay          *Art    `json:"artOverlay,omitempty"`
	FallbackMainArt     *Art    `json:"fallbackMainArt,omitempty"`
	ProviderDisplayName *string `json:"providerDisplayName,omitempty"`
	ProviderLogo        *Art    `json:"providerLogo,omitempty"`
	ProviderName        string  `json:"providerName,omitempty"`
}

// Template represents the display template
type Template struct {
	Art                *Art   `json:"art,omitempty"`
	BackgroundImageURL string `json:"backgroundImageUrl,omitempty"`
	TemplateType       string `json:"templateType,omitempty"`
}

// Transport represents transport controls state
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

// RateContentAction represents rating action information
type RateContentAction struct {
	MediaOwnerCustomerId string  `json:"mediaOwnerCustomerId,omitempty"`
	Rating               *string `json:"rating,omitempty"`
	Type                 string  `json:"type,omitempty"`
}

// Volume represents volume information
type Volume struct {
	Muted  bool `json:"muted,omitempty"`
	Volume int  `json:"volume,omitempty"` // Volume level (0-100)
}
