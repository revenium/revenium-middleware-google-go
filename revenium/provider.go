package revenium

type Provider string

const (
	ProviderGoogleAI Provider = "GOOGLE_AI"
	ProviderVertexAI Provider = "VERTEX_AI"
)

func DetectProvider(cfg *Config) Provider {
	if cfg == nil {
		return ProviderGoogleAI
	}

	// If Vertex is explicitly disabled, use Google AI
	if cfg.VertexDisabled {
		return ProviderGoogleAI
	}

	// Auto-detect based on configuration
	if cfg.ProjectID != "" {
		return ProviderVertexAI
	}

	return ProviderGoogleAI
}

func (p Provider) IsGoogleAI() bool {
	return p == ProviderGoogleAI
}

func (p Provider) IsVertexAI() bool {
	return p == ProviderVertexAI
}

func (p Provider) String() string {
	return string(p)
}
