package pluginshared

import (
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type PluginStruct struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Plugin Plugin `yaml:"plugin"`
}

type Plugin struct {
  // You can redefine your Config or Output 
	Config map[string]any `yaml:"config"`
	Output map[string]any `yaml:"output"`
}
type PluginInterface interface {
	ParseConfig(in map[string]any) error
	SetupPlugin(config map[string] any) error
	GetOutput(config map[string] any) GetOutputResponse
	Destroy(config map[string] any) error
}

type GetOutputResponse struct {
	Output map[string]interface{}
	Error  error // We'll use this to transport the error
}

// ###################Client####################
type PluginRPC struct{
	client *rpc.Client
}

func (c *PluginRPC) ParseConfig(in map[string]interface{}) error{
	var resp error
	err := c.client.Call("Plugin.ParseConfig", in, &resp)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func (c *PluginRPC) SetupPlugin(config map[string] any) error{
	var resp error
	err := c.client.Call("Plugin.SetupPlugin", config, &resp)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func (c *PluginRPC) GetOutput(config map[string] any) GetOutputResponse{
	var resp GetOutputResponse
	err := c.client.Call("Plugin.GetOutput", config, &resp)
	if err != nil {
		log.Fatal(err)
	}
	return resp
}

func (c *PluginRPC) Destroy(config map[string] any) error{
	var resp error
	err := c.client.Call("Plugin.Destroy", config, &resp)
	if err != nil {
		log.Fatal(err)
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

func (s *PluginRPCServer) SetupPlugin(args map[string]interface{}, resp *error) error{
	*resp = s.Impl.SetupPlugin(args)
	return nil
}

func (s *PluginRPCServer) GetOutput(args map[string]interface{}, resp *GetOutputResponse) error{
	*resp = s.Impl.GetOutput(args)
	return nil
}

func (s *PluginRPCServer) Destroy(args map[string]interface{}, resp *error) error{
	*resp = s.Impl.Destroy(args)
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
