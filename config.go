package pluginshared

import (
	"encoding/gob"

	"github.com/hashicorp/go-plugin"
)

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