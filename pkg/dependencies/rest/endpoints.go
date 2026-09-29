package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/internal/wire"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

const maxBehaviorNodeDepth = 16

// Notes: the APIs for this set of endpoints are only available to certain for device registrations of a certain set of device types,
// this list of device types is unknown.

// ListEndpointsOptions contains options for listing endpoints.
type ListEndpointsOptions struct {
	Owner      string   // Filter by owner (typically "~caller")
	Expand     []string // Expand features (e.g., "all", "feature:brightness")
	MaxResults int      // Maximum number of results (1-50, default: 10)
	NextToken  string   // Token for pagination
}

// GetEndpoints retrieves all endpoints for the authenticated user (owner)
// This is the primary enumeration API for getting endpoints by owner.
// GET /v2/endpoints?owner={owner}&expand={expand}&maxResults={maxResults}&nextToken={nextToken}.
func (c *Client) GetEndpoints(ctx context.Context, opts *ListEndpointsOptions) (*alexamodels.EndpointListResponse, error) {
	if opts == nil {
		opts = &ListEndpointsOptions{Owner: alexamodels.DefaultOwnerCaller}
	}

	if opts.Owner == "" {
		opts.Owner = alexamodels.DefaultOwnerCaller
	}

	path := buildEndpointListPath(alexamodels.APIPathV2Endpoints, opts)

	var response wire.WireEndpointListResponse

	err := c.doJSONRequest(ctx, apiroutes.MethodListRestEndpoints, path, nil, &response)
	if err != nil {
		return nil, err
	}

	return convertWireModel[alexamodels.EndpointListResponse](response)
}

// GetEndpointByID retrieves a specific endpoint by its ID
// GET /v2/endpoints/{endpointId}?expand={expand}.
func (c *Client) GetEndpointByID(ctx context.Context, endpointID string, expand []string) (*alexamodels.Device, error) {
	path := fmt.Sprintf(apiroutes.PathGetRestEndpoint, endpointID)

	if len(expand) > 0 {
		params := url.Values{}
		for _, e := range expand {
			params.Add(alexamodels.QueryParamExpand, e)
		}

		path += "?" + params.Encode()
	}

	var endpoint wire.WireDevice

	err := c.doJSONRequest(ctx, apiroutes.MethodGetRestEndpoint, path, nil, &endpoint)
	if err != nil {
		return nil, err
	}

	return convertWireModel[alexamodels.Device](endpoint)
}

// QueryEndpoints performs an advanced search query for endpoints
// POST /v2/endpoint-query.
func (c *Client) QueryEndpoints(
	ctx context.Context,
	query *alexamodels.EndpointQueryRequest,
	opts *ListEndpointsOptions,
) (*alexamodels.EndpointListResponse, error) {
	if query == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "query cannot be nil",
		}
	}

	path := apiroutes.PathQueryRestEndpoints

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
	{
		err := c.doJSONRequest(ctx, apiroutes.MethodQueryRestEndpoints, path, wireQuery, &response)
		if err != nil {
			return nil, err
		}
	}

	return convertWireModel[alexamodels.EndpointListResponse](response)
}

// ForgetEndpoint requests that Alexa forget the specified endpoint
// POST /v2/endpoints/{endpointId}/forget.
func (c *Client) ForgetEndpoint(ctx context.Context, endpointID string) error {
	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsForget, endpointID)

	err := c.doJSONRequest(ctx, apiroutes.MethodForgetEndpoint, path, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// DeregisterEndpoint deregisters the specified endpoint from your Amazon Business account
// POST /v2/endpoints/{endpointId}/deregister.
func (c *Client) DeregisterEndpoint(ctx context.Context, endpointID string) error {
	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsDeregister, endpointID)

	err := c.doJSONRequest(ctx, apiroutes.MethodDeregisterEndpoint, path, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// UpdateFriendlyName changes the friendly name for the specified endpoint
// POST /v2/endpoints/{endpointId}/friendlyName.
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

	err := c.doJSONRequest(ctx, apiroutes.MethodUpdateEndpointFriendlyName, path, request, nil)
	if err != nil {
		return err
	}

	return nil
}

// buildEndpointListPath builds a query string for endpoint list requests.
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
// Note: This is not part of the enumeration APIs but kept for backward compatibility.
func (c *Client) ControlEndpoint(ctx context.Context, endpointID string, command alexamodels.Command) error {
	command.DeviceID = endpointID
	path := fmt.Sprintf(alexamodels.APIPathV2EndpointsControl, endpointID)

	wireCommand, err := convertWireModel[wire.WireCommand](command)
	if err != nil {
		return &alexaapimodels.BadRequestError{Message: "failed to convert control command", Err: err}
	}

	{
		err := c.doJSONRequest(ctx, apiroutes.MethodControlRestEndpoint, path, wireCommand, nil)
		if err != nil {
			return err
		}
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
	if payload, ok := body.(map[string]interface{}); ok {
		body = wire.WireInterfacePayload(payload)
	}

	if body == nil {
		body = wire.WireInterfacePayload{}
	}

	err := c.doJSONRequest(ctx, apiroutes.MethodSendEndpointInterfaceMessage, path, body, nil)
	if err != nil {
		return err
	}

	return nil
}

// GetDevicesV2Options contains options for getting devices via the devices-v2 API.
type GetDevicesV2Options struct {
	CSRFToken string // CSRF token for cookie header
}

// GetDevicesV2 retrieves all devices using the devices-v2 API
// GET {alexaAmazonBaseURI}/api/devices-v2/device
// This API is used to enumerate 1P amazon devices, like the echo show, etc.
func (c *Client) GetDevicesV2(ctx context.Context, opts *GetDevicesV2Options) (*alexamodels.DevicesV2Response, error) {
	// Build the full URL using the client's base URI
	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathDevicesV2Device

	// Build custom headers
	customHeaders := make(map[string]string)
	if opts != nil && opts.CSRFToken != "" {
		customHeaders[apiroutes.HeaderCookie] = fmt.Sprintf(alexamodels.CookieCSRFFormat, opts.CSRFToken)
	}

	// Use the client's helper method to perform the request
	var response wire.WireDevicesV2Response

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodListFirstPartyDevices, baseURL, nil, customHeaders, &response, false)
	if err != nil {
		return nil, err
	}

	return convertWireModel[alexamodels.DevicesV2Response](response)
}

// Behavior API functions

// RunBehavior runs a behavior sequence using the behaviors preview API
// POST {alexaAmazonBaseURI}/api/behaviors/preview
// This is the base function for creating behavior operations.
func (c *Client) RunBehavior(ctx context.Context, sequenceJSON string) error {
	err := validateBehaviorSequence(sequenceJSON)
	if err != nil {
		return &alexaapimodels.BadRequestError{Message: "sequence is not schema-backed", Err: err}
	}

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathBehaviorsPreview

	request := wire.WireBehaviorPreviewRequest{
		BehaviorId:   alexamodels.DefaultBehaviorID,
		SequenceJson: sequenceJSON,
		Status:       alexamodels.DefaultBehaviorStatus,
	}

	err = c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSubmitBehaviorPreview, baseURL, request, nil, nil, true)
	if err != nil {
		return err
	}

	return nil
}

func validateBehaviorSequence(sequenceJSON string) error {
	var sequence alexamodels.Sequence

	decoder := json.NewDecoder(strings.NewReader(sequenceJSON))
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&sequence)
	if err != nil {
		return fmt.Errorf("decode behavior sequence: %w", err)
	}

	err = decoder.Decode(new(any))

	if err != io.EOF {
		return errTrailingBehaviorJSON
	}

	if !sequence.Type.Valid() || sequence.StartNode == nil {
		return errMissingSequenceFields
	}

	return validateBehaviorNode(sequence.StartNode, 0)
}

func validateBehaviorNode(value any, depth int) error {
	if depth > maxBehaviorNodeDepth {
		return errBehaviorDepthExceeded
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal behavior node: %w", err)
	}

	nodeType, err := behaviorNodeType(data)
	if err != nil {
		return err
	}

	switch nodeType {
	case string(alexamodels.ComAmazonAlexaBehaviorsModelOpaquePayloadOperationNode):
		return validateOpaqueOperationNode(data)
	case string(alexamodels.ComAmazonAlexaBehaviorsModelSerialNode):
		return validateBehaviorSequenceNode(data, depth, string(alexamodels.ComAmazonAlexaBehaviorsModelSerialNode), errSerialNodeWithoutChildren)
	case string(alexamodels.ComAmazonAlexaBehaviorsModelParallelNode):
		return validateBehaviorSequenceNode(data, depth, string(alexamodels.ComAmazonAlexaBehaviorsModelParallelNode), errParallelNodeWithoutChildren)
	default:
		return fmt.Errorf("%w %q", errUnsupportedBehaviorNode, nodeType)
	}
}

func behaviorNodeType(data []byte) (string, error) {
	var node struct {
		Type string `json:"@type"`
	}

	err := json.Unmarshal(data, &node)
	if err != nil {
		return "", fmt.Errorf("read behavior node type: %w", err)
	}

	return node.Type, nil
}

func decodeBehaviorNode(data []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()

	//nolint:wrapcheck // Preserve the decoder's JSON error type for validation callers.
	return decoder.Decode(target)
}

func validateOpaqueOperationNode(data []byte) error {
	var node alexamodels.OpaquePayloadOperationNode

	err := decodeBehaviorNode(data, &node)
	if err != nil {
		return err
	}

	if !node.Type.Valid() || node.OperationType == "" || node.OperationPayload == nil {
		return errIncompleteOperationNode
	}

	return nil
}

func validateBehaviorSequenceNode(data []byte, depth int, expectedType string, emptyNodeError error) error {
	var node struct {
		Type           string        `json:"@type"`
		NodesToExecute []interface{} `json:"nodesToExecute"`
	}

	err := decodeBehaviorNode(data, &node)
	if err != nil {
		return err
	}

	if node.Type != expectedType || len(node.NodesToExecute) == 0 {
		return emptyNodeError
	}

	for _, child := range node.NodesToExecute {
		err := validateBehaviorNode(child, depth+1)
		if err != nil {
			return err
		}
	}

	return nil
}

// SendSequence sends a sequence command to a device
// This is a helper that builds the operation node and calls RunBehavior
// The operationPayload should already contain all required fields including deviceType, deviceSerialNumber, locale, and customerId.
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
// Uses Alexa.DeviceControls.Stop sequence.
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
// Uses Alexa.DeviceControls.Volume sequence.
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
// Uses /api/np/command endpoint.
func (c *Client) PausePlayback(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type: wire.PauseCommand,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to pause playback",
			Err:     err,
		}
	}

	return nil
}

// ResumePlayback resumes playback on a device
// Uses /api/np/command endpoint.
func (c *Client) ResumePlayback(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type: wire.PlayCommand,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to resume playback",
			Err:     err,
		}
	}

	return nil
}

// NextTrack skips to the next track
// Uses /api/np/command endpoint.
func (c *Client) NextTrack(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type: wire.NextCommand,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to skip to next track",
			Err:     err,
		}
	}

	return nil
}

// PreviousTrack goes to the previous track
// Uses /api/np/command endpoint.
func (c *Client) PreviousTrack(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type: wire.PreviousCommand,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to go to previous track",
			Err:     err,
		}
	}

	return nil
}

// GetPlayerState retrieves the current player state for a device
// Uses /api/np/player endpoint.
func (c *Client) GetPlayerState(ctx context.Context, req *alexamodels.PlayerStateRequest) (*alexamodels.PlayerStateResponse, error) {
	if req.Endpoint == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPPlayer

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	var response wire.WirePlayerStateResponse

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodGetMediaPlayerState, baseURL, nil, nil, &response, true)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to get player state",
			Err:     err,
		}
	}

	return convertWireModel[alexamodels.PlayerStateResponse](response)
}

// ForwardMedia fast-forwards the current media
// Uses /api/np/command endpoint.
func (c *Client) ForwardMedia(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type: wire.ForwardCommand,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to forward media",
			Err:     err,
		}
	}

	return nil
}

// RewindMedia rewinds the current media
// Uses /api/np/command endpoint.
func (c *Client) RewindMedia(ctx context.Context, req *alexamodels.MediaControlRequest) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type: wire.RewindCommand,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to rewind media",
			Err:     err,
		}
	}

	return nil
}

// SetShuffle sets the shuffle state for media playback
// Uses /api/np/command endpoint.
func (c *Client) SetShuffle(ctx context.Context, req *alexamodels.MediaControlRequest, shuffle bool) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type:    wire.ShuffleCommand,
		Shuffle: &shuffle,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to set shuffle",
			Err:     err,
		}
	}

	return nil
}

// SetRepeat sets the repeat state for media playback
// Uses /api/np/command endpoint.
func (c *Client) SetRepeat(ctx context.Context, req *alexamodels.MediaControlRequest, repeat bool) error {
	if req.Endpoint == nil {
		return &alexaapimodels.BadRequestError{
			Message: "endpoint is required",
		}
	}

	deviceType := req.Endpoint.GetDeviceType()
	deviceSerialNumber := req.Endpoint.GetDeviceSerialNumber()

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathNPCommand

	command := wire.WireMediaCommand{
		Type:   wire.RepeatCommand,
		Repeat: &repeat,
	}

	params := url.Values{}
	params.Set(alexamodels.QueryParamDeviceSerialNumber, deviceSerialNumber)
	params.Set(alexamodels.QueryParamDeviceType, deviceType)
	baseURL += "?" + params.Encode()

	err := c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSendMediaCommand, baseURL, command, nil, nil, true)
	if err != nil {
		return &alexaapimodels.NetworkError{
			Message: "failed to set repeat",
			Err:     err,
		}
	}

	return nil
}

// SendNotification sends a mobile push notification
// Uses Alexa.Notifications.SendMobilePush sequence.
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
// Uses AlexaAnnouncement sequence.
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
// Uses Alexa.Speak or Alexa.CannedTts.Speak sequence.
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
// Uses Alexa.Music.PlaySearchPhrase sequence.
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
// Uses Alexa.Sound behavior sequence with soundStringId set to the audio URL.
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
// Uses Alexa.Operation.Video.PlaySearchPhrase sequence.
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
// This is a helper that builds the operation node and calls RunBehavior for Fire TV operations.
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

	baseURL := c.alexaAmazonBaseURI + alexamodels.APIPathBehaviorsPreview

	request := wire.WireBehaviorPreviewRequest{
		BehaviorId:   alexamodels.DefaultBehaviorID,
		SequenceJson: string(sequenceJSON),
		Status:       alexamodels.DefaultBehaviorStatus,
	}

	err = c.doJSONRequestWithFullURL(ctx, apiroutes.MethodSubmitBehaviorPreview, baseURL, request, nil, nil, true)
	if err != nil {
		return err
	}

	return nil
}

// FireTVTurnOn turns on a Fire TV device
// Uses Alexa.Operation.FireTV.TurnOn sequence.
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
// Uses Alexa.Operation.FireTV.TurnOff sequence.
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
// Uses Alexa.Operation.FireTV.TurnOn or Alexa.Operation.FireTV.TurnOff sequence.
func (c *Client) FireTVTurnOnOff(ctx context.Context, req *alexamodels.FireTVRequest, value bool) error {
	if value {
		return c.FireTVTurnOn(ctx, req)
	}

	return c.FireTVTurnOff(ctx, req)
}

// FireTVPauseVideo pauses video playback on a Fire TV device
// Uses Alexa.Operation.FireTV.PauseVideo sequence.
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
// Uses Alexa.Operation.FireTV.ResumeVideo sequence.
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
// Uses Alexa.Operation.FireTV.NavigateHome sequence.
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
