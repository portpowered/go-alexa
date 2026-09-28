package wire

type DirectivePayload struct {
	RenderingUpdates     []RenderingUpdate      `json:"renderingUpdates" binding:"required"`
	AdditionalProperties map[string]interface{} `json:"-,omitempty"`
}
