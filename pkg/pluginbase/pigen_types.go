package pluginbase

type PluginStruct struct {
	ID string `yaml:"id" json:"id"`
	RepoUrl string `yaml:"repo_url" json:"repo_url"`
	Version string `yaml:"version" json:"version"`
	Plugin Plugin `yaml:"plugin" json:"plugin"`
}

type Plugin struct {
	Label string `yaml:"label" json:"label"`
  // You can redefine your Config or Output 
	Config map[string]any `yaml:"config" json:"config"`
	Output map[string]any `yaml:"output" json:"output"`
}

type GetOutputResponse struct {
	Output map[string]any
	Error  error
}

// Add a transport-specific structure for RPC communication
type GetOutputRPCResponse struct {
	OutputJSON string // JSON-encoded output map
	Error  error
}
