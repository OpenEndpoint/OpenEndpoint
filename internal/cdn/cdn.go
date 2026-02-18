package cdn

// Handler handles CDN integration
// This is a stub for v3 implementation
type Handler struct {
	provider string
}

// NewHandler creates a new CDN handler
func NewHandler(provider string) *Handler {
	return &Handler{
		provider: provider,
	}
}

// InvalidateCache invalidates the CDN cache
func (h *Handler) InvalidateCache(paths []string) error {
	return nil
}

// GenerateSignedURL generates a signed CDN URL
func (h *Handler) GenerateSignedURL(path string, expiry int64) (string, error) {
	return path, nil
}

// GetProvider returns the CDN provider name
func (h *Handler) GetProvider() string {
	return h.provider
}
