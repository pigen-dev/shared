package pluginshared

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type CicdInterface interface {

	// Connect the repo passed in pigen.yaml to the cicd tool
	ConnectRepo(in map[string] interface{}) error

	// Create trigger on a repo branch
	CreateTrigger(in map[string] interface{}) error
	
	// Generate pipeline script

	GeneratScript(in map[string] interface{}) error

	//TODO: Return service account to give it access to deployed plugins
}



// ###################Client####################
type CicdRPC struct{
	client *rpc.Client
}

func (c *CicdRPC) ConnectRepo(in map[string]any) error{
	var resp error
	err := c.client.Call("Plugin.ConnectRepo", in, &resp)
	if err != nil {
		return err
	}
	return resp
}

func (c *CicdRPC) CreateTrigger(in map[string]any) error{
	var resp error
	err := c.client.Call("Plugin.CreateTrigger", in, &resp)
	if err != nil {
		return err
	}
	return resp
}

func (c *CicdRPC) GeneratScript(in map[string]any) error{
	var resp error
	err := c.client.Call("Plugin.GeneratScript", in, &resp)
	if err != nil {
		return err
	}
	return resp
}

// ###################Server####################
type CicdRPCServer struct{
	Impl CicdInterface
}


func (s *CicdRPCServer) ConnectRepo(args map[string]any, resp *error) error{
	err := s.Impl.ConnectRepo(args)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *CicdRPCServer) CreateTrigger(args map[string]any, resp *error) error{
	err := s.Impl.CreateTrigger(args)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *CicdRPCServer) GeneratScript(args map[string]any, resp *error) error{
	err := s.Impl.GeneratScript(args)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
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