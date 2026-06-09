package plugin

import "encoding/json"

// This file defines the JSON-RPC wire format used between [NewClient] and
// [NewServer]. These types are private implementation details of the plugin
// transport; the protocol can be documented separately if a non-Go plugin
// implementation is ever needed.
//
// A request is a JSON object {"id","method","params"}; a response is a JSON
// object {"id","result","error"}. Requests and responses are concatenated
// JSON values on the connection. Each request is answered by exactly one
// response with the same id; a non-empty error reports a failure of the
// invoked method.

const (
	methodInfo          = "Info"
	methodNewInstance   = "NewInstance"
	methodRunCmd        = "RunCmd"
	methodCloseInstance = "CloseInstance"
)

type rpcRequest struct {
	ID     uint64          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	ID     uint64          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// newInstanceResult is the result of the NewInstance method. Env holds the
// result of the new instance's Env method, folded into the creation reply to
// save a round trip.
type newInstanceResult struct {
	InstID int
	Env    map[string]string
}

type runCmdArgs struct {
	InstID int
	Params CmdParams
}

type closeInstanceArgs struct {
	InstID int
}
