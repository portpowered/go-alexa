package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/portpowered/go-alexa/pkg/alexaapimodels"
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
	"github.com/portpowered/go-alexa/pkg/internal/apiroutes"
)

// DeviceRegistrationConfig contains device information for MAP authentication.
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

// DefaultDeviceRegistrationConfig returns a default device registration config.
func DefaultDeviceRegistrationConfig(deviceSerial, deviceName string) *DeviceRegistrationConfig {
	return &DeviceRegistrationConfig{
		AppName:      alexaapimodels.DefaultAppName,
		AppVersion:   alexaapimodels.DefaultAppVersion,
		DeviceType:   alexaapimodels.DeviceTypeSimulator,
		Domain:       alexaapimodels.DefaultRegistrationDomain,
		DeviceModel:  alexaapimodels.DefaultDeviceModel,
		OSVersion:    alexaapimodels.DefaultOSVersion,
		DeviceSerial: deviceSerial,
		DeviceName:   deviceName,
		Manufacturer: alexaapimodels.DefaultManufacturer,
	}
}

// RegisterWithEmailPassword registers a device using email and password authentication
// POST https://api.amazon.com/auth/register
func (c *Client) RegisterWithEmailPassword(
	ctx context.Context,
	email string,
	password string,
	config *DeviceRegistrationConfig,
) (*alexamodels.RegistrationResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := alexamodels.WireRegistrationRequest{
		RequestedTokenType: []string{alexamodels.AuthTokenTypeBearer},
		RegistrationData: alexamodels.WireRegistrationData{
			AppName:      config.AppName,
			AppVersion:   config.AppVersion,
			DeviceType:   config.DeviceType,
			Domain:       config.Domain,
			DeviceModel:  config.DeviceModel,
			OsVersion:    config.OSVersion,
			DeviceSerial: config.DeviceSerial,
			DeviceName:   config.DeviceName,
		},
		AuthData: alexamodels.WireAuthData{
			CodePair: nil,
			EmailPassword: &alexamodels.WireEmailPasswordAuth{
				Email:    email,
				Password: password,
			},
		},
	}

	url := c.amazonapiBaseURI + apiroutes.PathRegisterDevice

	var response alexamodels.WireRegistrationResponse

	err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodRegisterDevice, url, req, &response)
	if err != nil {
		challenge, foundChallenge, challengeParseErr := registrationChallengeFromError(err)
		if challengeParseErr == nil && foundChallenge {
			return nil, &challenge
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
func (c *Client) RegisterWithCodePair(
	ctx context.Context,
	publicCode string,
	privateCode string,
	config *DeviceRegistrationConfig,
) (*alexamodels.RegistrationResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := alexamodels.WireRegistrationRequest{
		RequestedTokenType: []string{alexamodels.AuthTokenTypeBearer},
		RegistrationData: alexamodels.WireRegistrationData{
			AppName:      config.AppName,
			AppVersion:   config.AppVersion,
			DeviceType:   config.DeviceType,
			Domain:       config.Domain,
			DeviceModel:  config.DeviceModel,
			OsVersion:    config.OSVersion,
			DeviceSerial: config.DeviceSerial,
			DeviceName:   config.DeviceName,
		},
		AuthData: alexamodels.WireAuthData{
			EmailPassword: nil,
			CodePair: &alexamodels.WireCodePairAuth{
				PublicCode:  publicCode,
				PrivateCode: privateCode,
			},
		},
	}

	url := c.amazonapiBaseURI + apiroutes.PathRegisterDevice

	var response alexamodels.WireRegistrationResponse

	{
		err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodRegisterDevice, url, req, &response)
		if err != nil {
			return nil, &alexaapimodels.NetworkError{
				Message: "failed to register with code pair",
				Err:     err,
			}
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

	req := alexamodels.WireCodePairRequest{
		CodeData: alexamodels.WireCodeData{
			AppName:               config.AppName,
			AppVersion:            config.AppVersion,
			DeviceType:            config.DeviceType,
			Domain:                config.Domain,
			DeviceModel:           config.DeviceModel,
			OsVersion:             config.OSVersion,
			DeviceSerial:          config.DeviceSerial,
			DeviceName:            config.DeviceName,
			SecondaryRegistration: alexamodels.AuthSecondaryRegistrationFalse,
		},
		Scopes: []string{},
	}

	url := c.amazonapiBaseURI + apiroutes.PathCreateCodePair

	var response alexamodels.WireCodePairResponse

	{
		err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodCreateCodePair, url, req, &response)
		if err != nil {
			return nil, &alexaapimodels.NetworkError{
				Message: "failed to generate code pair",
				Err:     err,
			}
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

	req := alexamodels.WireTokenRefreshRequest{
		AppName:            config.AppName,
		AppVersion:         config.AppVersion,
		SourceTokenType:    alexamodels.AuthTokenTypeRefresh,
		SourceToken:        refreshToken,
		RequestedTokenType: alexamodels.AuthTokenTypeAccess,
		DeviceMetadata: alexamodels.WireDeviceMetadata{
			DeviceType: config.DeviceType, DeviceModel: config.DeviceModel,
			OsVersion: config.OSVersion, DeviceSerial: config.DeviceSerial,
			Manufacturer: config.Manufacturer,
		},
	}

	url := c.amazonapiBaseURI + apiroutes.PathRefreshAccessToken

	var response alexamodels.WireTokenRefreshResponse

	{
		err := c.doUnauthenticatedJSONRequest(ctx, apiroutes.MethodRefreshAccessToken, url, req, &response)
		if err != nil {
			return nil, &alexaapimodels.NetworkError{
				Message: "failed to refresh access token",
				Err:     err,
			}
		}
	}

	converted, err := convertWireModel[alexamodels.TokenRefreshResponse](response)
	if err != nil {
		return nil, &alexaapimodels.NetworkError{Message: "failed to convert token response", Err: err}
	}

	return converted, nil
}

// RegistrationChallengeError represents a challenge response from registration.
type RegistrationChallengeError struct {
	ChallengeReason              string
	RequiredAuthenticationMethod string
}

func (e *RegistrationChallengeError) Error() string {
	return "registration challenge: " + e.ChallengeReason
}

// IsOTPRequired checks if the error indicates OTP is required.
func (e *RegistrationChallengeError) IsOTPRequired() bool {
	return e.ChallengeReason == alexamodels.AuthChallengeMissingData
}

// IsCBLRequired checks if the error indicates CBL is required.
func (e *RegistrationChallengeError) IsCBLRequired() bool {
	return e.ChallengeReason == alexamodels.AuthChallengeWebView
}

// IsAuthenticationFailed checks if authentication failed.
func (e *RegistrationChallengeError) IsAuthenticationFailed() bool {
	return e.ChallengeReason == alexamodels.AuthChallengeFailed && e.RequiredAuthenticationMethod == alexamodels.AuthChallengePasswordMethod
}

// ExchangeRefreshTokenForCookies exchanges a refresh token for session cookies
// POST https://api.amazon.com/ap/exchangetoken/cookies
func (c *Client) ExchangeRefreshTokenForCookies(ctx context.Context, refreshToken, domain string) (map[string]*http.Cookie, error) {
	req := alexamodels.WireCookieExchangeRequest{
		AppName:            alexamodels.AuthCookieExchangeAppName,
		RequestedTokenType: alexamodels.AuthTokenTypeCookies,
		Domain:             domain,
		SourceTokenType:    alexamodels.AuthTokenTypeRefresh,
		SourceToken:        refreshToken,
	}

	url := fmt.Sprintf(apiroutes.ServerExchangeRefreshTokenForCookies, domain) + apiroutes.PathExchangeRefreshTokenForCookies

	var response alexamodels.WireCookieExchangeResponse

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

	{
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, &alexaapimodels.NetworkError{
				Message: "failed to decode response",
				Err:     err,
			}
		}
	}

	// Convert response cookies to http.Cookie map
	cookies := make(map[string]*http.Cookie)
	if response.Response == nil || response.Response.Tokens == nil || response.Response.Tokens.Cookies == nil {
		return cookies, nil
	}

	for domainName, cookieList := range *response.Response.Tokens.Cookies {
		for _, cookieData := range cookieList {
			cookie := &http.Cookie{}
			assignOptional(&cookie.Name, cookieData.Name)
			assignOptional(&cookie.Value, cookieData.Value)
			assignOptional(&cookie.Path, cookieData.Path)
			assignOptional(&cookie.Secure, cookieData.Secure)
			assignOptional(&cookie.HttpOnly, cookieData.HttpOnly)

			// Parse expiration if provided
			expires := ""

			assignOptional(&expires, cookieData.Expires)

			if expires != "" {
				{
					expTime, err := time.Parse(time.RFC1123, expires)
					if err == nil {
						cookie.Expires = expTime
					}
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

func registrationChallengeFromError(err error) (RegistrationChallengeError, bool, error) {
	requestErr := &UnauthenticatedRequestError{}
	if !errors.As(err, &requestErr) {
		return RegistrationChallengeError{ChallengeReason: "", RequiredAuthenticationMethod: ""}, false, nil
	}

	var response alexamodels.WireChallengeResponse

	parseErr := json.Unmarshal([]byte(requestErr.Body), &response)
	if parseErr != nil {
		return RegistrationChallengeError{}, false, fmt.Errorf("decode registration challenge: %w", parseErr)
	}

	if response.Response == nil {
		return RegistrationChallengeError{ChallengeReason: "", RequiredAuthenticationMethod: ""}, false, nil
	}

	challenge := response.Response.Challenge
	if challenge == nil || challenge.ChallengeReason == nil || *challenge.ChallengeReason == "" {
		return RegistrationChallengeError{ChallengeReason: "", RequiredAuthenticationMethod: ""}, false, nil
	}

	challengeError := RegistrationChallengeError{
		ChallengeReason:              "",
		RequiredAuthenticationMethod: "",
	}
	assignOptional(&challengeError.ChallengeReason, challenge.ChallengeReason)
	assignOptional(&challengeError.RequiredAuthenticationMethod, challenge.RequiredAuthenticationMethod)

	return challengeError, true, nil
}
