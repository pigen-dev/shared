package pluginbase

type PigenStepsFile struct {
	Type string `yaml:"type" json:"type"`
	Version string `yaml:"version" json:"version"`
	RepoURL string `yaml:"repo_url" json:"repo_url"` // Field renamed
	Config map[string]any `yaml:"config" json:"config"`
	Steps []Step `yaml:"steps" json:"steps"`
}

type Step struct {
	Step string `yaml:"step" json:"step"`
	Placeholders map[string]any `yaml:"placeholders" json:"placeholders"`
}

type ActionRequired struct {
	ActionURL string // Field renamed
	Error error
}

type CICDFile struct {
	FileScript []byte
	Error error
}
