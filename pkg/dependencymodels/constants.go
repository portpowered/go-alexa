package alexamodels

import "github.com/portpowered/go-alexa/pkg/internal/apiroutes"

// API Paths.
const (
	APIPathV2Endpoints             = apiroutes.PathListRestEndpoints
	APIPathV2EndpointQuery         = apiroutes.PathQueryRestEndpoints
	APIPathV2EndpointsForget       = apiroutes.PathForgetEndpoint
	APIPathV2EndpointsDeregister   = apiroutes.PathDeregisterEndpoint
	APIPathV2EndpointsFriendlyName = apiroutes.PathUpdateEndpointFriendlyName
	APIPathV2EndpointsControl      = apiroutes.PathControlRestEndpoint
	APIPathV2EndpointInterface     = apiroutes.PathSendEndpointInterfaceMessage
	APIPathDevicesV2Device         = apiroutes.PathListFirstPartyDevices
	APIPathBehaviorsPreview        = apiroutes.PathSubmitBehaviorPreview
	APIPathNPCommand               = apiroutes.PathSendMediaCommand
	APIPathNPPlayer                = apiroutes.PathGetMediaPlayerState
)

// Query Parameters.
const (
	QueryParamOwner              = apiroutes.QueryParamOwner
	QueryParamExpand             = apiroutes.QueryParamExpand
	QueryParamMaxResults         = apiroutes.QueryParamMaxResults
	QueryParamNextToken          = apiroutes.QueryParamNextToken
	QueryParamDeviceSerialNumber = apiroutes.QueryParamDeviceSerialNumber
	QueryParamDeviceType         = apiroutes.QueryParamDeviceType
	QueryParamCSRF               = apiroutes.HeaderCsrf
)

// Cookie Header.
const (
	CookieHeaderName = apiroutes.HeaderCookie
	CookieCSRFFormat = apiroutes.HeaderCsrf + "=%s"
)
