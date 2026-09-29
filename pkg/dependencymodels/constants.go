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

// Default Values.
const (
	DefaultOwnerCaller             = "~caller"
	DefaultLocale                  = "en-US"
	DefaultFriendlyNameType        = "PLAIN"
	DefaultBehaviorID              = "PREVIEW"
	DefaultBehaviorStatus          = "ENABLED"
	DefaultAnnouncementExpireAfter = "PT5S"
	DefaultAnnouncementSpeakType   = "text"
)

// Announcement Methods.
const (
	AnnouncementMethodSpeak = "speak"
	AnnouncementMethodShow  = "show"
	AnnouncementMethodAll   = "all"
)

// Device Families.
const (
	DeviceFamilyFireTV = "FIRE_TV"
)

// Skill IDs.
const (
	SkillIDAlexaDeviceControls = "amzn1.ask.1p.alexadevicecontrols"
	SkillIDRoutinesMessaging   = "amzn1.ask.1p.routines.messaging"
	SkillIDAlexaNotifications  = "amzn1.ask.1p.alexanotifications"
	SkillIDSaySomething        = "amzn1.ask.1p.saysomething"
	SkillIDRoutinesFireTV      = "amzn1.ask.1p.routines.firetv"
)

// Operation Types.
const (
	OperationTypeDeviceControlsStop          = "Alexa.DeviceControls.Stop"
	OperationTypeDeviceControlsVolume        = "Alexa.DeviceControls.Volume"
	OperationTypeNotificationsSendMobilePush = "Alexa.Notifications.SendMobilePush"
	OperationTypeAnnouncement                = "AlexaAnnouncement"
	OperationTypeSpeak                       = "Alexa.Speak"
	OperationTypeCannedTtsSpeak              = "Alexa.CannedTts.Speak"
	OperationTypeMusicPlaySearchPhrase       = "Alexa.Music.PlaySearchPhrase"
	OperationTypeVideoPlaySearchPhrase       = "Alexa.Operation.Video.PlaySearchPhrase"
	OperationTypeFireTVTurnOn                = "Alexa.Operation.FireTV.TurnOn"
	OperationTypeFireTVTurnOff               = "Alexa.Operation.FireTV.TurnOff"
	OperationTypeFireTVPauseVideo            = "Alexa.Operation.FireTV.PauseVideo"
	OperationTypeFireTVResumeVideo           = "Alexa.Operation.FireTV.ResumeVideo"
	OperationTypeFireTVNavigateHome          = "Alexa.Operation.FireTV.NavigateHome"
	OperationTypeSound                       = "Alexa.Sound"
)

// Model Types.
const (
	ModelTypeOpaquePayloadOperationNode = "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode"
	ModelTypeSequence                   = "com.amazon.alexa.behaviors.model.Sequence"
)

// Media Command Types.
const (
	MediaCommandTypePause    = "PauseCommand"
	MediaCommandTypePlay     = "PlayCommand"
	MediaCommandTypeNext     = "NextCommand"
	MediaCommandTypePrevious = "PreviousCommand"
	MediaCommandTypeForward  = "ForwardCommand"
	MediaCommandTypeRewind   = "RewindCommand"
	MediaCommandTypeShuffle  = "ShuffleCommand"
	MediaCommandTypeRepeat   = "RepeatCommand"
)

// Payload Keys.
const (
	PayloadKeyDeviceType            = "deviceType"
	PayloadKeyDeviceSerialNumber    = "deviceSerialNumber"
	PayloadKeyLocale                = "locale"
	PayloadKeyCustomerID            = "customerId"
	PayloadKeySkillID               = "skillId"
	PayloadKeyValue                 = "value"
	PayloadKeyNotificationMessage   = "notificationMessage"
	PayloadKeyAlexaURL              = "alexaUrl"
	PayloadKeyTitle                 = "title"
	PayloadKeyExpireAfter           = "expireAfter"
	PayloadKeyContent               = "content"
	PayloadKeyTarget                = "target"
	PayloadKeyTextToSpeak           = "textToSpeak"
	PayloadKeyCannedTtsStringID     = "cannedTtsStringId"
	PayloadKeySearchPhrase          = "searchPhrase"
	PayloadKeySanitizedSearchPhrase = "sanitizedSearchPhrase"
	PayloadKeyMusicProviderID       = "musicProviderId"
	PayloadKeyWaitTimeInSeconds     = "waitTimeInSeconds"
	PayloadKeyDeviceAccountID       = "deviceAccountId"
	PayloadKeySoundStringID         = "soundStringId"
)

// Cookie Header.
const (
	CookieHeaderName = apiroutes.HeaderCookie
	CookieCSRFFormat = "csrf=%s"
)

// Alexa URL.
const (
	AlexaURLBehaviors = "#v2/behaviors"
)
