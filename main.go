package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	serverName    = "ionoscloud-mcp"
	serverVersion = "0.1.0"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Server struct {
	tools []Tool
}

func NewServer() *Server {
	s := &Server{}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	s.tools = []Tool{
		{
			Name:        "list_datacenters",
			Description: "List all virtual data centers in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_datacenter",
			Description: "Get details of a specific virtual data center",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					}
				},
				"required": ["datacenter_id"]
			}`),
		},
		{
			Name:        "list_servers",
			Description: "List all servers in a data center",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					}
				},
				"required": ["datacenter_id"]
			}`),
		},
		{
			Name:        "get_server",
			Description: "Get details of a specific server",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"server_id": {
						"type": "string",
						"description": "The ID of the server"
					}
				},
				"required": ["datacenter_id", "server_id"]
			}`),
		},
		{
			Name:        "list_volumes",
			Description: "List all volumes in a data center",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					}
				},
				"required": ["datacenter_id"]
			}`),
		},
		{
			Name:        "get_volume",
			Description: "Get details of a specific volume",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"volume_id": {
						"type": "string",
						"description": "The ID of the volume"
					}
				},
				"required": ["datacenter_id", "volume_id"]
			}`),
		},
		{
			Name:        "list_images",
			Description: "List all available images (OS templates) in IONOS Cloud",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "list_locations",
			Description: "List all available locations (regions) in IONOS Cloud",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "list_snapshots",
			Description: "List all snapshots in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_snapshot",
			Description: "Get details of a specific snapshot",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"snapshot_id": {
						"type": "string",
						"description": "The ID of the snapshot"
					}
				},
				"required": ["snapshot_id"]
			}`),
		},
	}
}

func (s *Server) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    serverName,
				"version": serverVersion,
			},
		},
	}
}

func (s *Server) handleToolsList(req *JSONRPCRequest) *JSONRPCResponse {
	tools := make([]map[string]interface{}, len(s.tools))
	for i, tool := range s.tools {
		tools[i] = map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

func (s *Server) handleToolsCall(req *JSONRPCRequest) *JSONRPCResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    err.Error(),
			},
		}
	}

	result, err := s.executeTool(params.Name, params.Arguments)
	if err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
		}
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
}

func (s *Server) Run() error {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse JSON-RPC request: %v\n", err)
			continue
		}

		resp := s.handleRequest(&req)

		respBytes, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal JSON-RPC response: %v\n", err)
			// Send a basic error response
			errorResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &RPCError{
					Code:    -32603,
					Message: "Internal error: failed to marshal response",
				},
			}
			if errBytes, e := json.Marshal(errorResp); e == nil {
				writer.Write(errBytes)
				writer.WriteByte('\n')
				writer.Flush()
			}
			continue
		}

		writer.Write(respBytes)
		writer.WriteByte('\n')
		writer.Flush()
	}

	return nil
}

func main() {
	server := NewServer()
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
