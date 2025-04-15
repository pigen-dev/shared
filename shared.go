package pluginshared

import (
	"encoding/gob"
	"encoding/json"
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
	Error  error
}

// Add a transport-specific structure for RPC communication
type GetOutputRPCResponse struct {
	OutputJSON string // JSON-encoded output map
	Error  error
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
	gob.Register(GetOutputRPCResponse{})
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
	var rpcResp GetOutputRPCResponse
	err := c.client.Call("Plugin.GetOutput", new(any), &rpcResp)
	var outputMap map[string]interface{}
  var outputErr error
	if err != nil {
		outputErr = NewError(err.Error())
	} else {
		if rpcResp.Error != nil {
			outputErr = NewError(rpcResp.Error.Error())
		} else {
			if err := json.Unmarshal([]byte(rpcResp.OutputJSON), &outputMap); err != nil {
				outputErr = NewError(err.Error())
			}
		}
	}
	return GetOutputResponse{Output: outputMap, Error: outputErr}
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

func (s *PluginRPCServer) GetOutput(args any, resp *GetOutputRPCResponse) error{
	result := s.Impl.GetOutput()
	if result.Output != nil {
		jsonData, err := json.Marshal(result.Output)
		if err != nil {
				resp.OutputJSON = "{}"
				resp.Error = NewError(fmt.Errorf("failed to json marshal output: %w", err).Error())
				return nil
		}
		resp.OutputJSON = string(jsonData)
	} else {
			resp.OutputJSON = "{}"
	}

	// Serialize any error to JSON
	if result.Error != nil {
			resp.Error = NewError(result.Error.Error())
	} else {
			resp.Error = nil
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
