package alexamodels

// API Paths
const (
	APIPathV2Endpoints             = "/v2/endpoints"
	APIPathV2EndpointQuery         = "/v2/endpoint-query"
	APIPathV2EndpointsForget       = "/v2/endpoints/%s/forget"
	APIPathV2EndpointsDeregister   = "/v2/endpoints/%s/deregister"
	APIPathV2EndpointsFriendlyName = "/v2/endpoints/%s/friendlyName"
	APIPathV2EndpointsControl      = "/v2/endpoints/%s/control"
	APIPathV2EndpointInterface     = "/v2/endpoints/%s/interfaces/%s/%s/"
	APIPathDevicesV2Device         = "/api/devices-v2/device"
	APIPathBehaviorsPreview        = "/api/behaviors/preview"
	APIPathNPCommand               = "/api/np/command"
	APIPathNPPlayer                = "/api/np/player"
)

// Query Parameters
const (
	QueryParamOwner              = "owner"
	QueryParamExpand             = "expand"
	QueryParamMaxResults         = "maxResults"
	QueryParamNextToken          = "nextToken"
	QueryParamDeviceSerialNumber = "deviceSerialNumber"
	QueryParamDeviceType         = "deviceType"
	QueryParamCSRF               = "csrf"
)

// Default Values
const (
	DefaultOwnerCaller             = "~caller"
	DefaultLocale                  = "en-US"
	DefaultFriendlyNameType        = "PLAIN"
	DefaultBehaviorID              = "PREVIEW"
	DefaultBehaviorStatus          = "ENABLED"
	DefaultAnnouncementExpireAfter = "PT5S"
	DefaultAnnouncementSpeakType   = "text"
)

// Announcement Methods
const (
	AnnouncementMethodSpeak = "speak"
	AnnouncementMethodShow  = "show"
	AnnouncementMethodAll   = "all"
)

// Device Families
const (
	DeviceFamilyFireTV = "FIRE_TV"
)

// Skill IDs
const (
	SkillIDAlexaDeviceControls = "amzn1.ask.1p.alexadevicecontrols"
	SkillIDRoutinesMessaging   = "amzn1.ask.1p.routines.messaging"
	SkillIDAlexaNotifications  = "amzn1.ask.1p.alexanotifications"
	SkillIDSaySomething        = "amzn1.ask.1p.saysomething"
	SkillIDRoutinesFireTV      = "amzn1.ask.1p.routines.firetv"
)

// Operation Types
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

// Model Types
const (
	ModelTypeOpaquePayloadOperationNode = "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode"
	ModelTypeSequence                   = "com.amazon.alexa.behaviors.model.Sequence"
)

// Media Command Types
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

// Payload Keys
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

// Cookie Header
const (
	CookieHeaderName = "Cookie"
	CookieCSRFFormat = "csrf=%s"
)

// Alexa URL
const (
	AlexaURLBehaviors = "#v2/behaviors"
)
