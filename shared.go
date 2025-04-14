package pluginshared

import (
	"encoding/gob"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type PluginStruct struct {
	ID string `yaml:"id" json:"id"`
	RepoUrl string `yaml:"repo_url" json:"repo_url"`
	Version string `yaml:"version" json:"version"`
	Label string `yaml:"label" json:"label"`
	Plugin Plugin `yaml:"plugin" json:"plugin"`
}

type Plugin struct {
  // You can redefine your Config or Output 
	Config map[string]any `yaml:"config" json:"config"`
	Output map[string]any `yaml:"output" json:"output"`
}
type PluginInterface interface {
	ParseConfig(in map[string]any) error
	SetupPlugin() error
	GetOutput() GetOutputResponse
	Destroy() error
}

type GetOutputResponse struct {
	Output map[string]interface{}
	Error  *PluginError // We'll use this to transport the error
}

type PluginError struct {
	Message string
}

func (e *PluginError) Error() string {
	return e.Message
}

func init() {
	gob.Register(&PluginError{})
}

// ###################Client####################
type PluginRPC struct{
	client *rpc.Client
}

func (c *PluginRPC) ParseConfig(in map[string]interface{}) error{
	var resp error
	err := c.client.Call("Plugin.ParseConfig", in, &resp)
	if err != nil {
		return &PluginError{Message: err.Error()}
	}
	return resp
}

func (c *PluginRPC) SetupPlugin() error{
	var resp error
	err := c.client.Call("Plugin.SetupPlugin", new(any), &resp)
	if err != nil {
		return &PluginError{Message: err.Error()}
	}
	return resp
}

func (c *PluginRPC) GetOutput() GetOutputResponse{
	var resp GetOutputResponse
	err := c.client.Call("Plugin.GetOutput", new(any), &resp)
	if err != nil {
		return GetOutputResponse{Output: nil, Error: &PluginError{Message: err.Error()},}
	}
	return resp
}

func (c *PluginRPC) Destroy() error{
	var resp error
	err := c.client.Call("Plugin.Destroy", new(any), &resp)
	if err != nil {
		return &PluginError{Message: err.Error()}
	}
	return resp
}


// ###################Server####################
type PluginRPCServer struct{
	Impl PluginInterface
}

func (s *PluginRPCServer) ParseConfig(args map[string]interface{}, resp *error) error{
	*resp = s.Impl.ParseConfig(args)
	return nil
}

func (s *PluginRPCServer) SetupPlugin(args any, resp *error) error{
	*resp = s.Impl.SetupPlugin()
	return nil
}

func (s *PluginRPCServer) GetOutput(args any, resp *GetOutputResponse) error{
	*resp = s.Impl.GetOutput()
	return nil
}

func (s *PluginRPCServer) Destroy(args any, resp *error) error{
	*resp = s.Impl.Destroy()
	return nil
}

type PigenPlugin struct{
	Impl PluginInterface
}

func (g *PigenPlugin) Server(*plugin.MuxBroker)(any, error){
	return &PluginRPCServer{Impl: g.Impl}, nil
}

func (PigenPlugin) Client(b *plugin.MuxBroker, c *rpc.Client)(any, error){
	return &PluginRPC{client: c}, nil
}

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}
