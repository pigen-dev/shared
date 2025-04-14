package pluginshared

import (
	"encoding/gob"
	"errors"
	"fmt"
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
	Error  error // We'll use this to transport the error
}

// ################### Register Error Types #####################
func init() {
	// Register the basic error type returned by errors.New and fmt.Errorf (without %w)
	gob.Register(errors.New(""))

	// OPTIONAL BUT RECOMMENDED: If plugins might put wrapped errors into
	// GetOutputResponse.Error using fmt.Errorf("... %w", someError)
	// var dummyWrappedError error = fmt.Errorf("wrap: %w", errors.New("inner"))
	// gob.Register(dummyWrappedError)

	// Register any custom error types defined in this package (if any)
	// that might be assigned to GetOutputResponse.Error
	// gob.Register(&MySharedErrorType{})
}

// ###################Client####################
type PluginRPC struct{
	client *rpc.Client
}

func (c *PluginRPC) ParseConfig(in map[string]interface{}) error{
	var resp error
	err := c.client.Call("Plugin.ParseConfig", in, &resp)
	if err != nil {
		return fmt.Errorf("rpc call Plugin.ParseConfig failed: %w", err)
	}
	return resp
}

func (c *PluginRPC) SetupPlugin() error{
	var resp error
	err := c.client.Call("Plugin.SetupPlugin", new(any), &resp)
	if err != nil {
		return fmt.Errorf("rpc call Plugin.SetupPlugin failed: %w", err)
	}
	return resp
}

func (c *PluginRPC) GetOutput() GetOutputResponse{
	var resp GetOutputResponse
	err := c.client.Call("Plugin.GetOutput", new(any), &resp)
	if err != nil {
		return GetOutputResponse{Output: nil, Error: fmt.Errorf("rpc communication error getting output: %w", err),}
	}
	return resp
}

func (c *PluginRPC) Destroy() error{
	var resp error
	err := c.client.Call("Plugin.Destroy", new(any), &resp)
	if err != nil {
		return fmt.Errorf("rpc call Plugin.Destroy failed: %w", err)
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
