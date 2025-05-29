package pluginbase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Client RPC for Plugins
type PluginRPC struct{
	client *rpc.Client
}

func (c *PluginRPC) SetupPlugin(plugin Plugin) error{
	var resp error
	args, err := GobEncode(plugin) // GobEncode is in pluginbase.config
	if err != nil {
		log.Printf("Error encoding plugin: %v", err)
		return err
	}
	err = c.client.Call("Plugin.SetupPlugin", args, &resp)
	if err != nil {
		log.Printf("Error calling SetupPlugin: %v", err)
		return err
	}
	if resp != nil { // Check if the error returned from RPC is not nil
		return resp
	}
	return nil
}

func (c *PluginRPC) GetOutput(plugin Plugin) GetOutputResponse{
	var rpcResp GetOutputRPCResponse
	args, err := GobEncode(plugin) // GobEncode is in pluginbase.config
	if err != nil {
		return GetOutputResponse{
			Output: nil,
			Error: NewError(err.Error()), // NewError is in pluginbase.config
		}
	}
	err = c.client.Call("Plugin.GetOutput", args, &rpcResp)
	var outputMap map[string]any
  var outputErr error
	if err != nil {
		outputErr = NewError(err.Error()) // NewError is in pluginbase.config
	} else {
		if rpcResp.Error != nil {
			outputErr = NewError(rpcResp.Error.Error()) // NewError is in pluginbase.config
		} else {
			if err := json.Unmarshal([]byte(rpcResp.OutputJSON), &outputMap); err != nil {
				outputErr = NewError(err.Error()) // NewError is in pluginbase.config
			}
		}
	}
	return GetOutputResponse{Output: outputMap, Error: outputErr}
}

func (c *PluginRPC) Destroy(plugin Plugin) error{
	var resp error
	args, err := GobEncode(plugin) // GobEncode is in pluginbase.config
	if err != nil {
		return err
	}
	err = c.client.Call("Plugin.Destroy", args, &resp)
	if err != nil {
		return err
	}
	if resp != nil { // Check if the error returned from RPC is not nil
		return resp
	}
	return nil
}


// Server RPC for Plugins
type PluginRPCServer struct{
	Impl PluginInterface
}


func (s *PluginRPCServer) SetupPlugin(args JSONArgs, resp *error) error{
	var plugin Plugin
	err := GobDecode(args, &plugin) // GobDecode is in pluginbase.config
	if err != nil {
		*resp = NewError(fmt.Errorf("failed to decode args: %w", err).Error()) // NewError is in pluginbase.config
		return nil
	}
	err = s.Impl.SetupPlugin(plugin)
	if err != nil {
		*resp = NewError(err.Error()) // NewError is in pluginbase.config
	} else {
			*resp = nil
	}
	return nil
}

func (s *PluginRPCServer) GetOutput(args JSONArgs, resp *GetOutputRPCResponse) error{
	var plugin Plugin
	err := GobDecode(args, &plugin) // GobDecode is in pluginbase.config
	if err != nil {
		resp.OutputJSON = "{}"
		resp.Error = NewError(fmt.Errorf("failed to decode args: %w", err).Error()) // NewError is in pluginbase.config
		return nil
	}
	result := s.Impl.GetOutput(plugin)
	if result.Output != nil {
		jsonData, err := json.Marshal(result.Output)
		if err != nil {
				resp.OutputJSON = "{}"
				resp.Error = NewError(fmt.Errorf("failed to json marshal output: %w", err).Error()) // NewError is in pluginbase.config
				return nil
		}
		resp.OutputJSON = string(jsonData)
	} else {
			resp.OutputJSON = "{}"
	}

	// Serialize any error to JSON
	if result.Error != nil {
			resp.Error = NewError(result.Error.Error()) // NewError is in pluginbase.config
	} else {
			resp.Error = nil
	}

	return nil
}

func (s *PluginRPCServer) Destroy(args JSONArgs, resp *error) error{
	var plugin Plugin
	err := GobDecode(args, &plugin) // GobDecode is in pluginbase.config
	if err != nil {
		*resp = NewError(fmt.Errorf("failed to decode args: %w", err).Error()) // NewError is in pluginbase.config
		return nil
	}
	// Call the Destroy method on the plugin implementation
	err = s.Impl.Destroy(plugin)
	if err != nil {
		*resp = NewError(err.Error()) // NewError is in pluginbase.config
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
