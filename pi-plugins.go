package pluginshared

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type PluginStruct struct {
	ID string `yaml:"id" json:"id"`
	RepoUrl string `yaml:"repo_url" json:"repo_url"`
	Version string `yaml:"version" json:"version"`
	Plugin Plugin `yaml:"plugin" json:"plugin"`
}

type Plugin struct {
	Label string `yaml:"label" json:"label"`
  // You can redefine your Config or Output 
	Config map[string]any `yaml:"config" json:"config"`
	Output map[string]any `yaml:"output" json:"output"`
}
type PluginInterface interface {
	SetupPlugin(plugin Plugin) error
	GetOutput(plugin Plugin) GetOutputResponse
	Destroy(plugin Plugin) error
}

type GetOutputResponse struct {
	Output map[string]any
	Error  error
}

// Add a transport-specific structure for RPC communication
type GetOutputRPCResponse struct {
	OutputJSON string // JSON-encoded output map
	Error  error
}

// ###################Client####################
type PluginRPC struct{
	client *rpc.Client
}

func (c *PluginRPC) SetupPlugin(plugin Plugin) error{
	var resp error
	args, err := GobEncode(plugin)
	if err != nil {
		log.Printf("Error encoding plugin: %v", err)
		return err
	}
	err = c.client.Call("Plugin.SetupPlugin", args, &resp)
	if err != nil {
		log.Printf("Error calling SetupPlugin: %v", err)
		return err
	}
	return resp
}

func (c *PluginRPC) GetOutput(plugin Plugin) GetOutputResponse{
	var rpcResp GetOutputRPCResponse
	args, err := GobEncode(plugin)
	if err != nil {
		return GetOutputResponse{
			Output: nil,
			Error: NewError(err.Error()),
		}
	}
	err = c.client.Call("Plugin.GetOutput", args, &rpcResp)
	var outputMap map[string]any
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

func (c *PluginRPC) Destroy(plugin Plugin) error{
	var resp error
	args, err := GobEncode(plugin)
	if err != nil {
		return err
	}
	err = c.client.Call("Plugin.Destroy", args, &resp)
	if err != nil {
		return err
	}
	return resp
}


// ###################Server####################
type PluginRPCServer struct{
	Impl PluginInterface
}


func (s *PluginRPCServer) SetupPlugin(args JSONArgs, resp *error) error{
	var plugin Plugin
	err := GobDecode(args, &plugin)
	if err != nil {
		*resp = NewError(fmt.Errorf("failed to decode args: %w", err).Error())
		return nil
	}
	err = s.Impl.SetupPlugin(plugin)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *PluginRPCServer) GetOutput(args JSONArgs, resp *GetOutputRPCResponse) error{
	var plugin Plugin
	err := GobDecode(args, &plugin)
	if err != nil {
		resp.OutputJSON = "{}"
		resp.Error = NewError(fmt.Errorf("failed to decode args: %w", err).Error())
		return nil
	}
	result := s.Impl.GetOutput(plugin)
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

func (s *PluginRPCServer) Destroy(args JSONArgs, resp *error) error{
	var plugin Plugin
	err := GobDecode(args, &plugin)
	if err != nil {
		*resp = NewError(fmt.Errorf("failed to decode args: %w", err).Error())
		return nil
	}
	// Call the Destroy method on the plugin implementation
	err = s.Impl.Destroy(plugin)
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


