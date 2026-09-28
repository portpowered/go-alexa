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
	alexamodels "github.com/portpowered/go-alexa/pkg/dependencymodels"
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

	req := alexamodels.RegistrationRequest{
		RequestedTokenType: []string{"bearer"},
		RegistrationData: alexamodels.RegistrationData{
			AppName:      config.AppName,
			AppVersion:   config.AppVersion,
			DeviceType:   config.DeviceType,
			Domain:       config.Domain,
			DeviceModel:  config.DeviceModel,
			OSVersion:    config.OSVersion,
			DeviceSerial: config.DeviceSerial,
			DeviceName:   config.DeviceName,
		},
		AuthData: alexamodels.AuthData{
			EmailPasswordAuth: &alexamodels.EmailPasswordAuth{
				Email:    email,
				Password: password,
			},
		},
	}

	url := c.amazonapiBaseUri + "/auth/register"
	var response alexamodels.RegistrationResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, "POST", url, req, &response); err != nil {
		// Check if it's a challenge response
		if reqErr, ok := err.(*UnauthenticatedRequestError); ok {
			var challengeResp alexamodels.ChallengeResponse
			if jsonErr := json.Unmarshal([]byte(reqErr.Body), &challengeResp); jsonErr == nil && challengeResp.Response.Challenge.ChallengeReason != "" {
				return nil, &RegistrationChallengeError{
					ChallengeReason:              challengeResp.Response.Challenge.ChallengeReason,
					RequiredAuthenticationMethod: challengeResp.Response.Challenge.RequiredAuthenticationMethod,
				}
			}
		}
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to register with email/password",
			Err:     err,
		}
	}

	return &response, nil
}

// RegisterWithCodePair registers a device using code-based linking (CBL)
// POST https://api.amazon.com/auth/register
func (c *Client) RegisterWithCodePair(ctx context.Context, publicCode, privateCode string, config *DeviceRegistrationConfig) (*alexamodels.RegistrationResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := alexamodels.RegistrationRequest{
		RequestedTokenType: []string{"bearer"},
		RegistrationData: alexamodels.RegistrationData{
			AppName:      config.AppName,
			AppVersion:   config.AppVersion,
			DeviceType:   config.DeviceType,
			Domain:       config.Domain,
			DeviceModel:  config.DeviceModel,
			OSVersion:    config.OSVersion,
			DeviceSerial: config.DeviceSerial,
			DeviceName:   config.DeviceName,
		},
		AuthData: alexamodels.AuthData{
			CodePairAuth: &alexamodels.CodePairAuth{
				PublicCode:  publicCode,
				PrivateCode: privateCode,
			},
		},
	}

	url := c.amazonapiBaseUri + "/auth/register"
	var response alexamodels.RegistrationResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, "POST", url, req, &response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to register with code pair",
			Err:     err,
		}
	}

	return &response, nil
}

// GenerateCodePair generates a code pair for code-based linking
// POST https://api.amazon.com/auth/create/codepair
func (c *Client) GenerateCodePair(ctx context.Context, config *DeviceRegistrationConfig) (*alexamodels.CodePairResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := alexamodels.CodePairRequest{
		CodeData: struct {
			AppName               string `json:"app_name"`
			AppVersion            string `json:"app_version"`
			DeviceType            string `json:"device_type"`
			Domain                string `json:"domain"`
			DeviceModel           string `json:"device_model"`
			OSVersion             string `json:"os_version"`
			DeviceSerial          string `json:"device_serial"`
			DeviceName            string `json:"device_name"`
			SecondaryRegistration string `json:"secondary_registration"`
		}{
			AppName:               config.AppName,
			AppVersion:            config.AppVersion,
			DeviceType:            config.DeviceType,
			Domain:                config.Domain,
			DeviceModel:           config.DeviceModel,
			OSVersion:             config.OSVersion,
			DeviceSerial:          config.DeviceSerial,
			DeviceName:            config.DeviceName,
			SecondaryRegistration: "False",
		},
		Scopes: []string{},
	}

	url := c.amazonapiBaseUri + "/auth/create/codepair"
	var response alexamodels.CodePairResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, "POST", url, req, &response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to generate code pair",
			Err:     err,
		}
	}

	return &response, nil
}

// RefreshAccessToken refreshes an access token using a refresh token
// POST https://api.amazon.com/auth/token
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string, config *DeviceRegistrationConfig) (*alexamodels.TokenRefreshResponse, error) {
	if config == nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "device registration config is required",
		}
	}

	req := alexamodels.TokenRefreshRequest{
		AppName:            config.AppName,
		AppVersion:         config.AppVersion,
		SourceTokenType:    "refresh_token",
		SourceToken:        refreshToken,
		RequestedTokenType: "access_token",
	}
	req.DeviceMetadata.DeviceType = config.DeviceType
	req.DeviceMetadata.DeviceModel = config.DeviceModel
	req.DeviceMetadata.OSVersion = config.OSVersion
	req.DeviceMetadata.DeviceSerial = config.DeviceSerial
	req.DeviceMetadata.Manufacturer = config.Manufacturer

	url := c.amazonapiBaseUri + "/auth/token"
	var response alexamodels.TokenRefreshResponse
	if err := c.doUnauthenticatedJSONRequest(ctx, "POST", url, req, &response); err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to refresh access token",
			Err:     err,
		}
	}

	return &response, nil
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

// CookieExchangeRequest represents a request to exchange refresh token for cookies
type CookieExchangeRequest struct {
	AppName            string `json:"app_name"`
	RequestedTokenType string `json:"requested_token_type"`
	Domain             string `json:"domain"`
	SourceTokenType    string `json:"source_token_type"`
	SourceToken        string `json:"source_token"`
}

// CookieExchangeResponse represents the response from cookie exchange
type CookieExchangeResponse struct {
	Response struct {
		Tokens struct {
			Cookies map[string][]struct {
				Name     string `json:"Name"`
				Value    string `json:"Value"`
				Path     string `json:"Path"`
				Secure   bool   `json:"Secure"`
				HttpOnly bool   `json:"HttpOnly"`
				Expires  string `json:"Expires"`
			} `json:"cookies"`
		} `json:"tokens"`
	} `json:"response"`
}

// ExchangeRefreshTokenForCookies exchanges a refresh token for session cookies
// POST https://api.amazon.com/ap/exchangetoken/cookies
func (c *Client) ExchangeRefreshTokenForCookies(ctx context.Context, refreshToken, domain string) (map[string]*http.Cookie, error) {
	req := CookieExchangeRequest{
		AppName:            "Amazon Alexa",
		RequestedTokenType: "auth_cookies",
		Domain:             domain,
		SourceTokenType:    "refresh_token",
		SourceToken:        refreshToken,
	}

	url := "https://api." + domain + "/ap/exchangetoken/cookies"
	var response CookieExchangeResponse

	// Create a temporary client without auth for this unauthenticated request
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, &alexaapimodels.BadRequestError{
			Message: "failed to marshal request",
			Err:     err,
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, &alexaapimodels.NetworkError{
			Message: "failed to create request",
			Err:     err,
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-amzn-identity-auth-domain", "api."+domain)

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
	for domainName, cookieList := range response.Response.Tokens.Cookies {
		for _, cookieData := range cookieList {
			cookie := &http.Cookie{
				Name:     cookieData.Name,
				Value:    cookieData.Value,
				Path:     cookieData.Path,
				Secure:   cookieData.Secure,
				HttpOnly: cookieData.HttpOnly,
			}

			// Parse expiration if provided
			if cookieData.Expires != "" {
				if expTime, err := time.Parse(time.RFC1123, cookieData.Expires); err == nil {
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
