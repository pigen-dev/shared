package pluginbase

// PluginInterface defines the standard operations for Pigen plugins.
type PluginInterface interface {
	SetupPlugin(plugin Plugin) error
	GetOutput(plugin Plugin) GetOutputResponse
	Destroy(plugin Plugin) error
}
