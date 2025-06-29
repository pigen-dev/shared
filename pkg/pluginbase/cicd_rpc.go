package pluginbase

import (
	"encoding/json"
	"fmt" // Added for error wrapping
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// ###################Client####################
type CicdRPC struct{
	client *rpc.Client
}

func (c *CicdRPC) ConnectRepo(pigenStepsFile PigenStepsFile) ActionRequired{
	var resp ActionRequired
	// Convert the PigenStepsFile struct to JSON
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		resp.Error = NewError(fmt.Sprintf("failed to marshal request for ConnectRepo: %v", err))
		return resp
	}
	args := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}

	err = c.client.Call("Plugin.ConnectRepo", args, &resp)
	if err != nil {
		resp.Error = NewError(fmt.Sprintf("RPC call ConnectRepo failed: %v", err))
		return resp
	}
	return resp
}

func (c *CicdRPC) CreateTrigger(pigenStepsFile PigenStepsFile) error {
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		return NewError(fmt.Sprintf("failed to marshal request for CreateTrigger: %v", err))
	}
	jsonArgs := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}
	var serverResponse *CustomError // Server is expected to return *CustomError for business logic errors
	err = c.client.Call("Plugin.CreateTrigger", jsonArgs, &serverResponse)
	if err != nil { // RPC communication error
		return NewError(fmt.Sprintf("RPC call CreateTrigger failed: %v", err))
	}
	if serverResponse != nil { // Business logic error from server
		return serverResponse
	}
	return nil // Success
}

func (c *CicdRPC) GenerateScript(pigenStepsFile PigenStepsFile) CICDFile { // Renamed
	var resp CICDFile
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		resp.Error = NewError(fmt.Sprintf("failed to marshal request for GenerateScript: %v", err))
		return resp
	}
	jsonArgs := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}
	// RPC call string updated
	err = c.client.Call("Plugin.GenerateScript", jsonArgs, &resp)
	if err != nil {
		resp.Error = NewError(fmt.Sprintf("RPC call GenerateScript failed: %v", err))
		return resp
	}
	return resp
}

// ###################Server####################
type CicdRPCServer struct{
	Impl CicdInterface
}


func (s *CicdRPCServer) ConnectRepo(args JSONArgs, resp *ActionRequired) error {
	var pigenStepsFile PigenStepsFile
	// Use DecodeJSONArgs
	if err := DecodeJSONArgs(args, &pigenStepsFile); err != nil {
		*resp = ActionRequired{
			ActionUrl: "",
			Error:     NewError(err.Error()), // NewError is in pluginbase.config
		}
		return nil
	}

	actionRequired := s.Impl.ConnectRepo(pigenStepsFile)
	if actionRequired.Error != nil {
		*resp = ActionRequired{
			ActionUrl: "",
			Error:     NewError(actionRequired.Error.Error()), // NewError is in pluginbase.config
		}
	} else {
			*resp = actionRequired
	}
	return nil
}

func (s *CicdRPCServer) CreateTrigger(args JSONArgs, resp *error) error { // Note: resp is *error, should be *CustomError for consistency
	var pigenStepsFile PigenStepsFile
	// Use DecodeJSONArgs
	if err := DecodeJSONArgs(args, &pigenStepsFile); err != nil {
		// The resp type for this method is `*error`.
		// We should assign a `*CustomError` to it, which implements the error interface.
		customErr := NewError(err.Error()) // NewError is in pluginbase.config
		*resp = customErr
		return nil
	}
	errFromImpl := s.Impl.CreateTrigger(pigenStepsFile)
	if errFromImpl != nil {
		customErr := NewError(errFromImpl.Error()) // NewError is in pluginbase.config
		*resp = customErr
	} else {
		*resp = nil
	}
	return nil
}

func (s *CicdRPCServer) GenerateScript(args JSONArgs, resp *CICDFile) error { // Renamed
	var pigenStepsFile PigenStepsFile
	// Use DecodeJSONArgs
	if err := DecodeJSONArgs(args, &pigenStepsFile); err != nil {
		*resp = CICDFile{
			FileScript: nil,
			Error:      NewError(err.Error()), // NewError is in pluginbase.config
		}
		return nil
	}
	// Call the renamed interface method
	cicdFile := s.Impl.GenerateScript(pigenStepsFile)
	if cicdFile.Error != nil {
		*resp = CICDFile{
			FileScript: nil,
			Error:      NewError(cicdFile.Error.Error()), // NewError is in pluginbase.config
		}
	} else {
		*resp = cicdFile
	}
	return nil
}

type CicdPlugin struct{
	Impl CicdInterface
}

func (g *CicdPlugin) Server(*plugin.MuxBroker)(any, error){
	return &CicdRPCServer{Impl: g.Impl}, nil
}

func (CicdPlugin) Client(b *plugin.MuxBroker, c *rpc.Client)(any, error){
	return &CicdRPC{client: c}, nil
}
