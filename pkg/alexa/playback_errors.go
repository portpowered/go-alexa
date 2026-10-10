package alexa

import (
	"context"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
)

const (
	playbackDispatchFireTV           = "fire_tv"
	playbackDispatchAlexaMedia       = "alexa_media"
	playbackDispatchInterfaceMessage = "interface_message"
)

func playbackDispatchPath(endpoint alexaapimodels.EndpointInterface) string {
	if endpoint.GetDeviceFamily() == alexamodels.DeviceFamilyFireTV {
		return playbackDispatchFireTV
	}

	if endpoint.GetDeviceSerialNumber() != "" {
		return playbackDispatchAlexaMedia
	}

	return playbackDispatchInterfaceMessage
}

func (c *Session) controlPlaybackWithContext(ctx context.Context, request alexaapimodels.ControlRequest) (*alexaapimodels.ControlResponse, error) {
	response, err := c.controlPlayback(ctx, request)
	if err != nil {
		return response, &alexaapimodels.PlaybackDispatchError{
			DispatchPath: playbackDispatchPath(request.Target), Operation: request.Name, Err: err,
		}
	}

	return response, nil
}
