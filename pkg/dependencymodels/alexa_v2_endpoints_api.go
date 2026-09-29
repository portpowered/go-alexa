package alexamodels

// EndpointListResponse represents the paginated response from endpoint enumeration.
type EndpointListResponse struct {
	Results   []Device `json:"results"`
	NextToken string   `json:"nextToken,omitempty"`
}

// EndpointQueryRequest represents an advanced search query for endpoints.
type EndpointQueryRequest struct {
	Query EndpointQuery `json:"query"`
}

// EndpointQuery represents the query criteria for advanced search.
type EndpointQuery struct {
	AND []EndpointQueryClause `json:"and,omitempty"`
	OR  []EndpointQueryClause `json:"or,omitempty"`

	// IncludeFields specifies which additional fields to include in the response
	IncludeFields     *EndpointIncludeFields `json:"includeFields,omitempty"`
	PaginationContext *PaginationContext     `json:"paginationContext,omitempty"`
}

// EndpointIncludeFields specifies which optional fields to include in the response.
type EndpointIncludeFields struct {
	States       bool `json:"states,omitempty"`       // Include device states from REST API
	Capabilities bool `json:"capabilities,omitempty"` // Include capabilities/interfaces from REST API
}

// PaginationContext represents the pagination context for the query.
type PaginationContext struct {
	NextToken  string `json:"nextToken,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`

	// Note: the default behavior is false, so we grab everything.
	PerformPagination bool `json:"performPagination,omitempty"`
}

// EndpointQueryClause represents a single query clause.
type EndpointQueryClause struct {
	AssociatedUnits *AssociatedUnitsFilter `json:"associatedUnits,omitempty"`
	Manufacturer    *StringFilter          `json:"manufacturer,omitempty"`
	Model           *StringFilter          `json:"model,omitempty"`
}

// AssociatedUnitsFilter filters by associated units.
type AssociatedUnitsFilter struct {
	ID string `json:"id"`
}

// StringFilter filters by string value.
type StringFilter struct {
	Value struct {
		Text string `json:"text"`
	} `json:"value"`
}

// FriendlyNameRequest represents a request to update an endpoint's friendly name.
type FriendlyNameRequest struct {
	FriendlyName FriendlyNameValue `json:"friendlyName"`
}

// FriendlyNameValue represents the friendly name value structure.
type FriendlyNameValue struct {
	Type  string `json:"type"` // Typically "PLAIN"
	Value struct {
		Text string `json:"text"`
	} `json:"value"`
}
