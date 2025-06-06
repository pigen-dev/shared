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

type ActionRequired struct {
	ActionUrl string
	Error error
}

type CICDFile struct {
	FileScript []byte
	Error error
}
type CicdInterface interface {

	// Connect the repo passed in pigen.yaml to the cicd tool
	ConnectRepo(pigenStepsFile PigenStepsFile) ActionRequired

	// Create trigger on a repo branch
	CreateTrigger(pigenStepsFile PigenStepsFile) error
	
	// Generate pipeline script

	GeneratScript(pigenStepsFile PigenStepsFile) CICDFile

	//TODO: Return service account to give it access to deployed plugins
}

// ###################Client####################
type CicdRPC struct{
	client *rpc.Client
}

func (c *CicdRPC) ConnectRepo(pigenStepsFile PigenStepsFile) ActionRequired{
	var resp ActionRequired
	// Convert the PigenStepsFile struct to JSON
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		return ActionRequired{
			ActionUrl: "",
			Error: err,
		}
	}
	args := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}

	err = c.client.Call("Plugin.ConnectRepo", args, &resp)
	if err != nil {
		return ActionRequired{
			ActionUrl: "",
			Error: err,
		}
	}
	return resp
}

func (c *CicdRPC) CreateTrigger(pigenStepsFile PigenStepsFile) error{
	var resp error
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		return err
	}
	jsonArgs := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}
	err = c.client.Call("Plugin.CreateTrigger", jsonArgs, &resp)
	if err != nil {
		return err
	}
	return resp
}

func (c *CicdRPC) GeneratScript(pigenStepsFile PigenStepsFile) CICDFile{
	var resp CICDFile
	pigenStepsFileJSON, err := json.Marshal(pigenStepsFile)
	if err != nil {
		return CICDFile{
			FileScript: nil,
			Error: err,
		}
	}
	jsonArgs := JSONArgs{
		Data: string(pigenStepsFileJSON),
	}
	err = c.client.Call("Plugin.GeneratScript", jsonArgs, &resp)
	if err != nil {
		return CICDFile{
			FileScript: nil,
			Error: err,
		}
	}
	return resp
}

// ###################Server####################
type CicdRPCServer struct{
	Impl CicdInterface
}


func (s *CicdRPCServer) ConnectRepo(args JSONArgs, resp *ActionRequired) error {
	var pigenStepsFile PigenStepsFile
	
	if err := json.Unmarshal([]byte(args.Data), &pigenStepsFile); err != nil {
			*resp = ActionRequired{
					ActionUrl: "",
					Error: NewError(err.Error()),
				}
			return nil
	}
	
	actionRequired := s.Impl.ConnectRepo(pigenStepsFile)
	if actionRequired.Error != nil {
		*resp = ActionRequired{
			ActionUrl: "",
			Error: NewError(actionRequired.Error.Error()),
		}
	} else {
			*resp = actionRequired
	}
	return nil
}

func (s *CicdRPCServer) CreateTrigger(args JSONArgs, resp *error) error{
	var pigenStepsFile PigenStepsFile
	if err := json.Unmarshal([]byte(args.Data), &pigenStepsFile); err != nil {
			*resp = NewError(err.Error())
			return nil
	}
	err := s.Impl.CreateTrigger(pigenStepsFile)
	if err != nil {
		*resp = NewError(err.Error())
	} else {
			*resp = nil
	}
	return nil
}

func (s *CicdRPCServer) GeneratScript(args JSONArgs, resp *CICDFile) error{
	var pigenStepsFile PigenStepsFile
	if err := json.Unmarshal([]byte(args.Data), &pigenStepsFile); err != nil {
			*resp = CICDFile{
				FileScript: nil,
				Error: NewError(err.Error()),
			}
			return nil
	}
	cicdFile := s.Impl.GeneratScript(pigenStepsFile)
	if cicdFile.Error != nil {
		*resp = CICDFile{
			FileScript: nil,
			Error: NewError(cicdFile.Error.Error()),
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