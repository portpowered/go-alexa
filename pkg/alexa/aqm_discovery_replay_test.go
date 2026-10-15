package alexa_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/portpowered/go-alexa/pkg/alexa"
	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	replay "github.com/portpowered/go-alexa/pkg/testing"
)

func TestListEndpoints_AirQualityFallbackPreservesRangeConfiguration(t *testing.T) {
	t.Parallel()

	fixture := filepath.Join("..", "..", "tests", "replay", "fixtures", "synthetic", "alexa-aqm-discovery.json")

	player, err := replay.LoadSyntheticReplay(fixture)
	if err != nil {
		t.Fatal(err)
	}

	client, err := alexa.NewClient(alexa.WithGraphQLBaseURL("https://graphql.synthetic.test"),
		alexa.WithAlexaWebBaseURL("https://web.synthetic.test"),
		alexa.WithHTTPClient(&http.Client{Transport: player}))
	if err != nil {
		t.Fatal(err)
	}

	session, err := client.NewSession(alexa.WithBearerToken("synthetic-graphql-token"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		closeErr := session.Close()
		if closeErr != nil {
			t.Error(closeErr)
		}
	})

	response, err := session.ListEndpoints(context.Background(), alexaapimodels.EndpointQuery{})
	if err != nil {
		t.Fatal(err)
	}

	err = player.AssertConsumed()
	if err != nil {
		t.Fatal(err)
	}

	if len(response.Results) != 1 || response.Results[0].EndpointID != "synthetic-aqm" {
		t.Fatalf("fallback results = %#v", response.Results)
	}

	if response.Results[0].DisplayCategories.Primary != alexaapimodels.EndpointDisplayCategoryAirQualityMonitor {
		t.Fatalf("fallback category = %s", response.Results[0].DisplayCategories.Primary)
	}

	for _, feature := range response.Results[0].Features {
		if feature.Name == alexaapimodels.FeatureNameRange && feature.Instance == "Air Quality.pm25" {
			assertFallbackRangeConfig(t, feature.Config)

			return
		}
	}

	t.Fatalf("fallback range instance missing: %#v", response.Results[0].Features)
}

func assertFallbackRangeConfig(t *testing.T, config *alexaapimodels.FeatureConfig) {
	t.Helper()

	if config == nil || config.Range == nil {
		t.Fatal("fallback discarded range configuration")
	}

	rangeConfig := config.Range
	if rangeConfig.FriendlyName == nil || rangeConfig.FriendlyName.Value != "PM2.5" || rangeConfig.FriendlyName.Type != "PLAIN" {
		t.Errorf("friendly name = %#v", rangeConfig.FriendlyName)
	}

	if rangeConfig.UnitOfMeasure == nil || rangeConfig.UnitOfMeasure.Value != "ug/m3" || rangeConfig.UnitOfMeasure.Type != "PLAIN" {
		t.Errorf("unit of measure = %#v", rangeConfig.UnitOfMeasure)
	}

	if rangeConfig.SupportedRange == nil || *rangeConfig.SupportedRange != (alexaapimodels.SupportedRange{
		MinimumValue: 0,
		MaximumValue: 1000,
		Precision:    0.1,
	}) {
		t.Errorf("supported range = %#v", rangeConfig.SupportedRange)
	}

	if len(rangeConfig.Presets) != 1 || rangeConfig.Presets[0].RangeValue != 5 ||
		rangeConfig.Presets[0].FriendlyName == nil || rangeConfig.Presets[0].FriendlyName.Value != "Low" {
		t.Errorf("presets = %#v", rangeConfig.Presets)
	}
}
