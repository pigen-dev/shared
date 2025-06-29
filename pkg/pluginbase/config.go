package pluginbase

import (
	"encoding/gob"
	"encoding/json"
	"fmt" // Added for error wrapping

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

// EncodeJSONArgs marshals the given arguments into JSONArgs.
func EncodeJSONArgs(args any) (JSONArgs, error) {
	jsonData, err := json.Marshal(args)
	if err != nil {
		return JSONArgs{}, fmt.Errorf("EncodeJSONArgs: failed to marshal data to JSON: %w", err)
	}
	return JSONArgs{Data: string(jsonData)}, nil
}

// DecodeJSONArgs unmarshals data from JSONArgs into the provided 'out' interface.
func DecodeJSONArgs(data JSONArgs, out any) error {
	err := json.Unmarshal([]byte(data.Data), out)
	if err != nil {
		return fmt.Errorf("DecodeJSONArgs: failed to unmarshal JSON data: %w", err)
	}
	return nil
}