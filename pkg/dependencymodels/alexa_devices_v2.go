package alexamodels

// AppDevice represents an app device in the appDeviceList, which means that its not really a physical device, but like an app instance.
type AppDevice struct {
	DeviceAccountId string `json:"deviceAccountId"`
	DeviceType      string `json:"deviceType"`
	SerialNumber    string `json:"serialNumber"`
}

// DeviceV2 represents a device from the devices-v2 API.
type DeviceV2 struct {
	AccountName            string      `json:"accountName"`
	AppDeviceList          []AppDevice `json:"appDeviceList"`
	AssociatedUnitIds      interface{} `json:"associatedUnitIds"` // Can be null or array
	Capabilities           []string    `json:"capabilities"`
	Charging               *bool       `json:"charging"` // Can be null
	ClusterMembers         []string    `json:"clusterMembers"`
	DeviceAccountId        string      `json:"deviceAccountId"`
	DeviceFamily           string      `json:"deviceFamily"`
	DeviceOwnerCustomerId  string      `json:"deviceOwnerCustomerId"`
	DeviceType             string      `json:"deviceType"`
	DeviceTypeFriendlyName *string     `json:"deviceTypeFriendlyName"` // Can be null
	Essid                  *string     `json:"essid"`                  // Can be null
	Language               *string     `json:"language"`               // Can be null
	MacAddress             *string     `json:"macAddress"`             // Can be null
	Online                 bool        `json:"online"`
	ParentClusters         []string    `json:"parentClusters"`
	PostalCode             *string     `json:"postalCode"`            // Can be null
	RegistrationId         *string     `json:"registrationId"`        // Can be null
	RemainingBatteryLevel  *int        `json:"remainingBatteryLevel"` // Can be null
	SerialNumber           string      `json:"serialNumber"`
	SoftwareVersion        string      `json:"softwareVersion"`
}

// DevicesV2Response represents the response from the devices-v2 API.
type DevicesV2Response struct {
	Devices []DeviceV2 `json:"devices"`
}

// DeviceFamily identifies a family of Alexa devices.
type DeviceFamily string

const (
	// WholeHomeDeviceFamily identifies a whole-home device family.
	WholeHomeDeviceFamily DeviceFamily = "WHA"
	// FireTVDeviceFamily identifies Fire TV devices.
	FireTVDeviceFamily DeviceFamily = "FIRE_TV"
	// EchoDeviceFamily identifies Echo devices.
	EchoDeviceFamily DeviceFamily = "ECHO"
	// KnightDeviceFamily identifies Knight devices.
	KnightDeviceFamily DeviceFamily = "KNIGHT"
)

// CapabilityInterface identifies a capability reported by the devices API.
type CapabilityInterface string

// We do not know what these capabilities do internally. We believe they
// identify which widgets to expose for an endpoint.

const (
	// TemperatureSensorCapability identifies temperature sensor support.
	TemperatureSensorCapability CapabilityInterface = "TEMPERATURE_SENSOR"
	// FlashBriefingCapability identifies flash briefing support.
	FlashBriefingCapability CapabilityInterface = "FLASH_BRIEFING"
	// AlexaPresenceCapability identifies Alexa presence support.
	AlexaPresenceCapability CapabilityInterface = "ALEXA_PRESENCE"
	// MotionSensorRangeStringCapability identifies string-valued motion-sensor ranges.
	MotionSensorRangeStringCapability CapabilityInterface = "MOTION_SENSOR_RANGE_STRING"
	// AdaptiveListeningCapability identifies adaptive-listening support.
	AdaptiveListeningCapability CapabilityInterface = "ADAPTIVE_LISTENING"
	// EfdcardsCapability identifies EFDCARDS support.
	EfdcardsCapability CapabilityInterface = "EFDCARDS"
	// MotionDetectionCapability identifies devices that expose motion detection.
	MotionDetectionCapability CapabilityInterface = "MOTION_DETECTION"
	// AlexaVoiceCapability identifies Alexa voice support.
	AlexaVoiceCapability CapabilityInterface = "ALEXA_VOICE"
	// MusicSkillCapability identifies music-skill support.
	MusicSkillCapability CapabilityInterface = "MUSIC_SKILL"
	// AudioControlsCapability identifies this device capability.
	AudioControlsCapability CapabilityInterface = "AUDIO_CONTROLS"
	// UpdateWifiCapability identifies this device capability.
	UpdateWifiCapability CapabilityInterface = "UPDATE_WIFI"
	// RemindersCapability identifies this device capability.
	RemindersCapability CapabilityInterface = "REMINDERS"
	// MultiWakewordsSupportedCapability identifies this device capability.
	MultiWakewordsSupportedCapability CapabilityInterface = "MULTI_WAKEWORDS_SUPPORTED"
	// EarconsCapability identifies this device capability.
	EarconsCapability CapabilityInterface = "EARCONS"
	// AudioPlayerCapability identifies this device capability.
	AudioPlayerCapability CapabilityInterface = "AUDIO_PLAYER"
	// TahoeByodCapability identifies this device capability.
	TahoeByodCapability CapabilityInterface = "TAHOE_BYOD"
	// PairBtSourceCapability identifies this device capability.
	PairBtSourceCapability CapabilityInterface = "PAIR_BT_SOURCE"
	// SleepCapability identifies this device capability.
	SleepCapability CapabilityInterface = "SLEEP"
	// IHeartRadioCapability identifies this device capability.
	IHeartRadioCapability CapabilityInterface = "I_HEART_RADIO"
	// AudibleCapability identifies this device capability.
	AudibleCapability CapabilityInterface = "AUDIBLE"
	// SpeechRecognizerUssCapability identifies this device capability.
	SpeechRecognizerUssCapability CapabilityInterface = "SPEECH_RECOGNIZER_USS"
	// PoptartCapability identifies this device capability.
	PoptartCapability CapabilityInterface = "POPTART"
	// TupleCategoryACapability identifies this device capability.
	TupleCategoryACapability CapabilityInterface = "TUPLE_CATEGORY_A"
	// ChangeNameCapability identifies this device capability.
	ChangeNameCapability CapabilityInterface = "CHANGE_NAME"
	// TupleCapability identifies this device capability.
	TupleCapability CapabilityInterface = "TUPLE"
	// DeezerCapability identifies this device capability.
	DeezerCapability CapabilityInterface = "DEEZER"
	// AlexaDeviceRebootCapability identifies this device capability.
	AlexaDeviceRebootCapability CapabilityInterface = "ALEXA_DEVICE_REBOOT"
	// DeregisterDeviceCapability identifies this device capability.
	DeregisterDeviceCapability CapabilityInterface = "DEREGISTER_DEVICE"
	// TimersAndAlarmsCapability identifies this device capability.
	TimersAndAlarmsCapability CapabilityInterface = "TIMERS_AND_ALARMS"
	// SupportsConnectedHomeCloudOnlyCapability identifies this device capability.
	SupportsConnectedHomeCloudOnlyCapability CapabilityInterface = "SUPPORTS_CONNECTED_HOME_CLOUD_ONLY"
	// KindleBooksCapability identifies this device capability.
	KindleBooksCapability CapabilityInterface = "KINDLE_BOOKS"
	// FaceToTalkCapability identifies this device capability.
	FaceToTalkCapability CapabilityInterface = "FACE_TO_TALK"
	// GadgetsCapability identifies this device capability.
	GadgetsCapability CapabilityInterface = "GADGETS"
	// EqualizerControllerTrebleCapability identifies this device capability.
	EqualizerControllerTrebleCapability CapabilityInterface = "EQUALIZER_CONTROLLER_TREBLE"
	// VoiceTrainingCapability identifies this device capability.
	VoiceTrainingCapability CapabilityInterface = "VOICE_TRAINING"
	// TapGesturesResumeMediaCapability identifies this device capability.
	TapGesturesResumeMediaCapability CapabilityInterface = "TAP_GESTURES_RESUME_MEDIA"
	// EqualizerControllerMidrangeCapability identifies this device capability.
	EqualizerControllerMidrangeCapability CapabilityInterface = "EQUALIZER_CONTROLLER_MIDRANGE"
	// WakeWordSensitivityCapability identifies this device capability.
	WakeWordSensitivityCapability CapabilityInterface = "WAKE_WORD_SENSITIVITY"
	// FarFieldWakeWordCapability identifies this device capability.
	FarFieldWakeWordCapability CapabilityInterface = "FAR_FIELD_WAKE_WORD"
	// DreamTrainingCapability identifies this device capability.
	DreamTrainingCapability CapabilityInterface = "DREAM_TRAINING"
	// AppleMusicCapability identifies this device capability.
	AppleMusicCapability CapabilityInterface = "APPLE_MUSIC"
	// VolumeSettingCapability identifies this device capability.
	VolumeSettingCapability CapabilityInterface = "VOLUME_SETTING"
	// SetLocaleCapability identifies this device capability.
	SetLocaleCapability CapabilityInterface = "SET_LOCALE"
	// PairBtSinkCapability identifies this device capability.
	PairBtSinkCapability CapabilityInterface = "PAIR_BT_SINK"
	// SupportsLocaleCapability identifies this device capability.
	SupportsLocaleCapability CapabilityInterface = "SUPPORTS_LOCALE"
	// TimersAlarmsNotificationsVolumeCapability identifies this device capability.
	TimersAlarmsNotificationsVolumeCapability CapabilityInterface = "TIMERS_ALARMS_NOTIFICATIONS_VOLUME"
	// AmazonMusicCapability identifies this device capability.
	AmazonMusicCapability CapabilityInterface = "AMAZON_MUSIC"
	// EqualizerControllerBassCapability identifies this device capability.
	EqualizerControllerBassCapability CapabilityInterface = "EQUALIZER_CONTROLLER_BASS"
	// AlexaNetworkingCapability identifies this device capability.
	AlexaNetworkingCapability CapabilityInterface = "ALEXA_NETWORKING"
	// GuardEarconCapability identifies this device capability.
	GuardEarconCapability CapabilityInterface = "GUARD_EARCON"
	// TuneInCapability identifies this device capability.
	TuneInCapability CapabilityInterface = "TUNE_IN"
	// SupportCalendarAlertCapability identifies this device capability.
	SupportCalendarAlertCapability CapabilityInterface = "SUPPORT_CALENDAR_ALERT"
	// AscendingAlarmVolumeCapability identifies this device capability.
	AscendingAlarmVolumeCapability CapabilityInterface = "ASCENDING_ALARM_VOLUME"
	// TapGesturesCapability identifies this device capability.
	TapGesturesCapability CapabilityInterface = "TAP_GESTURES"
	// LemurAlphaCapability identifies this device capability.
	LemurAlphaCapability CapabilityInterface = "LEMUR_ALPHA"
	// PandoraCapability identifies this device capability.
	PandoraCapability CapabilityInterface = "PANDORA"
	// GoldfishCapability identifies this device capability.
	GoldfishCapability CapabilityInterface = "GOLDFISH"
	// SoundSettingsCapability identifies this device capability.
	SoundSettingsCapability CapabilityInterface = "SOUND_SETTINGS"
	// SiriusxmCapability identifies this device capability.
	SiriusxmCapability CapabilityInterface = "SIRIUSXM"
	// TapGesturesSingleTapCapability identifies this device capability.
	TapGesturesSingleTapCapability CapabilityInterface = "TAP_GESTURES_SINGLE_TAP"
	// SupportsLocaleSwitchCapability identifies this device capability.
	SupportsLocaleSwitchCapability CapabilityInterface = "SUPPORTS_LOCALE_SWITCH"
	// CustomAlarmToneCapability identifies this device capability.
	CustomAlarmToneCapability CapabilityInterface = "CUSTOM_ALARM_TONE"
	// AlexaNetworkingWifiCapability identifies this device capability.
	AlexaNetworkingWifiCapability CapabilityInterface = "ALEXA_NETWORKING_WIFI"
	// PersistentConnectionCapability identifies this device capability.
	PersistentConnectionCapability CapabilityInterface = "PERSISTENT_CONNECTION"
	// SetTimeZoneCapability identifies this device capability.
	SetTimeZoneCapability CapabilityInterface = "SET_TIME_ZONE"
	// SupportsSoftwareVersionCapability identifies this device capability.
	SupportsSoftwareVersionCapability CapabilityInterface = "SUPPORTS_SOFTWARE_VERSION"
	// BtPairingFlowV2Capability identifies this device capability.
	BtPairingFlowV2Capability CapabilityInterface = "BT_PAIRING_FLOW_V2"
	// DialogInterfaceVersionCapability identifies this device capability.
	DialogInterfaceVersionCapability CapabilityInterface = "DIALOG_INTERFACE_VERSION"
	// RequiresOobeForSetupCapability identifies this device capability.
	RequiresOobeForSetupCapability CapabilityInterface = "REQUIRES_OOBE_FOR_SETUP"
	// MicrophoneCapability identifies this device capability.
	MicrophoneCapability CapabilityInterface = "MICROPHONE"
	// ActiveAfterFroCapability identifies this device capability.
	ActiveAfterFroCapability CapabilityInterface = "ACTIVE_AFTER_FRO"
	// DsVolumeSettingCapability identifies this device capability.
	DsVolumeSettingCapability CapabilityInterface = "DS_VOLUME_SETTING"
	// TidalCapability identifies this device capability.
	TidalCapability CapabilityInterface = "TIDAL"
	// AdaptiveVolumeCapability identifies this device capability.
	AdaptiveVolumeCapability CapabilityInterface = "ADAPTIVE_VOLUME"
	// SalmonCapability identifies this device capability.
	SalmonCapability CapabilityInterface = "SALMON"
	// AlexaNetworkingSpeedtestCapability identifies this device capability.
	AlexaNetworkingSpeedtestCapability CapabilityInterface = "ALEXA_NETWORKING_SPEEDTEST"
	// LocalVoiceCapability identifies this device capability.
	LocalVoiceCapability CapabilityInterface = "LOCAL_VOICE"
	// DeregisterFactoryResetCapability identifies this device capability.
	DeregisterFactoryResetCapability CapabilityInterface = "DEREGISTER_FACTORY_RESET"
	// LocalizationCapability identifies this device capability.
	LocalizationCapability CapabilityInterface = "LOCALIZATION"
	// PairRemoteCapability identifies this device capability.
	PairRemoteCapability CapabilityInterface = "PAIR_REMOTE"
	// AuxSettingsCapability identifies this device capability.
	AuxSettingsCapability CapabilityInterface = "AUX_SETTINGS"
	// AlexaCaptioningPreferencesCapability identifies this device capability.
	AlexaCaptioningPreferencesCapability CapabilityInterface = "ALEXA_CAPTIONING_PREFERENCES"
	// ClockFormat24HrCapability identifies this device capability.
	ClockFormat24HrCapability CapabilityInterface = "CLOCK_FORMAT_24_HR"
	// LiveViewCapability identifies this device capability.
	LiveViewCapability CapabilityInterface = "LIVE_VIEW"
	// AlexaCaptioningCapability identifies this device capability.
	AlexaCaptioningCapability CapabilityInterface = "ALEXA_CAPTIONING"
	// SharknadoCapability identifies this device capability.
	SharknadoCapability CapabilityInterface = "SHARKNADO"
	// TupleCategoryBCapability identifies this device capability.
	TupleCategoryBCapability CapabilityInterface = "TUPLE_CATEGORY_B"
	// MediaPlayerAnimationCapability identifies this device capability.
	MediaPlayerAnimationCapability CapabilityInterface = "MEDIA_PLAYER_ANIMATION"
	// TimezoneCapability identifies this device capability.
	TimezoneCapability CapabilityInterface = "TIMEZONE"
	// ArthurTargetCapability identifies this device capability.
	ArthurTargetCapability CapabilityInterface = "ARTHUR_TARGET"
	// SupportsConnectedHomeAllCapability identifies this device capability.
	SupportsConnectedHomeAllCapability CapabilityInterface = "SUPPORTS_CONNECTED_HOME_ALL"
)
