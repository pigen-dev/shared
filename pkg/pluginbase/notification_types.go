package pluginbase

type PipelineNotification struct {
	CicdType string `json:"cicd_type"`
	RepoURL string `json:"repo_url"` // Field renamed to RepoURL, JSON tag preserved
	Branch string `json:"branch"`
	Status string `json:"status"`
	Metadata map[string]string `json:"metadata"`
}