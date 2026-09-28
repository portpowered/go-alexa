package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/rest/internal/wire"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

// Notes: the APIs for this set of endpoints are only available to certain for device registrations of a certain set of device types,
// this list of device types is unknown.

// ListEndpointsOptions contains options for listing endpoints
type ListEndpointsOptions struct {
	Owner      string   // Filter by owner (typically "~caller")
	Expand     []string // Expand features (e.g., "all", "feature:brightness")
	MaxResults int      // Maximum number of results (1-50, default: 10)
	NextToken  string   // Token for pagination
}

// GetEndpoints retrieves all endpoints for the authenticated user (owner)
// This is the primary enumeration API for getting endpoints by owner.
// GET /v2/endpoints?owner={owner}&expand={expand}&maxResults={maxResults}&nextToken={nextToken}
func (c *Client) GetEndpoints(ctx context.Context, opts *ListEndpointsOptions) (*alexamodels.EndpointListResponse, error) {
	if opts == nil {
		opts = &ListEndpointsOptions{Owner: alexamodels.DefaultOwnerCaller}
	}
	if opts.Owner == "" {
		opts.Owner = alexamodels.DefaultOwnerCaller
	}

	path := buildEndpointListPath(alexamodels.APIPathV2Endpoints, opts)
	var response wire.WireEndpointListResponse
	if err := c.doJSONRequest(ctx, "GET", path, nil, &response); err != nil {
		return nil, err
	}
	return convertWireModel[alexamodels.EndpointListResponse](response)
}

// GetEndpointByID retrieves a specific endpoint by its ID
// GET /v2/endpoints/{endpointId}?expand={expand}
func (c *Client) GetEndpointByID(ctx context.Context, endpointID string, expand []string) (*alexamodels.Device, error) {
	path := fmt.Sprintf("%s/%s", alexamodels.APIPathV2Endpoints, endpointID)
	if len(expand) > 0 {
		params := url.Values{}
		for _, e := range expand {
			params.Add(alexamodels.QueryParamExpand, e)
		}
		path += "?" + params.Encode()
	}

	var endpoint wire.WireDevice
	if err := c.doJSONRequest(ctx, "GET", path, nil, &endpoint); err != nil {
		return nil, err
	}
	return convertWireModel[alexamodels.Device](endpoint)
}

// QueryEndpoints performs an advanced search query for endpoints
// POST /v2/endpoint-query
func (c *Client) QueryEndpoints(ctx context.Context, query *alexamodels.EndpointQueryRequest, opts *ListEndpointsOptions) (*alexamodels.EndpointListResponse, error) {
	if query == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "query cannot be nil",
		}
	}

	path := alexamodels.APIPathV2EndpointQuery
	if opts != nil {
		// Add query parameters for pagination
		params := url.Values{}
		if opts.NextToken != "" {
			params.Set(alexamodels.QueryParamNextToken, opts.NextToken)
		}
		if opts.MaxResults > 0 {
			params.Set(alexamodels.QueryParamMaxResults, strconv.Itoa(opts.MaxResults))
		}
		if len(opts.Expand) > 0 {
			for _, e := range opts.Expand {
				params.Add(alexamodels.QueryParamExpand, e)
			}
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	wireQuery, err := convertWireModel[wire.WireEndpointQueryRequest](query)
	if err != nil {
		return nil, &alexaapimodels.BadRequestError{Message: "failed to convert endpoint query", Err: err}
	}
	var response wire.WireEndpointListResponse
	if err := c.doJSONRequest(ctx, "POST", path, wireQuery, &response); err != nil {
		return nil, err
	}
	return convertWireModel[alexamodels.EndpointListResponse](response)
}

// ForgetEndpoint requests that Alexa forget the specified endpoint
// POST /v2/endpoints/{endpointId}/forget
func (c *Client) ForgetEndpoint(ctx context.Context, endpointID string) error {
	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsForget, endpointID)
	if err := c.doJSONRequest(ctx, "POST", path, nil, nil); err != nil {
		return err
	}
	return nil
}

// DeregisterEndpoint deregisters the specified endpoint from your Amazon Business account
// POST /v2/endpoints/{endpointId}/deregister
func (c *Client) DeregisterEndpoint(ctx context.Context, endpointID string) error {
	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsDeregister, endpointID)
	if err := c.doJSONRequest(ctx, "POST", path, nil, nil); err != nil {
		return err
	}
	return nil
}

// UpdateFriendlyName changes the friendly name for the specified endpoint
// POST /v2/endpoints/{endpointId}/friendlyName
func (c *Client) UpdateFriendlyName(ctx context.Context, endpointID string, friendlyName string) error {
	request := wire.WireFriendlyNameRequest{
		FriendlyName: wire.WireFriendlyNameValue{
			Type: alexamodels.DefaultFriendlyNameType,
			Value: wire.WireFriendlyNameText{
				Text: friendlyName,
			},
		},
	}

	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsFriendlyName, endpointID)
	if err := c.doJSONRequest(ctx, "POST", path, request, nil); err != nil {
		return err
	}
	return nil
}

// buildEndpointListPath builds a query string for endpoint list requests
func buildEndpointListPath(basePath string, opts *ListEndpointsOptions) string {
	if opts == nil {
		return basePath
	}

	params := url.Values{}
	if opts.Owner != "" {
		params.Set(alexamodels.QueryParamOwner, opts.Owner)
	}
	if len(opts.Expand) > 0 {
		for _, e := range opts.Expand {
			params.Add(alexamodels.QueryParamExpand, e)
		}
	}
	if opts.MaxResults > 0 {
		params.Set(alexamodels.QueryParamMaxResults, strconv.Itoa(opts.MaxResults))
	}
	if opts.NextToken != "" {
		params.Set(alexamodels.QueryParamNextToken, opts.NextToken)
	}

	if len(params) > 0 {
		return basePath + "?" + params.Encode()
	}
	return basePath
}

// Legacy methods for backward compatibility

// ControlEndpoint sends a control command to an endpoint
// Note: This is not part of the enumeration APIs but kept for backward compatibility
func (c *Client) ControlEndpoint(ctx context.Context, endpointID string, command alexamodels.Command) error {
	command.DeviceID = endpointID
	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsControl, endpointID)
	wireCommand, err := convertWireModel[wire.WireCommand](command)
	if err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to convert control command", Err: err}
	}
	if err := c.doJSONRequest(ctx, "POST", path, wireCommand, nil); err != nil {
		return err
	}
	return nil
}

// SendInterfaceMessage sends a direct interface message to an endpoint feature operation.
// This uses the REST interface message API:
// POST /v2/endpoints/{endpointId}/interfaces/{feature}/{operation}/
// The trailing slash is required by the service router.
func (c *Client) SendInterfaceMessage(ctx context.Context, req *alexamodels.InterfaceMessageRequest) error {
	if req == nil {
		return &alexaapimodels.BadRequestError{
			Message: "request is required",
		}
	}

	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	if req.FeatureName == "" {
		return &alexaapimodels.BadRequestError{
			Message: "feature name is required",
		}
	}

	if req.OperationName == "" {
		return &alexaapimodels.BadRequestError{
			Message: "operation name is required",
		}
	}

	endpointID := req.Endpoint.GetEndpointId()
	if endpointID == "" {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint ID is required",
		}
	}

	path := fmt.Sprintf(alexamodels.APIPathV2EndpointInterface, endpointID, req.FeatureName, req.OperationName)

	// Service expects a JSON body; use an empty object when no payload is provided.
	body := req.Payload
	if body == nil {
		body = map[string]interface{}{}
	}

	if err := c.doJSONRequest(ctx, http.MethodPost, path, body, nil); err != nil {
		return err
	}

	return nil
}

// GetDevicesV2Options contains options for getting devices via the devices-v2 API
type GetDevicesV2Options struct {
	CSRFToken string // CSRF token for cookie header
}

// GetDevicesV2 retrieves all devices using the devices-v2 API
// GET {alexaAmazonBaseUri}/api/devices-v2/device
// This API is used to enumerate 1P amazon devices, like the echo show, etc.
func (c *Client) GetDevicesV2(ctx context.Context, opts *GetDevicesV2Options) (*alexamodels.DevicesV2Response, error) {
	// Build the full URL using the client's base URI
	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathDevicesV2Device

	// Build custom headers
	customHeaders := make(map[string]string)
	if opts != nil && opts.CSRFToken != "" {
		customHeaders[alexamodels.CookieHeaderName] = fmt.Sprintf(alexamodels.CookieCSRFFormat, opts.CSRFToken)
	}

	// Use the client's helper method to perform the request
	var response wire.WireDevicesV2Response
	if err := c.doJSONRequestWithFullURL(ctx, "GET", baseURL, nil, customHeaders, &response, false); err != nil {
		return nil, err
	}

	return convertWireModel[alexamodels.DevicesV2Response](response)
}

// Behavior API functions

// RunBehavior runs a behavior sequence using the behaviors preview API
// POST {alexaAmazonBaseUri}/api/behaviors/preview
// This is the base function for creating behavior operations
func (c *Client) RunBehavior(ctx context.Context, sequenceJSON string) error {
	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathBehaviorsPreview

	request := wire.WireBehaviorPreviewRequest{
		BehaviorId:   alexamodels.DefaultBehaviorID,
		SequenceJson: sequenceJSON,
		Status:       alexamodels.DefaultBehaviorStatus,
	}

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, request, nil, nil, true); err != nil {
		return err
	}
	return nil
}

// SendSequence sends a sequence command to a device
// This is a helper that builds the operation node and calls RunBehavior
// The operationPayload should already contain all required fields including deviceType, deviceSerialNumber, locale, and customerId
func (c *Client) SendSequence(ctx context.Context, sequenceType string, operationPayload map[string]interface{}) error {
	node := alexamodels.OpaquePayloadOperationNode{
		Type:             alexamodels.ModelTypeOpaquePayloadOperationNode,
		OperationType:    sequenceType,
		OperationPayload: operationPayload,
	}

	sequence := alexamodels.Sequence{
		Type:      alexamodels.ModelTypeSequence,
		StartNode: node,
	}

	sequenceJSON, err := json.Marshal(sequence)
	if err != nil {
		return &alexaapimodels.BadRequestError{
			Message: "failed to marshal sequence",
			Err:     err,
		}
	}

	return c.RunBehavior(ctx, string(sequenceJSON))
}

// Media control functions

// StopPlayback stops playback on a device
// Uses Alexa.DeviceControls.Stop sequence
func (c *Client) StopPlayback(ctx context.Context, req *alexamodels.StopPlaybackRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:         deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber: deviceSerialNumber,
		alexamodels.PayloadKeyLocale:             alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:         customerID,
		alexamodels.PayloadKeySkillID:            alexamodels.SkillIDAlexaDeviceControls,
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeDeviceControlsStop, payload)
}

// SetVolume sets the volume on a device
// Uses Alexa.DeviceControls.Volume sequence
func (c *Client) SetVolume(ctx context.Context, req *alexamodels.VolumeControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	// Validate volume is between 0 and 100
	if req.Volume != nil && (*req.Volume < 0 || *req.Volume > 100) {
		return &alexaapimodels.BadRequestError{
			Message: "volume needs to be between 0 and 100",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:         deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber: deviceSerialNumber,
		alexamodels.PayloadKeyLocale:             alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:         customerID,
		alexamodels.PayloadKeyValue:              req.Volume,
		alexamodels.PayloadKeySkillID:            alexamodels.SkillIDAlexaDeviceControls,
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeDeviceControlsVolume, payload)
}

// PausePlayback pauses playback on a device
// Uses /api/np/command endpoint
func (c *Client) PausePlayback(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type: alexamodels.MediaCommandTypePause,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to pause playback",
			Err:     err,
		}
	}
	return nil
}

// ResumePlayback resumes playback on a device
// Uses /api/np/command endpoint
func (c *Client) ResumePlayback(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type: alexamodels.MediaCommandTypePlay,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to resume playback",
			Err:     err,
		}
	}
	return nil
}

// NextTrack skips to the next track
// Uses /api/np/command endpoint
func (c *Client) NextTrack(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type: alexamodels.MediaCommandTypeNext,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to skip to next track",
			Err:     err,
		}
	}
	return nil
}

// PreviousTrack goes to the previous track
// Uses /api/np/command endpoint
func (c *Client) PreviousTrack(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type: alexamodels.MediaCommandTypePrevious,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to go to previous track",
			Err:     err,
		}
	}
	return nil
}

// GetPlayerState retrieves the current player state for a device
// Uses /api/np/player endpoint
func (c *Client) GetPlayerState(ctx context.Context, req *alexamodels.PlayerStateRequest) (*alexamodels.PlayerStateResponse, error) {
	if req.Endpoint == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPPlayer

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	var response wire.WirePlayerStateResponse
	if err := c.doJSONRequestWithFullURL(ctx, "GET", baseURL, nil, nil, &response, true); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to get player state",
			Err:     err,
		}
	}
	return convertWireModel[alexamodels.PlayerStateResponse](response)
}

// ForwardMedia fast-forwards the current media
// Uses /api/np/command endpoint
func (c *Client) ForwardMedia(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type: alexamodels.MediaCommandTypeForward,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to forward media",
			Err:     err,
		}
	}
	return nil
}

// RewindMedia rewinds the current media
// Uses /api/np/command endpoint
func (c *Client) RewindMedia(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type: alexamodels.MediaCommandTypeRewind,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to rewind media",
			Err:     err,
		}
	}
	return nil
}

// SetShuffle sets the shuffle state for media playback
// Uses /api/np/command endpoint
func (c *Client) SetShuffle(ctx context.Context, req *alexamodels.MediaControlRequest, shuffle bool) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type:    alexamodels.MediaCommandTypeShuffle,
		Shuffle: &shuffle,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to set shuffle",
			Err:     err,
		}
	}
	return nil
}

// SetRepeat sets the repeat state for media playback
// Uses /api/np/command endpoint
func (c *Client) SetRepeat(ctx context.Context, req *alexamodels.MediaControlRequest, repeat bool) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathNPCommand

	command := alexamodels.MediaCommand{
		Type:   alexamodels.MediaCommandTypeRepeat,
		Repeat: &repeat,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, command, nil, nil, true); err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to set repeat",
			Err:     err,
		}
	}
	return nil
}

// SendNotification sends a mobile push notification
// Uses Alexa.Notifications.SendMobilePush sequence
func (c *Client) SendNotification(ctx context.Context, req *alexamodels.SendNotificationRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:          deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber:  deviceSerialNumber,
		alexamodels.PayloadKeyLocale:              alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:          customerID,
		alexamodels.PayloadKeyNotificationMessage: req.Message,
		alexamodels.PayloadKeyAlexaURL:            alexamodels.AlexaURLBehaviors,
		alexamodels.PayloadKeyTitle:               req.Title,
		alexamodels.PayloadKeySkillID:             alexamodels.SkillIDRoutinesMessaging,
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeNotificationsSendMobilePush, payload)
}

// SendAnnouncement sends an announcement to Alexa devices
// Uses AlexaAnnouncement sequence
func (c *Client) SendAnnouncement(ctx context.Context, req *alexamodels.SendAnnouncementRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	locale := req.Locale
	if locale == "" {
		locale = req.Endpoint.GetLocale()
	}
	if locale == "" {
		locale = alexamodels.DefaultLocale
	}

	// Build display and speak content based on method
	var display alexamodels.AnnouncementDisplay
	speak := alexamodels.AnnouncementSpeak{Type: alexamodels.DefaultAnnouncementSpeakType}

	switch req.Method {
	case alexamodels.AnnouncementMethodSpeak:
		display = alexamodels.AnnouncementDisplay{Title: "", Body: ""}
		speak.Value = req.Message
	case alexamodels.AnnouncementMethodShow:
		display = alexamodels.AnnouncementDisplay{Title: req.Title, Body: req.Message}
		speak.Value = ""
	default: // "all"
		display = alexamodels.AnnouncementDisplay{Title: req.Title, Body: req.Message}
		speak.Value = req.Message
	}

	content := []alexamodels.AnnouncementContent{
		{
			Locale:  locale,
			Display: display,
			Speak:   speak,
		},
	}

	// Build target
	target := alexamodels.Target{
		CustomerID: c.customerID,
		Devices: []alexamodels.DeviceTarget{
			{
				DeviceType:         req.Endpoint.GetDeviceType(),
				DeviceSerialNumber: req.Endpoint.GetDeviceSerialNumber(),
			},
		},
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:         deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber: deviceSerialNumber,
		alexamodels.PayloadKeyLocale:             locale,
		alexamodels.PayloadKeyCustomerID:         customerID,
		alexamodels.PayloadKeyExpireAfter:        alexamodels.DefaultAnnouncementExpireAfter,
		alexamodels.PayloadKeyContent:            content,
		alexamodels.PayloadKeyTarget:             target,
		alexamodels.PayloadKeySkillID:            alexamodels.SkillIDAlexaNotifications,
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeAnnouncement, payload)
}

// SendTTS sends a text-to-speech message
// Uses Alexa.Speak or Alexa.CannedTts.Speak sequence
func (c *Client) SendTTS(ctx context.Context, req *alexamodels.SendTTSRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	// // Check if it's a canned TTS message
	// if len(req.Message) > 20 && req.Message[:20] == "alexa.cannedtts.speak" {
	// 	customerID := req.CustomerID
	// 	if c.customerID != "" {
	// 		customerID = c.customerID
	// 	}
	// 	payload := map[string]interface{}{
	// 		alexamodels.PayloadKeyDeviceType:         req.Endpoint.GetDeviceType(),
	// 		alexamodels.PayloadKeyDeviceSerialNumber: req.Endpoint.GetDeviceSerialNumber(),
	// 		alexamodels.PayloadKeyLocale:             alexamodels.DefaultLocale,
	// 		alexamodels.PayloadKeyCustomerID:         customerID,
	// 		alexamodels.PayloadKeyCannedTtsStringID: req.Message,
	// 		alexamodels.PayloadKeySkillID:         alexamodels.SkillIDSaySomething,
	// 	}
	// 	return c.SendSequence(ctx, alexamodels.OperationTypeCannedTtsSpeak, payload)
	// }

	// Regular TTS
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:         req.Endpoint.GetDeviceType(),
		alexamodels.PayloadKeyDeviceSerialNumber: req.Endpoint.GetDeviceSerialNumber(),
		alexamodels.PayloadKeyLocale:             alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:         customerID,
		alexamodels.PayloadKeyTextToSpeak:        req.Message,
		alexamodels.PayloadKeySkillID:            alexamodels.SkillIDSaySomething,
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeSpeak, payload)
}

// PlayMusic plays music based on a search query
// Uses Alexa.Music.PlaySearchPhrase sequence
func (c *Client) PlayMusic(ctx context.Context, req *alexamodels.PlayMusicRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:            deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber:    deviceSerialNumber,
		alexamodels.PayloadKeyLocale:                alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:            customerID,
		alexamodels.PayloadKeySearchPhrase:          req.SearchPhrase,
		alexamodels.PayloadKeySanitizedSearchPhrase: req.SearchPhrase,
		alexamodels.PayloadKeyMusicProviderID:       req.ProviderID,
	}

	if req.TimerSeconds != nil && *req.TimerSeconds > 0 {
		payload[alexamodels.PayloadKeyWaitTimeInSeconds] = *req.TimerSeconds
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeMusicPlaySearchPhrase, payload)
}

// PlayAudioURI plays audio from a public HTTPS URI on a device
// Uses Alexa.Sound behavior sequence with soundStringId set to the audio URL
func (c *Client) PlayAudioURI(ctx context.Context, req *alexamodels.PlayAudioURIRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:         deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber: deviceSerialNumber,
		alexamodels.PayloadKeyLocale:             alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:         customerID,
		alexamodels.PayloadKeySoundStringID:      req.URI,
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeSound, payload)
}

// PlayVideo plays video based on a search query
// Uses Alexa.Operation.Video.PlaySearchPhrase sequence
func (c *Client) PlayVideo(ctx context.Context, req *alexamodels.PlayVideoRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()
	customerID := req.CustomerID
	if c.customerID != "" {
		customerID = c.customerID
	}

	// Combine search phrase with provider: "{searchPhrase} on {provider}"
	combinedSearchPhrase := req.SearchPhrase
	if req.VideoProviderID != "" {
		combinedSearchPhrase = fmt.Sprintf("%s on %s", req.SearchPhrase, req.VideoProviderID)
	}

	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceType:            deviceType,
		alexamodels.PayloadKeyDeviceSerialNumber:    deviceSerialNumber,
		alexamodels.PayloadKeyLocale:                alexamodels.DefaultLocale,
		alexamodels.PayloadKeyCustomerID:            customerID,
		alexamodels.PayloadKeySearchPhrase:          combinedSearchPhrase,
		alexamodels.PayloadKeySanitizedSearchPhrase: combinedSearchPhrase,
	}

	if req.TimerSeconds != nil && *req.TimerSeconds > 0 {
		payload[alexamodels.PayloadKeyWaitTimeInSeconds] = *req.TimerSeconds
	}

	return c.SendSequence(ctx, alexamodels.OperationTypeVideoPlaySearchPhrase, payload)
}

// Fire TV operation functions

// SendFireTVSequence sends a Fire TV sequence command
// This is a helper that builds the operation node and calls RunBehavior for Fire TV operations
func (c *Client) SendFireTVSequence(ctx context.Context, deviceAccountID, operationType string) error {
	// Build the operation payload with only deviceAccountId and skillId
	payload := map[string]interface{}{
		alexamodels.PayloadKeyDeviceAccountID: deviceAccountID,
		alexamodels.PayloadKeySkillID:         alexamodels.SkillIDRoutinesFireTV,
	}

	node := alexamodels.OpaquePayloadOperationNode{
		Type:             alexamodels.ModelTypeOpaquePayloadOperationNode,
		OperationType:    operationType,
		OperationPayload: payload,
	}

	sequence := alexamodels.Sequence{
		Type:      alexamodels.ModelTypeSequence,
		StartNode: node,
	}

	sequenceJSON, err := json.Marshal(sequence)
	if err != nil {
		return &alexaapimodels.BadRequestError{
			Message: "failed to marshal sequence",
			Err:     err,
		}
	}

	baseURL := c.alexaAmazonBaseUri + alexamodels.APIPathBehaviorsPreview

	request := alexamodels.BehaviorPreviewRequest{
		BehaviorID:   alexamodels.DefaultBehaviorID,
		SequenceJSON: string(sequenceJSON),
		Status:       alexamodels.DefaultBehaviorStatus,
	}

	if err := c.doJSONRequestWithFullURL(ctx, "POST", baseURL, request, nil, nil, true); err != nil {
		return err
	}
	return nil
}

// FireTVTurnOn turns on a Fire TV device
// Uses Alexa.Operation.FireTV.TurnOn sequence
func (c *Client) FireTVTurnOn(ctx context.Context, req *alexamodels.FireTVRequest) error {
	deviceAccountID := req.Endpoint.GetDeviceAccountId()
	if deviceAccountID == "" && req.Endpoint != nil {
		deviceAccountID = req.Endpoint.GetDeviceAccountId()
	}

	if deviceAccountID == "" {
		return &alexaapimodels.BadRequestError{
			Message: "deviceAccountId is required",
		}
	}

	if req.Endpoint != nil {
		deviceFamily := req.Endpoint.GetDeviceFamily()
		if deviceFamily != alexamodels.DeviceFamilyFireTV {
			return &alexaapimodels.BadRequestError{
				Message: "device is not a Fire TV",
			}
		}
	}

	return c.SendFireTVSequence(ctx, deviceAccountID, alexamodels.OperationTypeFireTVTurnOn)
}

// FireTVTurnOff turns off a Fire TV device
// Uses Alexa.Operation.FireTV.TurnOff sequence
func (c *Client) FireTVTurnOff(ctx context.Context, req *alexamodels.FireTVRequest) error {
	deviceAccountID := req.Endpoint.GetDeviceAccountId()
	if deviceAccountID == "" && req.Endpoint != nil {
		deviceAccountID = req.Endpoint.GetDeviceAccountId()
	}

	if deviceAccountID == "" {
		return &alexaapimodels.BadRequestError{
			Message: "deviceAccountId is required",
		}
	}

	if req.Endpoint != nil {
		deviceFamily := req.Endpoint.GetDeviceFamily()
		if deviceFamily != alexamodels.DeviceFamilyFireTV {
			return &alexaapimodels.BadRequestError{
				Message: "device is not a Fire TV",
			}
		}
	}

	return c.SendFireTVSequence(ctx, deviceAccountID, alexamodels.OperationTypeFireTVTurnOff)
}

// FireTVTurnOnOff turns on or off a Fire TV device based on the value parameter
// If value is true, turns on; if false, turns off
// Uses Alexa.Operation.FireTV.TurnOn or Alexa.Operation.FireTV.TurnOff sequence
func (c *Client) FireTVTurnOnOff(ctx context.Context, req *alexamodels.FireTVRequest, value bool) error {
	if value {
		return c.FireTVTurnOn(ctx, req)
	}
	return c.FireTVTurnOff(ctx, req)
}

// FireTVPauseVideo pauses video playback on a Fire TV device
// Uses Alexa.Operation.FireTV.PauseVideo sequence
func (c *Client) FireTVPauseVideo(ctx context.Context, req *alexamodels.FireTVRequest) error {
	deviceAccountID := req.Endpoint.GetDeviceAccountId()
	if deviceAccountID == "" && req.Endpoint != nil {
		deviceAccountID = req.Endpoint.GetDeviceAccountId()
	}

	if deviceAccountID == "" {
		return &alexaapimodels.BadRequestError{
			Message: "deviceAccountId is required",
		}
	}

	if req.Endpoint != nil {
		deviceFamily := req.Endpoint.GetDeviceFamily()
		if deviceFamily != alexamodels.DeviceFamilyFireTV {
			return &alexaapimodels.BadRequestError{
				Message: "device is not a Fire TV",
			}
		}
	}

	return c.SendFireTVSequence(ctx, deviceAccountID, alexamodels.OperationTypeFireTVPauseVideo)
}

// FireTVResumeVideo resumes video playback on a Fire TV device
// Uses Alexa.Operation.FireTV.ResumeVideo sequence
func (c *Client) FireTVResumeVideo(ctx context.Context, req *alexamodels.FireTVRequest) error {
	deviceAccountID := req.Endpoint.GetDeviceAccountId()
	if deviceAccountID == "" && req.Endpoint != nil {
		deviceAccountID = req.Endpoint.GetDeviceAccountId()
	}

	if deviceAccountID == "" {
		return &alexaapimodels.BadRequestError{
			Message: "deviceAccountId is required",
		}
	}

	if req.Endpoint != nil {
		deviceFamily := req.Endpoint.GetDeviceFamily()
		if deviceFamily != alexamodels.DeviceFamilyFireTV {
			return &alexaapimodels.BadRequestError{
				Message: "device is not a Fire TV",
			}
		}
	}

	return c.SendFireTVSequence(ctx, deviceAccountID, alexamodels.OperationTypeFireTVResumeVideo)
}

// FireTVNavigateHome navigates to the home screen on a Fire TV device
// Uses Alexa.Operation.FireTV.NavigateHome sequence
func (c *Client) FireTVNavigateHome(ctx context.Context, req *alexamodels.FireTVRequest) error {
	deviceAccountID := req.Endpoint.GetDeviceAccountId()
	if deviceAccountID == "" && req.Endpoint != nil {
		deviceAccountID = req.Endpoint.GetDeviceAccountId()
	}

	if deviceAccountID == "" {
		return &alexaapimodels.BadRequestError{
			Message: "deviceAccountId is required",
		}
	}

	if req.Endpoint != nil {
		deviceFamily := req.Endpoint.GetDeviceFamily()
		if deviceFamily != alexamodels.DeviceFamilyFireTV {
			return &alexaapimodels.BadRequestError{
				Message: "device is not a Fire TV",
			}
		}
	}

	return c.SendFireTVSequence(ctx, deviceAccountID, alexamodels.OperationTypeFireTVNavigateHome)
}
