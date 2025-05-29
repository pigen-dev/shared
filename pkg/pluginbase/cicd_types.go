package pluginbase

type PigenStepsFile struct {
	Type string `yaml:"type" json:"type"`
	Version string `yaml:"version" json:"version"`
	RepoUrl string `yaml:"repo_url" json:"repo_url"`
	Config map[string]any `yaml:"config" json:"config"`
	Steps []Step `yaml:"steps" json:"steps"`
}

type Step struct {
	Step string `yaml:"step" json:"step"`
	Placeholders map[string]any `yaml:"placeholders" json:"placeholders"`
}

type ActionRequired struct {
	ActionUrl string
	Error error
}

type CICDFile struct {
	FileScript []byte
	Error error
}
