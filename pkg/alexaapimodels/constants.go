package alexaapimodels

import "github.com/portpowered/go-alexa/pkg/internal/apiroutes"

const (
	// AmazonApiServiceUriNa is the Amazon API origin for North America.
	AmazonApiServiceUriNa = apiroutes.ServerAmazonAPIBaseNa
	// AmazonApiServiceUriEu is the Amazon API origin for Europe.
	AmazonApiServiceUriEu = apiroutes.ServerAmazonAPIBaseEu
	// AmazonApiServiceUriJp is the Amazon API origin for Japan.
	AmazonApiServiceUriJp = apiroutes.ServerAmazonAPIBaseJp

	// AuthorizationUriNa is the authorization page URL for North America.
	AuthorizationUriNa = AmazonApiServiceUriNa + apiroutes.PathOpenAuthorizationPage
	// AuthorizationUriEu is the authorization page URL for Europe.
	AuthorizationUriEu = AmazonApiServiceUriEu + apiroutes.PathOpenAuthorizationPage
	// AuthorizationUriJp is the authorization page URL for Japan.
	AuthorizationUriJp = AmazonApiServiceUriJp + apiroutes.PathOpenAuthorizationPage

	// ApiServiceUriNa is the Alexa API origin for North America.
	ApiServiceUriNa = apiroutes.ServerAlexaAPIBaseNa
	// ApiServiceUriEu is the Alexa API origin for Europe.
	ApiServiceUriEu = apiroutes.ServerAlexaAPIBaseEu
	// ApiServiceUriJp is the Alexa API origin for Japan.
	ApiServiceUriJp = apiroutes.ServerAlexaAPIBaseJp

	// AlexaAmazonBaseUriNa is the Alexa web origin for North America.
	AlexaAmazonBaseUriNa = apiroutes.ServerAlexaWebBaseNa
	// AlexaAmazonBaseUriEu is the Alexa web origin for Europe.
	AlexaAmazonBaseUriEu = apiroutes.ServerAlexaWebBaseEu
	// AlexaAmazonBaseUriJp is the Alexa web origin for Japan.
	AlexaAmazonBaseUriJp = apiroutes.ServerAlexaWebBaseJp

	// Http2ConnectionUriNa is the Alexa event endpoint for North America.
	// https://developer.amazon.com/en-US/docs/alexa/alexa-voice-service/api-overview.html#endpoints
	Http2ConnectionUriNa = apiroutes.ServerEventAuthorityNa
	// Http2ConnectionUriEu is the Alexa event endpoint for Europe.
	Http2ConnectionUriEu = apiroutes.ServerEventAuthorityEu
	// Http2ConnectionUriJp is the Alexa event endpoint for Japan.
	Http2ConnectionUriJp = apiroutes.ServerEventAuthorityJp

	// DeviceTypeIphone identifies the Alexa mobile app device type.
	DeviceTypeIphone = "A2IVLV5VM2W81"
	// DeviceTypeSimulator identifies the Alexa simulator device type.
	DeviceTypeSimulator = "A3UMCOIO3URGGU"

	// EventNameColorTemperatureState is the event name for color-temperature updates.
	EventNameColorTemperatureState = "colorTemperatureState"
	// EventNamePowerState is the event name for power-state updates.
	EventNamePowerState = "powerstate"
	// EventNameSpeakerState is the event name for speaker-state updates.
	EventNameSpeakerState = "speakerState"
	// EventNameBrightnessState is the event name for brightness updates.
	EventNameBrightnessState = "brightnessState"
	// EventNameColorState is the event name for color updates.
	EventNameColorState = "colorState"
	// EventNameLockState is the event name for lock-state updates.
	EventNameLockState = "lockState"
	// EventNameModeState is the event name for mode-state updates.
	EventNameModeState = "modeState"
	// EventNameRangeState is the event name for range-state updates.
	EventNameRangeState = "rangeState"
	// EventNameToggleState is the event name for the corresponding state update.
	EventNameToggleState = "toggleState"
	// EventNamePercentageState is the event name for the corresponding state update.
	EventNamePercentageState = "percentageState"
	// EventNamePowerLevelState is the event name for the corresponding state update.
	EventNamePowerLevelState = "powerLevelState"
	// EventNameThermostatModeState is the event name for the corresponding state update.
	EventNameThermostatModeState = "thermostatModeState"
	// EventNameSetpointState is the event name for the corresponding state update.
	EventNameSetpointState = "setpointState"
	// EventNameTemperatureState is the event name for the corresponding state update.
	EventNameTemperatureState = "temperatureState"
	// EventNameDetectionState is the event name for the corresponding state update.
	EventNameDetectionState = "detectionState"
	// EventNameActionState is the event name for the corresponding state update.
	EventNameActionState = "actionState"
	// EventNameReachabilityState is the event name for the corresponding state update.
	EventNameReachabilityState = "reachabilityState"
	// EventNameBatteryState is the event name for the corresponding state update.
	EventNameBatteryState = "batteryState"
	// EventNameIlluminanceState is the event name for the corresponding state update.
	EventNameIlluminanceState = "illuminanceState"
	// EventNameGeolocationState is the event name for the corresponding state update.
	EventNameGeolocationState = "geolocationState"
	// EventNameStatusCodeState is the event name for the corresponding state update.
	EventNameStatusCodeState = "statusCodeState"
	// EventNameArmState is the event name for the corresponding state update.
	EventNameArmState = "armState"
	// EventNameRelativeHumidityState is the event name for the corresponding state update.
	EventNameRelativeHumidityState = "relativeHumidityState"
)

// EndpointDisplayCategory names a category that describes an endpoint.
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

// ValidDisplayCategory reports whether s is a supported endpoint category.
func ValidDisplayCategory(s string) bool {
	_, ok := validEndpointDisplayCategory[EndpointDisplayCategory(s)]

	return ok
}

const (
	// EndpointDisplayCategoryOther is the fallback endpoint category.
	EndpointDisplayCategoryOther EndpointDisplayCategory = "OTHER"
	// EndpointDisplayCategoryActivityTrigger marks an activity-trigger endpoint.
	EndpointDisplayCategoryActivityTrigger EndpointDisplayCategory = "ACTIVITY_TRIGGER"
	// EndpointDisplayCategoryAirConditioner marks an air-conditioner endpoint.
	EndpointDisplayCategoryAirConditioner EndpointDisplayCategory = "AIR_CONDITIONER"
	// EndpointDisplayCategoryAirFreshener marks an air-freshener endpoint.
	EndpointDisplayCategoryAirFreshener EndpointDisplayCategory = "AIR_FRESHENER"
	// EndpointDisplayCategoryAirPurifier marks an air-purifier endpoint.
	EndpointDisplayCategoryAirPurifier EndpointDisplayCategory = "AIR_PURIFIER"
	// EndpointDisplayCategoryAirQualityMonitor marks an air-quality-monitor endpoint.
	EndpointDisplayCategoryAirQualityMonitor EndpointDisplayCategory = "AIR_QUALITY_MONITOR"
	// EndpointDisplayCategoryAlexaVoiceEnabled marks a voice-enabled endpoint.
	EndpointDisplayCategoryAlexaVoiceEnabled EndpointDisplayCategory = "ALEXA_VOICE_ENABLED"
	// EndpointDisplayCategoryAutoAccessory marks an automotive accessory endpoint.
	EndpointDisplayCategoryAutoAccessory EndpointDisplayCategory = "AUTO_ACCESSORY"
	// EndpointDisplayCategoryBluetoothSpeaker marks a Bluetooth speaker endpoint.
	EndpointDisplayCategoryBluetoothSpeaker EndpointDisplayCategory = "BLUETOOTH_SPEAKER"
	// EndpointDisplayCategoryBurner identifies this supported endpoint display category.
	EndpointDisplayCategoryBurner EndpointDisplayCategory = "BURNER"
	// EndpointDisplayCategoryCamera identifies this supported endpoint display category.
	EndpointDisplayCategoryCamera EndpointDisplayCategory = "CAMERA"
	// EndpointDisplayCategoryCarbonMonoxideSensor identifies this supported endpoint display category.
	EndpointDisplayCategoryCarbonMonoxideSensor EndpointDisplayCategory = "CARBON_MONOXIDE_SENSOR"
	// EndpointDisplayCategoryChristmasTree identifies this supported endpoint display category.
	EndpointDisplayCategoryChristmasTree EndpointDisplayCategory = "CHRISTMAS_TREE"
	// EndpointDisplayCategoryCoffeeMaker identifies this supported endpoint display category.
	EndpointDisplayCategoryCoffeeMaker EndpointDisplayCategory = "COFFEE_MAKER"
	// EndpointDisplayCategoryComputer identifies this supported endpoint display category.
	EndpointDisplayCategoryComputer EndpointDisplayCategory = "COMPUTER"
	// EndpointDisplayCategoryContactSensor identifies this supported endpoint display category.
	EndpointDisplayCategoryContactSensor EndpointDisplayCategory = "CONTACT_SENSOR"
	// EndpointDisplayCategoryCookingAppliance identifies this supported endpoint display category.
	EndpointDisplayCategoryCookingAppliance EndpointDisplayCategory = "COOKING_APPLIANCE"
	// EndpointDisplayCategoryCooktop identifies this supported endpoint display category.
	EndpointDisplayCategoryCooktop EndpointDisplayCategory = "COOKTOP"
	// EndpointDisplayCategoryCoolingCabinet identifies this supported endpoint display category.
	EndpointDisplayCategoryCoolingCabinet EndpointDisplayCategory = "COOLING_CABINET"
	// EndpointDisplayCategoryDishwasher identifies this supported endpoint display category.
	EndpointDisplayCategoryDishwasher EndpointDisplayCategory = "DISHWASHER"
	// EndpointDisplayCategoryDoor identifies this supported endpoint display category.
	EndpointDisplayCategoryDoor EndpointDisplayCategory = "DOOR"
	// EndpointDisplayCategoryDoorbell identifies this supported endpoint display category.
	EndpointDisplayCategoryDoorbell EndpointDisplayCategory = "DOORBELL"
	// EndpointDisplayCategoryDryer identifies this supported endpoint display category.
	EndpointDisplayCategoryDryer EndpointDisplayCategory = "DRYER"
	// EndpointDisplayCategoryExteriorBlind identifies this supported endpoint display category.
	EndpointDisplayCategoryExteriorBlind EndpointDisplayCategory = "EXTERIOR_BLIND"
	// EndpointDisplayCategoryFan identifies this supported endpoint display category.
	EndpointDisplayCategoryFan EndpointDisplayCategory = "FAN"
	// EndpointDisplayCategoryGameConsole identifies this supported endpoint display category.
	EndpointDisplayCategoryGameConsole EndpointDisplayCategory = "GAME_CONSOLE"
	// EndpointDisplayCategoryGarageDoor identifies this supported endpoint display category.
	EndpointDisplayCategoryGarageDoor EndpointDisplayCategory = "GARAGE_DOOR"
	// EndpointDisplayCategoryHeadphones identifies this supported endpoint display category.
	EndpointDisplayCategoryHeadphones EndpointDisplayCategory = "HEADPHONES"
	// EndpointDisplayCategoryHub identifies this supported endpoint display category.
	EndpointDisplayCategoryHub EndpointDisplayCategory = "HUB"
	// EndpointDisplayCategoryHumiditySensor identifies this supported endpoint display category.
	EndpointDisplayCategoryHumiditySensor EndpointDisplayCategory = "HUMIDITY_SENSOR"
	// EndpointDisplayCategoryInteriorBlind identifies this supported endpoint display category.
	EndpointDisplayCategoryInteriorBlind EndpointDisplayCategory = "INTERIOR_BLIND"
	// EndpointDisplayCategoryLaptop identifies this supported endpoint display category.
	EndpointDisplayCategoryLaptop EndpointDisplayCategory = "LAPTOP"
	// EndpointDisplayCategoryLight identifies this supported endpoint display category.
	EndpointDisplayCategoryLight EndpointDisplayCategory = "LIGHT"
	// EndpointDisplayCategoryLightSensor identifies this supported endpoint display category.
	EndpointDisplayCategoryLightSensor EndpointDisplayCategory = "LIGHT_SENSOR"
	// EndpointDisplayCategoryMicrowave identifies this supported endpoint display category.
	EndpointDisplayCategoryMicrowave EndpointDisplayCategory = "MICROWAVE"
	// EndpointDisplayCategoryMobilePhone identifies this supported endpoint display category.
	EndpointDisplayCategoryMobilePhone EndpointDisplayCategory = "MOBILE_PHONE"
	// EndpointDisplayCategoryMotionSensor identifies this supported endpoint display category.
	EndpointDisplayCategoryMotionSensor EndpointDisplayCategory = "MOTION_SENSOR"
	// EndpointDisplayCategoryMusicSystem identifies this supported endpoint display category.
	EndpointDisplayCategoryMusicSystem EndpointDisplayCategory = "MUSIC_SYSTEM"
	// EndpointDisplayCategoryNetworkHardware identifies this supported endpoint display category.
	EndpointDisplayCategoryNetworkHardware EndpointDisplayCategory = "NETWORK_HARDWARE"
	// EndpointDisplayCategoryOven identifies this supported endpoint display category.
	EndpointDisplayCategoryOven EndpointDisplayCategory = "OVEN"
	// EndpointDisplayCategoryPhone identifies this supported endpoint display category.
	EndpointDisplayCategoryPhone EndpointDisplayCategory = "PHONE"
	// EndpointDisplayCategoryPrinter identifies this supported endpoint display category.
	EndpointDisplayCategoryPrinter EndpointDisplayCategory = "PRINTER"
	// EndpointDisplayCategoryRefrigerator identifies this supported endpoint display category.
	EndpointDisplayCategoryRefrigerator EndpointDisplayCategory = "REFRIGERATOR"
	// EndpointDisplayCategoryRemote identifies this supported endpoint display category.
	EndpointDisplayCategoryRemote EndpointDisplayCategory = "REMOTE"
	// EndpointDisplayCategoryRouter identifies this supported endpoint display category.
	EndpointDisplayCategoryRouter EndpointDisplayCategory = "ROUTER"
	// EndpointDisplayCategorySceneTrigger identifies this supported endpoint display category.
	EndpointDisplayCategorySceneTrigger EndpointDisplayCategory = "SCENE_TRIGGER"
	// EndpointDisplayCategoryScreen identifies this supported endpoint display category.
	EndpointDisplayCategoryScreen EndpointDisplayCategory = "SCREEN"
	// EndpointDisplayCategorySecurityPanel identifies this supported endpoint display category.
	EndpointDisplayCategorySecurityPanel EndpointDisplayCategory = "SECURITY_PANEL"
	// EndpointDisplayCategorySecuritySystem identifies this supported endpoint display category.
	EndpointDisplayCategorySecuritySystem EndpointDisplayCategory = "SECURITY_SYSTEM"
	// EndpointDisplayCategorySlowCooker identifies this supported endpoint display category.
	EndpointDisplayCategorySlowCooker EndpointDisplayCategory = "SLOW_COOKER"
	// EndpointDisplayCategorySmartlock identifies this supported endpoint display category.
	EndpointDisplayCategorySmartlock EndpointDisplayCategory = "SMARTLOCK"
	// EndpointDisplayCategorySmartplug identifies this supported endpoint display category.
	EndpointDisplayCategorySmartplug EndpointDisplayCategory = "SMARTPLUG"
	// EndpointDisplayCategorySmokeSensor identifies this supported endpoint display category.
	EndpointDisplayCategorySmokeSensor EndpointDisplayCategory = "SMOKE_SENSOR"
	// EndpointDisplayCategorySpeaker identifies this supported endpoint display category.
	EndpointDisplayCategorySpeaker EndpointDisplayCategory = "SPEAKER"
	// EndpointDisplayCategoryStreamingDevice identifies this supported endpoint display category.
	EndpointDisplayCategoryStreamingDevice EndpointDisplayCategory = "STREAMING_DEVICE"
	// EndpointDisplayCategorySwitch identifies this supported endpoint display category.
	EndpointDisplayCategorySwitch EndpointDisplayCategory = "SWITCH"
	// EndpointDisplayCategoryTablet identifies this supported endpoint display category.
	EndpointDisplayCategoryTablet EndpointDisplayCategory = "TABLET"
	// EndpointDisplayCategoryTemperatureSensor identifies this supported endpoint display category.
	EndpointDisplayCategoryTemperatureSensor EndpointDisplayCategory = "TEMPERATURE_SENSOR"
	// EndpointDisplayCategoryThermostat identifies this supported endpoint display category.
	EndpointDisplayCategoryThermostat EndpointDisplayCategory = "THERMOSTAT"
	// EndpointDisplayCategoryTracker identifies this supported endpoint display category.
	EndpointDisplayCategoryTracker EndpointDisplayCategory = "TRACKER"
	// EndpointDisplayCategoryTV identifies this supported endpoint display category.
	EndpointDisplayCategoryTV EndpointDisplayCategory = "TV"
	// EndpointDisplayCategoryVacuumCleaner identifies this supported endpoint display category.
	EndpointDisplayCategoryVacuumCleaner EndpointDisplayCategory = "VACUUM_CLEANER"
	// EndpointDisplayCategoryVehicle identifies this supported endpoint display category.
	EndpointDisplayCategoryVehicle EndpointDisplayCategory = "VEHICLE"
	// EndpointDisplayCategoryWasher identifies this supported endpoint display category.
	EndpointDisplayCategoryWasher EndpointDisplayCategory = "WASHER"
	// EndpointDisplayCategoryWaterHeater identifies this supported endpoint display category.
	EndpointDisplayCategoryWaterHeater EndpointDisplayCategory = "WATER_HEATER"
	// EndpointDisplayCategoryWaterLeakSensor identifies this supported endpoint display category.
	EndpointDisplayCategoryWaterLeakSensor EndpointDisplayCategory = "WATER_LEAK_SENSOR"
	// EndpointDisplayCategoryWearable identifies this supported endpoint display category.
	EndpointDisplayCategoryWearable EndpointDisplayCategory = "WEARABLE"
)
