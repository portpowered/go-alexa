// Package alexamodels provides dependency API request and response models.
package alexamodels

// UserInfo represents user information from the /api/users/me endpoint.
type UserInfo struct {
	// This is the country code for country of residence for the customer.
	CountryOfResidence string `json:"countryOfResidence"`
	// This is the amazon marketplace ID for the user.
	EffectiveMarketPlaceID string   `json:"effectiveMarketPlaceId"`
	Email                  string   `json:"email"`
	EulaAcceptance         bool     `json:"eulaAcceptance"`
	Features               []string `json:"features"`
	FullName               string   `json:"fullName"`
	HasActiveDopplers      bool     `json:"hasActiveDopplers"`
	// This ID is the user's unique identifier in the amazon account system.
	ID                    string `json:"id"`
	MarketPlaceDomainName string `json:"marketPlaceDomainName"`
	MarketPlaceID         string `json:"marketPlaceId"`
	MarketPlaceLocale     string `json:"marketPlaceLocale"`
}
