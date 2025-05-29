package pluginbase

type PipelineNotification struct {
	CicdType string `json:"cicd_type"`
	RepoUrl string `json:"repo_url"`
	Branch string `json:"branch"`
	Status string `json:"status"`
	Metadata map[string]string `json:"metadata"`
}