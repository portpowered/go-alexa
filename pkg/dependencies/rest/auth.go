package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	"github.com/portpowered/go-alexa/pkg/dependencies/internal/wire"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

// DeviceRegistrationConfig contains device information for MAP authentication
type DeviceRegistrationConfig struct {
	AppName      string
	AppVersion   string
	DeviceType   string
	Domain       string
	DeviceModel  string
	OSVersion    string
	DeviceSerial string
	DeviceName   string
	Manufacturer string
}

// DefaultDeviceRegistrationConfig returns a default device registration config
func DefaultDeviceRegistrationConfig(deviceSerial, deviceName string) *DeviceRegistrationConfig {
	return &DeviceRegistrationConfig{
		AppName:      "Client SDK",
		AppVersion:   "1.0",
		DeviceType:   alexaapimodels.DeviceTypeSimulator,
		Domain:       "Device",
		DeviceModel:  "Client SDK",
		OSVersion:    "0",
		DeviceSerial: deviceSerial,
		DeviceName:   deviceName,
		Manufacturer: "Amazon",
	}
}

// RegisterWithEmailPassword registers a device using email and password authentication
// POST https://api.amazon.com/auth/register
func (c *Client) RegisterWithEmailPassword(ctx context.Context, email, password string, config *DeviceRegistrationConfig) (*alexamodels.RegistrationResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := wire.WireRegistrationRequest{
		RequestedTokenType: []string{"bearer"},
		RegistrationData: wire.WireRegistrationData{
			AppName:      config.AppName,
			AppVersion:   config.AppVersion,
			DeviceType:   config.DeviceType,
			Domain:       config.Domain,
			DeviceModel:  config.DeviceModel,
			OsVersion:    config.OSVersion,
			DeviceSerial: config.DeviceSerial,
			DeviceName:   config.DeviceName,
		},
		AuthData: wire.WireAuthData{
			EmailPassword: &wire.WireEmailPasswordAuth{
				Email:    email,
				Password: password,
			},
		},
	}

	url := c.amazonapiBaseUri + apiroutes.PathRegisterDevice
	var response wire.WireRegistrationResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodRegisterDevice, url, req, &response); err != nil {
		// Check if it's a challenge response
		if reqErr, ok := err.(*UnauthenticatedRequestError); ok {
			var challengeResp wire.WireChallengeResponse
			if jsonErr := json.Unmarshal([]byte(reqErr.Body), &challengeResp); jsonErr == nil && challengeResp.Response != nil && challengeResp.Response.Challenge != nil && challengeResp.Response.Challenge.ChallengeReason != nil && *challengeResp.Response.Challenge.ChallengeReason != "" {
				reason := valueOrZero(challengeResp.Response.Challenge.ChallengeReason)
				authMethod := valueOrZero(challengeResp.Response.Challenge.RequiredAuthenticationMethod)
				return nil, &RegistrationChallengeError{
					ChallengeReason:              reason,
					RequiredAuthenticationMethod: authMethod,
				}
			}
		}
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to register with email/password",
			Err:     err,
		}
	}

	converted, err := convertWireModel[alexamodels.RegistrationResponse](response)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{Message: "failed to convert registration response", Err: err}
	}
	return converted, nil
}

// RegisterWithCodePair registers a device using code-based linking (CBL)
// POST https://api.amazon.com/auth/register
func (c *Client) RegisterWithCodePair(ctx context.Context, publicCode, privateCode string, config *DeviceRegistrationConfig) (*alexamodels.RegistrationResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := wire.WireRegistrationRequest{
		RequestedTokenType: []string{"bearer"},
		RegistrationData: wire.WireRegistrationData{
			AppName:      config.AppName,
			AppVersion:   config.AppVersion,
			DeviceType:   config.DeviceType,
			Domain:       config.Domain,
			DeviceModel:  config.DeviceModel,
			OsVersion:    config.OSVersion,
			DeviceSerial: config.DeviceSerial,
			DeviceName:   config.DeviceName,
		},
		AuthData: wire.WireAuthData{
			CodePair: &wire.WireCodePairAuth{
				PublicCode:  publicCode,
				PrivateCode: privateCode,
			},
		},
	}

	url := c.amazonapiBaseUri + apiroutes.PathRegisterDevice
	var response wire.WireRegistrationResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodRegisterDevice, url, req, &response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to register with code pair",
			Err:     err,
		}
	}

	converted, err := convertWireModel[alexamodels.RegistrationResponse](response)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{Message: "failed to convert registration response", Err: err}
	}
	return converted, nil
}

// GenerateCodePair generates a code pair for code-based linking
// POST https://api.amazon.com/auth/create/codepair
func (c *Client) GenerateCodePair(ctx context.Context, config *DeviceRegistrationConfig) (*alexamodels.CodePairResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := wire.WireCodePairRequest{
		CodeData: wire.WireCodeData{
			AppName:               config.AppName,
			AppVersion:            config.AppVersion,
			DeviceType:            config.DeviceType,
			Domain:                config.Domain,
			DeviceModel:           config.DeviceModel,
			OsVersion:             config.OSVersion,
			DeviceSerial:          config.DeviceSerial,
			DeviceName:            config.DeviceName,
			SecondaryRegistration: "False",
		},
		Scopes: []string{},
	}

	url := c.amazonapiBaseUri + apiroutes.PathCreateCodePair
	var response wire.WireCodePairResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodCreateCodePair, url, req, &response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to generate code pair",
			Err:     err,
		}
	}

	converted, err := convertWireModel[alexamodels.CodePairResponse](response)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{Message: "failed to convert code-pair response", Err: err}
	}
	return converted, nil
}

// RefreshAccessToken refreshes an access token using a refresh token
// POST https://api.amazon.com/auth/token
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string, config *DeviceRegistrationConfig) (*alexamodels.TokenRefreshResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := wire.WireTokenRefreshRequest{
		AppName:            config.AppName,
		AppVersion:         config.AppVersion,
		SourceTokenType:    "refresh_token",
		SourceToken:        refreshToken,
		RequestedTokenType: "access_token",
	}
	req.DeviceMetadata.DeviceType = config.DeviceType
	req.DeviceMetadata.DeviceModel = config.DeviceModel
	req.DeviceMetadata.OsVersion = config.OSVersion
	req.DeviceMetadata.DeviceSerial = config.DeviceSerial
	req.DeviceMetadata.Manufacturer = config.Manufacturer

	url := c.amazonapiBaseUri + apiroutes.PathRefreshAccessToken
	var response wire.WireTokenRefreshResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodRefreshAccessToken, url, req, &response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to refresh access token",
			Err:     err,
		}
	}

	converted, err := convertWireModel[alexamodels.TokenRefreshResponse](response)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{Message: "failed to convert token response", Err: err}
	}
	return converted, nil
}

// RegistrationChallengeError represents a challenge response from registration
type RegistrationChallengeError struct {
	ChallengeReason              string
	RequiredAuthenticationMethod string
}

func (e *RegistrationChallengeError) Error() string {
	return fmt.Sprintf("registration challenge: %s", e.ChallengeReason)
}

// IsOTPRequired checks if the error indicates OTP is required
func (e *RegistrationChallengeError) IsOTPRequired() bool {
	return e.ChallengeReason == "MissingRequiredAuthenticationData"
}

// IsCBLRequired checks if the error indicates CBL is required
func (e *RegistrationChallengeError) IsCBLRequired() bool {
	return e.ChallengeReason == "HandleOnWebView"
}

// IsAuthenticationFailed checks if authentication failed
func (e *RegistrationChallengeError) IsAuthenticationFailed() bool {
	return e.ChallengeReason == "AuthenticationFailed" && e.RequiredAuthenticationMethod == "GenericClaimPassword"
}

// ExchangeRefreshTokenForCookies exchanges a refresh token for session cookies
// POST https://api.amazon.com/ap/exchangetoken/cookies
func (c *Client) ExchangeRefreshTokenForCookies(ctx context.Context, refreshToken, domain string) (map[string]*http.Cookie, error) {
	req := wire.WireCookieExchangeRequest{
		AppName:            "Amazon Alexa",
		RequestedTokenType: "auth_cookies",
		Domain:             domain,
		SourceTokenType:    "refresh_token",
		SourceToken:        refreshToken,
	}

	url := fmt.Sprintf(apiroutes.ServerExchangeRefreshTokenForCookies, domain) + apiroutes.PathExchangeRefreshTokenForCookies
	var response wire.WireCookieExchangeResponse

	// Create a temporary client without auth for this unauthenticated request
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "failed to marshal request",
			Err:     err,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, apiroutes.MethodExchangeRefreshTokenForCookies, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	httpReq.Header.Set(apiroutes.HeaderContentType, "application/json")
	httpReq.Header.Set(apiroutes.HeaderXAmznIdentityAuthDomain, "api."+domain)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to exchange token for cookies",
			Err:     err,
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &alexaapimodels.NetworkError{
			Message: fmt.Sprintf("cookie exchange failed with status %d: %s", resp.StatusCode, string(bodyBytes)),
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to decode response",
			Err:     err,
		}
	}

	// Convert response cookies to http.Cookie map
	cookies := make(map[string]*http.Cookie)
	if response.Response == nil || response.Response.Tokens == nil || response.Response.Tokens.Cookies == nil {
		return cookies, nil
	}
	for domainName, cookieList := range *response.Response.Tokens.Cookies {
		for _, cookieData := range cookieList {
			cookie := &http.Cookie{
				Name:     valueOrZero(cookieData.Name),
				Value:    valueOrZero(cookieData.Value),
				Path:     valueOrZero(cookieData.Path),
				Secure:   valueOrZero(cookieData.Secure),
				HttpOnly: valueOrZero(cookieData.HttpOnly),
			}

			// Parse expiration if provided
			expires := valueOrZero(cookieData.Expires)
			if expires != "" {
				if expTime, err := time.Parse(time.RFC1123, expires); err == nil {
					cookie.Expires = expTime
				}
			}

			// Use domain name as key (remove leading dot if present)
			domainKey := domainName
			if len(domainKey) > 0 && domainKey[0] == '.' {
				domainKey = domainKey[1:]
			}
			cookies[domainKey+":"+cookie.Name] = cookie
		}
	}

	return cookies, nil
}
