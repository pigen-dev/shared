package pluginbase

// CicdInterface defines the operations for CICD plugins.
type CicdInterface interface {
	// Connect the repo passed in pigen.yaml to the cicd tool
	ConnectRepo(pigenStepsFile PigenStepsFile) ActionRequired

	// Create trigger on a repo branch
	CreateTrigger(pigenStepsFile PigenStepsFile) error

	// Generate pipeline script
	GeneratScript(pigenStepsFile PigenStepsFile) CICDFile

	//TODO: Return service account to give it access to deployed plugins
}
