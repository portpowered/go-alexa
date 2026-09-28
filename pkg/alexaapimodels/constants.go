package alexaapimodels

const (
	// authorization URIs for each region
	AuthorizationUriNa = "https://api.amazon.com/ap/oa"
	AuthorizationUriEu = "https://api.amazon.co.uk/ap/oa"
	AuthorizationUriJp = "https://api.amazon.co.jp/ap/oa"

	// Amazon API services
	AmazonApiServiceUriNa = "https://api.amazon.com"
	AmazonApiServiceUriEu = "https://api.amazon.co.uk"
	AmazonApiServiceUriJp = "https://api.amazon.co.jp"

	// Alexa API services
	ApiServiceUriNa = "https://api.amazonalexa.com"
	ApiServiceUriEu = "https://api.eu.amazonalexa.com"
	ApiServiceUriJp = "https://api.jp.amazonalexa.com"

	// Alexa Amazon web domain (for user info and other web APIs)
	AlexaAmazonBaseUriNa = "https://alexa.amazon.com"
	AlexaAmazonBaseUriEu = "https://alexa.amazon.co.uk"
	AlexaAmazonBaseUriJp = "https://alexa.amazon.co.jp"

	// HTTP2 Connection endpoints for each region
	// https://developer.amazon.com/en-US/docs/alexa/alexa-voice-service/api-overview.html#endpoints
	Http2ConnectionUriNa = "alexa.na.gateway.devices.a2z.com"
	Http2ConnectionUriEu = "alexa.eu.gateway.devices.a2z.com"
	Http2ConnectionUriJp = "alexa.fe.gateway.devices.a2z.com"

	// First party devices within Amazon are registered with a unique device type that is used to uniquely identify the device class, such as mobile alexa app,  etc.
	DeviceTypeIphone    = "A2IVLV5VM2W81"
	DeviceTypeSimulator = "A3UMCOIO3URGGU"

	// Event names for endpoint feature state changes
	EventNameColorTemperatureState = "colorTemperatureState"
	EventNamePowerState            = "powerstate"
	EventNameSpeakerState          = "speakerState"
	EventNameBrightnessState       = "brightnessState"
	EventNameColorState            = "colorState"
	EventNameLockState             = "lockState"
	EventNameModeState             = "modeState"
	EventNameRangeState            = "rangeState"
	EventNameToggleState           = "toggleState"
	EventNamePercentageState       = "percentageState"
	EventNamePowerLevelState       = "powerLevelState"
	EventNameThermostatModeState   = "thermostatModeState"
	EventNameSetpointState         = "setpointState"
	EventNameTemperatureState      = "temperatureState"
	EventNameDetectionState        = "detectionState"
	EventNameActionState           = "actionState"
	EventNameReachabilityState     = "reachabilityState"
	EventNameBatteryState          = "batteryState"
	EventNameIlluminanceState      = "illuminanceState"
	EventNameGeolocationState      = "geolocationState"
	EventNameStatusCodeState       = "statusCodeState"
	EventNameArmState              = "armState"
	EventNameRelativeHumidityState = "relativeHumidityState"
)

type EndpointDisplayCategory string

var validEndpointDisplayCategory = map[EndpointDisplayCategory]struct{}{
	EndpointDisplayCategoryOther:                {},
	EndpointDisplayCategoryActivityTrigger:      {},
	EndpointDisplayCategoryAirConditioner:       {},
	EndpointDisplayCategoryAirFreshener:         {},
	EndpointDisplayCategoryAirPurifier:          {},
	EndpointDisplayCategoryAirQualityMonitor:    {},
	EndpointDisplayCategoryAlexaVoiceEnabled:    {},
	EndpointDisplayCategoryAutoAccessory:        {},
	EndpointDisplayCategoryBluetoothSpeaker:     {},
	EndpointDisplayCategoryBurner:               {},
	EndpointDisplayCategoryCamera:               {},
	EndpointDisplayCategoryCarbonMonoxideSensor: {},
	EndpointDisplayCategoryChristmasTree:        {},
	EndpointDisplayCategoryCoffeeMaker:          {},
	EndpointDisplayCategoryComputer:             {},
	EndpointDisplayCategoryContactSensor:        {},
	EndpointDisplayCategoryCookingAppliance:     {},
	EndpointDisplayCategoryCooktop:              {},
	EndpointDisplayCategoryCoolingCabinet:       {},
	EndpointDisplayCategoryDishwasher:           {},
	EndpointDisplayCategoryDoor:                 {},
	EndpointDisplayCategoryDoorbell:             {},
	EndpointDisplayCategoryDryer:                {},
	EndpointDisplayCategoryExteriorBlind:        {},
	EndpointDisplayCategoryFan:                  {},
	EndpointDisplayCategoryGameConsole:          {},
	EndpointDisplayCategoryGarageDoor:           {},
	EndpointDisplayCategoryHeadphones:           {},
	EndpointDisplayCategoryHub:                  {},
	EndpointDisplayCategoryHumiditySensor:       {},
	EndpointDisplayCategoryInteriorBlind:        {},
	EndpointDisplayCategoryLaptop:               {},
	EndpointDisplayCategoryLight:                {},
	EndpointDisplayCategoryLightSensor:          {},
	EndpointDisplayCategoryMicrowave:            {},
	EndpointDisplayCategoryMobilePhone:          {},
	EndpointDisplayCategoryMotionSensor:         {},
	EndpointDisplayCategoryMusicSystem:          {},
	EndpointDisplayCategoryNetworkHardware:      {},
	EndpointDisplayCategoryOven:                 {},
	EndpointDisplayCategoryPhone:                {},
	EndpointDisplayCategoryPrinter:              {},
	EndpointDisplayCategoryRefrigerator:         {},
	EndpointDisplayCategoryRemote:               {},
	EndpointDisplayCategoryRouter:               {},
	EndpointDisplayCategorySceneTrigger:         {},
	EndpointDisplayCategoryScreen:               {},
	EndpointDisplayCategorySecurityPanel:        {},
	EndpointDisplayCategorySecuritySystem:       {},
	EndpointDisplayCategorySlowCooker:           {},
	EndpointDisplayCategorySmartlock:            {},
	EndpointDisplayCategorySmartplug:            {},
	EndpointDisplayCategorySmokeSensor:          {},
	EndpointDisplayCategorySpeaker:              {},
	EndpointDisplayCategoryStreamingDevice:      {},
	EndpointDisplayCategorySwitch:               {},
	EndpointDisplayCategoryTablet:               {},
	EndpointDisplayCategoryTemperatureSensor:    {},
	EndpointDisplayCategoryThermostat:           {},
	EndpointDisplayCategoryTracker:              {},
	EndpointDisplayCategoryTV:                   {},
	EndpointDisplayCategoryVacuumCleaner:        {},
	EndpointDisplayCategoryVehicle:              {},
	EndpointDisplayCategoryWasher:               {},
	EndpointDisplayCategoryWaterHeater:          {},
	EndpointDisplayCategoryWaterLeakSensor:      {},
	EndpointDisplayCategoryWearable:             {},
}

func ValidDisplayCategory(s string) bool {
	_, ok := validEndpointDisplayCategory[EndpointDisplayCategory(s)]
	return ok
}

const (
	EndpointDisplayCategoryOther                EndpointDisplayCategory = "OTHER"
	EndpointDisplayCategoryActivityTrigger      EndpointDisplayCategory = "ACTIVITY_TRIGGER"
	EndpointDisplayCategoryAirConditioner       EndpointDisplayCategory = "AIR_CONDITIONER"
	EndpointDisplayCategoryAirFreshener         EndpointDisplayCategory = "AIR_FRESHENER"
	EndpointDisplayCategoryAirPurifier          EndpointDisplayCategory = "AIR_PURIFIER"
	EndpointDisplayCategoryAirQualityMonitor    EndpointDisplayCategory = "AIR_QUALITY_MONITOR"
	EndpointDisplayCategoryAlexaVoiceEnabled    EndpointDisplayCategory = "ALEXA_VOICE_ENABLED"
	EndpointDisplayCategoryAutoAccessory        EndpointDisplayCategory = "AUTO_ACCESSORY"
	EndpointDisplayCategoryBluetoothSpeaker     EndpointDisplayCategory = "BLUETOOTH_SPEAKER"
	EndpointDisplayCategoryBurner               EndpointDisplayCategory = "BURNER"
	EndpointDisplayCategoryCamera               EndpointDisplayCategory = "CAMERA"
	EndpointDisplayCategoryCarbonMonoxideSensor EndpointDisplayCategory = "CARBON_MONOXIDE_SENSOR"
	EndpointDisplayCategoryChristmasTree        EndpointDisplayCategory = "CHRISTMAS_TREE"
	EndpointDisplayCategoryCoffeeMaker          EndpointDisplayCategory = "COFFEE_MAKER"
	EndpointDisplayCategoryComputer             EndpointDisplayCategory = "COMPUTER"
	EndpointDisplayCategoryContactSensor        EndpointDisplayCategory = "CONTACT_SENSOR"
	EndpointDisplayCategoryCookingAppliance     EndpointDisplayCategory = "COOKING_APPLIANCE"
	EndpointDisplayCategoryCooktop              EndpointDisplayCategory = "COOKTOP"
	EndpointDisplayCategoryCoolingCabinet       EndpointDisplayCategory = "COOLING_CABINET"
	EndpointDisplayCategoryDishwasher           EndpointDisplayCategory = "DISHWASHER"
	EndpointDisplayCategoryDoor                 EndpointDisplayCategory = "DOOR"
	EndpointDisplayCategoryDoorbell             EndpointDisplayCategory = "DOORBELL"
	EndpointDisplayCategoryDryer                EndpointDisplayCategory = "DRYER"
	EndpointDisplayCategoryExteriorBlind        EndpointDisplayCategory = "EXTERIOR_BLIND"
	EndpointDisplayCategoryFan                  EndpointDisplayCategory = "FAN"
	EndpointDisplayCategoryGameConsole          EndpointDisplayCategory = "GAME_CONSOLE"
	EndpointDisplayCategoryGarageDoor           EndpointDisplayCategory = "GARAGE_DOOR"
	EndpointDisplayCategoryHeadphones           EndpointDisplayCategory = "HEADPHONES"
	EndpointDisplayCategoryHub                  EndpointDisplayCategory = "HUB"
	EndpointDisplayCategoryHumiditySensor       EndpointDisplayCategory = "HUMIDITY_SENSOR"
	EndpointDisplayCategoryInteriorBlind        EndpointDisplayCategory = "INTERIOR_BLIND"
	EndpointDisplayCategoryLaptop               EndpointDisplayCategory = "LAPTOP"
	EndpointDisplayCategoryLight                EndpointDisplayCategory = "LIGHT"
	EndpointDisplayCategoryLightSensor          EndpointDisplayCategory = "LIGHT_SENSOR"
	EndpointDisplayCategoryMicrowave            EndpointDisplayCategory = "MICROWAVE"
	EndpointDisplayCategoryMobilePhone          EndpointDisplayCategory = "MOBILE_PHONE"
	EndpointDisplayCategoryMotionSensor         EndpointDisplayCategory = "MOTION_SENSOR"
	EndpointDisplayCategoryMusicSystem          EndpointDisplayCategory = "MUSIC_SYSTEM"
	EndpointDisplayCategoryNetworkHardware      EndpointDisplayCategory = "NETWORK_HARDWARE"
	EndpointDisplayCategoryOven                 EndpointDisplayCategory = "OVEN"
	EndpointDisplayCategoryPhone                EndpointDisplayCategory = "PHONE"
	EndpointDisplayCategoryPrinter              EndpointDisplayCategory = "PRINTER"
	EndpointDisplayCategoryRefrigerator         EndpointDisplayCategory = "REFRIGERATOR"
	EndpointDisplayCategoryRemote               EndpointDisplayCategory = "REMOTE"
	EndpointDisplayCategoryRouter               EndpointDisplayCategory = "ROUTER"
	EndpointDisplayCategorySceneTrigger         EndpointDisplayCategory = "SCENE_TRIGGER"
	EndpointDisplayCategoryScreen               EndpointDisplayCategory = "SCREEN"
	EndpointDisplayCategorySecurityPanel        EndpointDisplayCategory = "SECURITY_PANEL"
	EndpointDisplayCategorySecuritySystem       EndpointDisplayCategory = "SECURITY_SYSTEM"
	EndpointDisplayCategorySlowCooker           EndpointDisplayCategory = "SLOW_COOKER"
	EndpointDisplayCategorySmartlock            EndpointDisplayCategory = "SMARTLOCK"
	EndpointDisplayCategorySmartplug            EndpointDisplayCategory = "SMARTPLUG"
	EndpointDisplayCategorySmokeSensor          EndpointDisplayCategory = "SMOKE_SENSOR"
	EndpointDisplayCategorySpeaker              EndpointDisplayCategory = "SPEAKER"
	EndpointDisplayCategoryStreamingDevice      EndpointDisplayCategory = "STREAMING_DEVICE"
	EndpointDisplayCategorySwitch               EndpointDisplayCategory = "SWITCH"
	EndpointDisplayCategoryTablet               EndpointDisplayCategory = "TABLET"
	EndpointDisplayCategoryTemperatureSensor    EndpointDisplayCategory = "TEMPERATURE_SENSOR"
	EndpointDisplayCategoryThermostat           EndpointDisplayCategory = "THERMOSTAT"
	EndpointDisplayCategoryTracker              EndpointDisplayCategory = "TRACKER"
	EndpointDisplayCategoryTV                   EndpointDisplayCategory = "TV"
	EndpointDisplayCategoryVacuumCleaner        EndpointDisplayCategory = "VACUUM_CLEANER"
	EndpointDisplayCategoryVehicle              EndpointDisplayCategory = "VEHICLE"
	EndpointDisplayCategoryWasher               EndpointDisplayCategory = "WASHER"
	EndpointDisplayCategoryWaterHeater          EndpointDisplayCategory = "WATER_HEATER"
	EndpointDisplayCategoryWaterLeakSensor      EndpointDisplayCategory = "WATER_LEAK_SENSOR"
	EndpointDisplayCategoryWearable             EndpointDisplayCategory = "WEARABLE"
)
