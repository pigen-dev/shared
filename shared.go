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
	Error  error // We'll use this to transport the error
}

type CustomError struct {
	Message string
}

// Error implements the error interface
func (e *CustomError) Error() string {
	return e.Message
}

// NewError creates a new CustomError
func NewError(message string) *CustomError {
	return &CustomError{Message: message}
}

func init() {
	gob.Register(&CustomError{})
	gob.Register(&GetOutputResponse{})
}

// ###################Client####################
type PluginRPC struct{
	client *rpc.Client
}

func (c *PluginRPC) ParseConfig(in map[string]interface{}) error{
	var resp error
	err := c.client.Call("Plugin.ParseConfig", in, &resp)
	if err != nil {
		return err
	}
	return resp
}

func (c *PluginRPC) SetupPlugin() error{
	var resp error
	err := c.client.Call("Plugin.SetupPlugin", new(any), &resp)
	if err != nil {
		return err
	}
	return resp
}

func (c *PluginRPC) GetOutput() GetOutputResponse{
	var resp GetOutputResponse
	err := c.client.Call("Plugin.GetOutput", new(any), &resp)
	if err != nil {
		return GetOutputResponse{Output: nil, Error: err}
	}
	return resp
}

func (c *PluginRPC) Destroy() error{
	var resp error
	err := c.client.Call("Plugin.Destroy", new(any), &resp)
	if err != nil {
		return err
	}
	return resp
}


// ###################Server####################
type PluginRPCServer struct{
	Impl PluginInterface
}

func (s *PluginRPCServer) ParseConfig(args map[string]interface{}, resp *error) error{
	err := s.Impl.ParseConfig(args)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *PluginRPCServer) SetupPlugin(args any, resp *error) error{
	err := s.Impl.SetupPlugin()
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *PluginRPCServer) GetOutput(args any, resp *GetOutputResponse) error{
	output := s.Impl.GetOutput()
	if output.Error != nil {
		*resp = GetOutputResponse{Output: nil, Error: NewError(output.Error.Error())}
	} else {
		*resp = output
	}
	return nil
}

func (s *PluginRPCServer) Destroy(args any, resp *error) error{
	err := s.Impl.Destroy()
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
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
