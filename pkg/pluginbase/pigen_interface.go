package pluginbase

type PluginInterface interface {
	SetupPlugin(plugin Plugin) error
	GetOutput(plugin Plugin) GetOutputResponse
	Destroy(plugin Plugin) error
}
