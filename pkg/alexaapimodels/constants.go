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
