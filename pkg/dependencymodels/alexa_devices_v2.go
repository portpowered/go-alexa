package alexamodels

// AppDevice represents an app device in the appDeviceList, which means that its not really a physical device, but like an app instance.
type AppDevice struct {
	DeviceAccountId string `json:"deviceAccountId"`
	DeviceType      string `json:"deviceType"`
	SerialNumber    string `json:"serialNumber"`
}

// DeviceV2 represents a device from the devices-v2 API
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

// DevicesV2Response represents the response from the devices-v2 API
type DevicesV2Response struct {
	Devices []DeviceV2 `json:"devices"`
}

type DeviceFamily string

const (
	WholeHomeDeviceFamily DeviceFamily = "WHA"
	FireTVDeviceFamily    DeviceFamily = "FIRE_TV"
	EchoDeviceFamily      DeviceFamily = "ECHO"
	KnightDeviceFamily    DeviceFamily = "KNIGHT"
)

type CapabilityInterface string

// Notes, we don't actually know what these capabilities do internally, but we generally think that they are used to denote which widgets to expose for an endpoint.

const (
	TemperatureSensorCapability               CapabilityInterface = "TEMPERATURE_SENSOR"
	FlashBriefingCapability                   CapabilityInterface = "FLASH_BRIEFING"
	AlexaPresenceCapability                   CapabilityInterface = "ALEXA_PRESENCE"
	MotionSensorRangeStringCapability         CapabilityInterface = "MOTION_SENSOR_RANGE_STRING"
	AdaptiveListeningCapability               CapabilityInterface = "ADAPTIVE_LISTENING"
	EfdcardsCapability                        CapabilityInterface = "EFDCARDS"
	MotionDetectionCapability                 CapabilityInterface = "MOTION_DETECTION"
	AlexaVoiceCapability                      CapabilityInterface = "ALEXA_VOICE"
	MusicSkillCapability                      CapabilityInterface = "MUSIC_SKILL"
	AudioControlsCapability                   CapabilityInterface = "AUDIO_CONTROLS"
	UpdateWifiCapability                      CapabilityInterface = "UPDATE_WIFI"
	RemindersCapability                       CapabilityInterface = "REMINDERS"
	MultiWakewordsSupportedCapability         CapabilityInterface = "MULTI_WAKEWORDS_SUPPORTED"
	EarconsCapability                         CapabilityInterface = "EARCONS"
	AudioPlayerCapability                     CapabilityInterface = "AUDIO_PLAYER"
	TahoeByodCapability                       CapabilityInterface = "TAHOE_BYOD"
	PairBtSourceCapability                    CapabilityInterface = "PAIR_BT_SOURCE"
	SleepCapability                           CapabilityInterface = "SLEEP"
	IHeartRadioCapability                     CapabilityInterface = "I_HEART_RADIO"
	AudibleCapability                         CapabilityInterface = "AUDIBLE"
	SpeechRecognizerUssCapability             CapabilityInterface = "SPEECH_RECOGNIZER_USS"
	PoptartCapability                         CapabilityInterface = "POPTART"
	TupleCategoryACapability                  CapabilityInterface = "TUPLE_CATEGORY_A"
	ChangeNameCapability                      CapabilityInterface = "CHANGE_NAME"
	TupleCapability                           CapabilityInterface = "TUPLE"
	DeezerCapability                          CapabilityInterface = "DEEZER"
	AlexaDeviceRebootCapability               CapabilityInterface = "ALEXA_DEVICE_REBOOT"
	DeregisterDeviceCapability                CapabilityInterface = "DEREGISTER_DEVICE"
	TimersAndAlarmsCapability                 CapabilityInterface = "TIMERS_AND_ALARMS"
	SupportsConnectedHomeCloudOnlyCapability  CapabilityInterface = "SUPPORTS_CONNECTED_HOME_CLOUD_ONLY"
	KindleBooksCapability                     CapabilityInterface = "KINDLE_BOOKS"
	FaceToTalkCapability                      CapabilityInterface = "FACE_TO_TALK"
	GadgetsCapability                         CapabilityInterface = "GADGETS"
	EqualizerControllerTrebleCapability       CapabilityInterface = "EQUALIZER_CONTROLLER_TREBLE"
	VoiceTrainingCapability                   CapabilityInterface = "VOICE_TRAINING"
	TapGesturesResumeMediaCapability          CapabilityInterface = "TAP_GESTURES_RESUME_MEDIA"
	EqualizerControllerMidrangeCapability     CapabilityInterface = "EQUALIZER_CONTROLLER_MIDRANGE"
	WakeWordSensitivityCapability             CapabilityInterface = "WAKE_WORD_SENSITIVITY"
	FarFieldWakeWordCapability                CapabilityInterface = "FAR_FIELD_WAKE_WORD"
	DreamTrainingCapability                   CapabilityInterface = "DREAM_TRAINING"
	AppleMusicCapability                      CapabilityInterface = "APPLE_MUSIC"
	VolumeSettingCapability                   CapabilityInterface = "VOLUME_SETTING"
	SetLocaleCapability                       CapabilityInterface = "SET_LOCALE"
	PairBtSinkCapability                      CapabilityInterface = "PAIR_BT_SINK"
	SupportsLocaleCapability                  CapabilityInterface = "SUPPORTS_LOCALE"
	TimersAlarmsNotificationsVolumeCapability CapabilityInterface = "TIMERS_ALARMS_NOTIFICATIONS_VOLUME"
	AmazonMusicCapability                     CapabilityInterface = "AMAZON_MUSIC"
	EqualizerControllerBassCapability         CapabilityInterface = "EQUALIZER_CONTROLLER_BASS"
	AlexaNetworkingCapability                 CapabilityInterface = "ALEXA_NETWORKING"
	GuardEarconCapability                     CapabilityInterface = "GUARD_EARCON"
	TuneInCapability                          CapabilityInterface = "TUNE_IN"
	SupportCalendarAlertCapability            CapabilityInterface = "SUPPORT_CALENDAR_ALERT"
	AscendingAlarmVolumeCapability            CapabilityInterface = "ASCENDING_ALARM_VOLUME"
	TapGesturesCapability                     CapabilityInterface = "TAP_GESTURES"
	LemurAlphaCapability                      CapabilityInterface = "LEMUR_ALPHA"
	PandoraCapability                         CapabilityInterface = "PANDORA"
	GoldfishCapability                        CapabilityInterface = "GOLDFISH"
	SoundSettingsCapability                   CapabilityInterface = "SOUND_SETTINGS"
	SiriusxmCapability                        CapabilityInterface = "SIRIUSXM"
	TapGesturesSingleTapCapability            CapabilityInterface = "TAP_GESTURES_SINGLE_TAP"
	SupportsLocaleSwitchCapability            CapabilityInterface = "SUPPORTS_LOCALE_SWITCH"
	CustomAlarmToneCapability                 CapabilityInterface = "CUSTOM_ALARM_TONE"
	AlexaNetworkingWifiCapability             CapabilityInterface = "ALEXA_NETWORKING_WIFI"
	PersistentConnectionCapability            CapabilityInterface = "PERSISTENT_CONNECTION"
	SetTimeZoneCapability                     CapabilityInterface = "SET_TIME_ZONE"
	SupportsSoftwareVersionCapability         CapabilityInterface = "SUPPORTS_SOFTWARE_VERSION"
	BtPairingFlowV2Capability                 CapabilityInterface = "BT_PAIRING_FLOW_V2"
	DialogInterfaceVersionCapability          CapabilityInterface = "DIALOG_INTERFACE_VERSION"
	RequiresOobeForSetupCapability            CapabilityInterface = "REQUIRES_OOBE_FOR_SETUP"
	MicrophoneCapability                      CapabilityInterface = "MICROPHONE"
	ActiveAfterFroCapability                  CapabilityInterface = "ACTIVE_AFTER_FRO"
	DsVolumeSettingCapability                 CapabilityInterface = "DS_VOLUME_SETTING"
	TidalCapability                           CapabilityInterface = "TIDAL"
	AdaptiveVolumeCapability                  CapabilityInterface = "ADAPTIVE_VOLUME"
	SalmonCapability                          CapabilityInterface = "SALMON"
	AlexaNetworkingSpeedtestCapability        CapabilityInterface = "ALEXA_NETWORKING_SPEEDTEST"
	LocalVoiceCapability                      CapabilityInterface = "LOCAL_VOICE"
	DeregisterFactoryResetCapability          CapabilityInterface = "DEREGISTER_FACTORY_RESET"
	LocalizationCapability                    CapabilityInterface = "LOCALIZATION"
	PairRemoteCapability                      CapabilityInterface = "PAIR_REMOTE"
	AuxSettingsCapability                     CapabilityInterface = "AUX_SETTINGS"
	AlexaCaptioningPreferencesCapability      CapabilityInterface = "ALEXA_CAPTIONING_PREFERENCES"
	ClockFormat24HrCapability                 CapabilityInterface = "CLOCK_FORMAT_24_HR"
	LiveViewCapability                        CapabilityInterface = "LIVE_VIEW"
	AlexaCaptioningCapability                 CapabilityInterface = "ALEXA_CAPTIONING"
	SharknadoCapability                       CapabilityInterface = "SHARKNADO"
	TupleCategoryBCapability                  CapabilityInterface = "TUPLE_CATEGORY_B"
	MediaPlayerAnimationCapability            CapabilityInterface = "MEDIA_PLAYER_ANIMATION"
	TimezoneCapability                        CapabilityInterface = "TIMEZONE"
	ArthurTargetCapability                    CapabilityInterface = "ARTHUR_TARGET"
	SupportsConnectedHomeAllCapability        CapabilityInterface = "SUPPORTS_CONNECTED_HOME_ALL"
)
