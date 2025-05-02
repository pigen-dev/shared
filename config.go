package pluginshared

import (
	"encoding/gob"
	"encoding/json"

	"github.com/hashicorp/go-plugin"
)

type JSONArgs struct {
	Data string
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

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func GobEncode(args any) (JSONArgs, error) {
	jsonData, err := json.Marshal(args)
	if err != nil {
		return JSONArgs{}, err
	}
	return JSONArgs{Data: string(jsonData)}, nil
}

func GobDecode(data JSONArgs, out any) error {
	err := json.Unmarshal([]byte(data.Data), out)
	if err != nil {
		return err
	}
	return nil
}