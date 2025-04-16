package pluginshared

import (
	"encoding/json"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type PigenStepsFile struct {
	Type string `yaml:"type" json:"type"`
	Version string `yaml:"version" json:"version"`
	RepoUrl string `yaml:"repo_url" json:"repo_url"`
	Config map[string]any `yaml:"config" json:"config"`
	Steps []Step `yaml:"steps" json:"steps"`
}

type Step struct {
	Step string `yaml:"step" json:"step"`
	Placeholders map[string]any `yaml:"placeholders" json:"placeholders"`
}

type JSONArgs struct {
	Data string
}

type CicdInterface interface {

	// Connect the repo passed in pigen.yaml to the cicd tool
	ConnectRepo(pigenStepsFile PigenStepsFile) error

	// Create trigger on a repo branch
	CreateTrigger(pigenStepsFile PigenStepsFile) error
	
	// Generate pipeline script

	GeneratScript(pigenStepsFile PigenStepsFile) error

	//TODO: Return service account to give it access to deployed plugins
}

// ###################Client####################
type CicdRPC struct{
	client *rpc.Client
}

func (c *CicdRPC) ConnectRepo(pigenStepsFile PigenStepsFile) error{
	var resp error
	// Convert the PigenStepsFile struct to JSON
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		return err
	}
	args := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}

	err = c.client.Call("Plugin.ConnectRepo", args, &resp)
	if err != nil {
			return err
	}
	return resp
}

func (c *CicdRPC) CreateTrigger(pigenStepsFile PigenStepsFile) error{
	var resp error
	err := c.client.Call("Plugin.CreateTrigger", pigenStepsFile, &resp)
	if err != nil {
		return err
	}
	return resp
}

func (c *CicdRPC) GeneratScript(pigenStepsFile PigenStepsFile) error{
	var resp error
	err := c.client.Call("Plugin.GeneratScript", pigenStepsFile, &resp)
	if err != nil {
		return err
	}
	return resp
}

// ###################Server####################
type CicdRPCServer struct{
	Impl CicdInterface
}


func (s *CicdRPCServer) ConnectRepo(args JSONArgs, resp *error) error {
	var pigenStepsFile PigenStepsFile
	
	if err := json.Unmarshal([]byte(args.Data), &pigenStepsFile); err != nil {
			*resp = NewError(err.Error())
			return nil
	}
	
	err := s.Impl.ConnectRepo(pigenStepsFile)
	if err != nil {
			*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *CicdRPCServer) CreateTrigger(pigenStepsFile PigenStepsFile, resp *error) error{
	err := s.Impl.CreateTrigger(pigenStepsFile)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *CicdRPCServer) GeneratScript(pigenStepsFile PigenStepsFile, resp *error) error{
	err := s.Impl.GeneratScript(pigenStepsFile)
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