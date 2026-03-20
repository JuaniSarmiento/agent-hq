package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string     `json:"jsonrpc"`
	ID      any        `json:"id,omitempty"`
	Result  any        `json:"result,omitempty"`
	Error   *RPCError  `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Server handles MCP protocol communication over stdio.
type Server struct {
	handler *ToolHandler
	reader  io.Reader
	writer  io.Writer
}

// NewServer creates a new MCP server.
func NewServer(handler *ToolHandler, reader io.Reader, writer io.Writer) *Server {
	return &Server{
		handler: handler,
		reader:  reader,
		writer:  writer,
	}
}

// Run reads JSON-RPC requests from stdin and writes responses to stdout.
func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	// Allow large messages (1MB).
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("invalid JSON-RPC request: %v", err)
			s.writeResponse(JSONRPCResponse{
				JSONRPC: "2.0",
				Error: &RPCError{
					Code:    -32700,
					Message: "Parse error",
				},
			})
			continue
		}

		resp := s.handleRequest(req)
		s.writeResponse(resp)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "agent-hq",
					"version": "0.1.0",
				},
			},
		}

	case "notifications/initialized":
		// Client acknowledgment — no response needed for notifications.
		// Return empty response; caller can skip writing if ID is nil.
		return JSONRPCResponse{JSONRPC: "2.0"}

	case "tools/list":
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.handler.ListTools(),
			},
		}

	case "tools/call":
		result, err := s.handler.CallTool(req.Params)
		if err != nil {
			return JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": fmt.Sprintf("Error: %v", err),
						},
					},
					"isError": true,
				},
			}
		}
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &RPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

func (s *Server) writeResponse(resp JSONRPCResponse) {
	// Don't write responses for notifications (no ID).
	if resp.ID == nil && resp.Error == nil && resp.Result == nil {
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		return
	}
	fmt.Fprintf(s.writer, "%s\n", data)
}
