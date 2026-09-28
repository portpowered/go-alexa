package wire

type DirectiveMessage struct {
	Directive            *Directive             `json:"directive" binding:"required"`
	AdditionalProperties map[string]interface{} `json:"-,omitempty"`
}
