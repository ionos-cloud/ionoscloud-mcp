package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	ionosdns "github.com/ionos-cloud/sdk-go-dns"
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
	tools     []Tool
	client    *ionoscloud.APIClient
	dnsClient *ionosdns.APIClient
	ctx       context.Context
	dnsCtx    context.Context
}

func NewServer() *Server {
	s := &Server{}

	// Initialize IONOS Cloud client
	username := os.Getenv("IONOS_USERNAME")
	password := os.Getenv("IONOS_PASSWORD")
	token := os.Getenv("IONOS_TOKEN")

	// Validate that at least one authentication method is provided
	if username == "" && token == "" {
		fmt.Fprintf(os.Stderr, "Warning: No IONOS Cloud credentials found. Set IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN environment variables.\n")
	}

	configuration := ionoscloud.NewConfiguration(username, password, token, "")
	s.client = ionoscloud.NewAPIClient(configuration)
	s.ctx = context.Background()

	// Initialize DNS client
	dnsConfiguration := ionosdns.NewConfiguration(username, password, token, "")
	s.dnsClient = ionosdns.NewAPIClient(dnsConfiguration)
	// DNS API authentication via context
	if token != "" {
		s.dnsCtx = context.WithValue(context.Background(), ionosdns.ContextAPIKeys, map[string]ionosdns.APIKey{
			"tokenAuth": {Key: token, Prefix: "Bearer"},
		})
	} else if username != "" && password != "" {
		s.dnsCtx = context.WithValue(context.Background(), ionosdns.ContextBasicAuth, ionosdns.BasicAuth{
			UserName: username,
			Password: password,
		})
	} else {
		s.dnsCtx = context.Background()
	}

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
		// Networking - LANs
		{
			Name:        "list_lans",
			Description: "List all LANs in a data center",
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
			Name:        "get_lan",
			Description: "Get details of a specific LAN",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"lan_id": {
						"type": "string",
						"description": "The ID of the LAN"
					}
				},
				"required": ["datacenter_id", "lan_id"]
			}`),
		},
		// Networking - NICs
		{
			Name:        "list_nics",
			Description: "List all NICs attached to a server",
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
			Name:        "get_nic",
			Description: "Get details of a specific NIC",
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
					},
					"nic_id": {
						"type": "string",
						"description": "The ID of the NIC"
					}
				},
				"required": ["datacenter_id", "server_id", "nic_id"]
			}`),
		},
		// Networking - IP Blocks
		{
			Name:        "list_ipblocks",
			Description: "List all IP blocks in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_ipblock",
			Description: "Get details of a specific IP block",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"ipblock_id": {
						"type": "string",
						"description": "The ID of the IP block"
					}
				},
				"required": ["ipblock_id"]
			}`),
		},
		// Networking - Firewall Rules
		{
			Name:        "list_firewall_rules",
			Description: "List all firewall rules on a NIC",
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
					},
					"nic_id": {
						"type": "string",
						"description": "The ID of the NIC"
					}
				},
				"required": ["datacenter_id", "server_id", "nic_id"]
			}`),
		},
		{
			Name:        "get_firewall_rule",
			Description: "Get details of a specific firewall rule",
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
					},
					"nic_id": {
						"type": "string",
						"description": "The ID of the NIC"
					},
					"firewallrule_id": {
						"type": "string",
						"description": "The ID of the firewall rule"
					}
				},
				"required": ["datacenter_id", "server_id", "nic_id", "firewallrule_id"]
			}`),
		},
		// Networking - NAT Gateways
		{
			Name:        "list_nat_gateways",
			Description: "List all NAT gateways in a data center",
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
			Name:        "get_nat_gateway",
			Description: "Get details of a specific NAT gateway",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"nat_gateway_id": {
						"type": "string",
						"description": "The ID of the NAT gateway"
					}
				},
				"required": ["datacenter_id", "nat_gateway_id"]
			}`),
		},
		// Networking - Private Cross Connects
		{
			Name:        "list_pccs",
			Description: "List all Private Cross Connects in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_pcc",
			Description: "Get details of a specific Private Cross Connect",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"pcc_id": {
						"type": "string",
						"description": "The ID of the Private Cross Connect"
					}
				},
				"required": ["pcc_id"]
			}`),
		},
		// Load Balancers - Application Load Balancers
		{
			Name:        "list_application_load_balancers",
			Description: "List all Application Load Balancers in a data center",
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
			Name:        "get_application_load_balancer",
			Description: "Get details of a specific Application Load Balancer",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"alb_id": {
						"type": "string",
						"description": "The ID of the Application Load Balancer"
					}
				},
				"required": ["datacenter_id", "alb_id"]
			}`),
		},
		{
			Name:        "list_alb_forwarding_rules",
			Description: "List all forwarding rules of an Application Load Balancer",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"alb_id": {
						"type": "string",
						"description": "The ID of the Application Load Balancer"
					}
				},
				"required": ["datacenter_id", "alb_id"]
			}`),
		},
		{
			Name:        "get_alb_forwarding_rule",
			Description: "Get details of a specific ALB forwarding rule",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"alb_id": {
						"type": "string",
						"description": "The ID of the Application Load Balancer"
					},
					"rule_id": {
						"type": "string",
						"description": "The ID of the forwarding rule"
					}
				},
				"required": ["datacenter_id", "alb_id", "rule_id"]
			}`),
		},
		// Load Balancers - Network Load Balancers
		{
			Name:        "list_network_load_balancers",
			Description: "List all Network Load Balancers in a data center",
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
			Name:        "get_network_load_balancer",
			Description: "Get details of a specific Network Load Balancer",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"datacenter_id": {
						"type": "string",
						"description": "The ID of the data center"
					},
					"nlb_id": {
						"type": "string",
						"description": "The ID of the Network Load Balancer"
					}
				},
				"required": ["datacenter_id", "nlb_id"]
			}`),
		},
		// Load Balancers - Target Groups
		{
			Name:        "list_target_groups",
			Description: "List all target groups in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_target_group",
			Description: "Get details of a specific target group",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"target_group_id": {
						"type": "string",
						"description": "The ID of the target group"
					}
				},
				"required": ["target_group_id"]
			}`),
		},
		// Kubernetes
		{
			Name:        "list_k8s_clusters",
			Description: "List all Kubernetes clusters in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_k8s_cluster",
			Description: "Get details of a specific Kubernetes cluster",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"k8s_cluster_id": {
						"type": "string",
						"description": "The ID of the Kubernetes cluster"
					}
				},
				"required": ["k8s_cluster_id"]
			}`),
		},
		{
			Name:        "get_k8s_kubeconfig",
			Description: "Get the kubeconfig for a Kubernetes cluster",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"k8s_cluster_id": {
						"type": "string",
						"description": "The ID of the Kubernetes cluster"
					}
				},
				"required": ["k8s_cluster_id"]
			}`),
		},
		{
			Name:        "list_k8s_nodepools",
			Description: "List all node pools in a Kubernetes cluster",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"k8s_cluster_id": {
						"type": "string",
						"description": "The ID of the Kubernetes cluster"
					}
				},
				"required": ["k8s_cluster_id"]
			}`),
		},
		{
			Name:        "get_k8s_nodepool",
			Description: "Get details of a specific node pool",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"k8s_cluster_id": {
						"type": "string",
						"description": "The ID of the Kubernetes cluster"
					},
					"nodepool_id": {
						"type": "string",
						"description": "The ID of the node pool"
					}
				},
				"required": ["k8s_cluster_id", "nodepool_id"]
			}`),
		},
		{
			Name:        "list_k8s_nodes",
			Description: "List all nodes in a node pool",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"k8s_cluster_id": {
						"type": "string",
						"description": "The ID of the Kubernetes cluster"
					},
					"nodepool_id": {
						"type": "string",
						"description": "The ID of the node pool"
					}
				},
				"required": ["k8s_cluster_id", "nodepool_id"]
			}`),
		},
		{
			Name:        "get_k8s_node",
			Description: "Get details of a specific node",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"k8s_cluster_id": {
						"type": "string",
						"description": "The ID of the Kubernetes cluster"
					},
					"nodepool_id": {
						"type": "string",
						"description": "The ID of the node pool"
					},
					"node_id": {
						"type": "string",
						"description": "The ID of the node"
					}
				},
				"required": ["k8s_cluster_id", "nodepool_id", "node_id"]
			}`),
		},
		{
			Name:        "list_k8s_versions",
			Description: "List all available Kubernetes versions",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		// User Management - Users
		{
			Name:        "list_users",
			Description: "List all users in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_user",
			Description: "Get details of a specific user",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"user_id": {
						"type": "string",
						"description": "The ID of the user"
					}
				},
				"required": ["user_id"]
			}`),
		},
		// User Management - Groups
		{
			Name:        "list_groups",
			Description: "List all groups in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_group",
			Description: "Get details of a specific group",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"group_id": {
						"type": "string",
						"description": "The ID of the group"
					}
				},
				"required": ["group_id"]
			}`),
		},
		{
			Name:        "list_group_members",
			Description: "List all users in a group",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"group_id": {
						"type": "string",
						"description": "The ID of the group"
					}
				},
				"required": ["group_id"]
			}`),
		},
		{
			Name:        "list_user_groups",
			Description: "List all groups a user belongs to",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"user_id": {
						"type": "string",
						"description": "The ID of the user"
					}
				},
				"required": ["user_id"]
			}`),
		},
		// User Management - S3 Keys
		{
			Name:        "list_s3_keys",
			Description: "List all S3 keys for a user",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"user_id": {
						"type": "string",
						"description": "The ID of the user"
					}
				},
				"required": ["user_id"]
			}`),
		},
		{
			Name:        "get_s3_key",
			Description: "Get details of a specific S3 key",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"user_id": {
						"type": "string",
						"description": "The ID of the user"
					},
					"key_id": {
						"type": "string",
						"description": "The ID of the S3 key"
					}
				},
				"required": ["user_id", "key_id"]
			}`),
		},
		// User Management - Contract
		{
			Name:        "get_contract",
			Description: "Get contract information and resource limits",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "list_resources",
			Description: "List all resources by type",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"resource_type": {
						"type": "string",
						"description": "The type of resource (optional, e.g., datacenter, image, snapshot, ipblock)"
					}
				},
				"required": []
			}`),
		},
		// DNS
		{
			Name:        "list_dns_zones",
			Description: "List all DNS zones in your IONOS Cloud account",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {},
				"required": []
			}`),
		},
		{
			Name:        "get_dns_zone",
			Description: "Get details of a specific DNS zone",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"zone_id": {
						"type": "string",
						"description": "The ID of the DNS zone"
					}
				},
				"required": ["zone_id"]
			}`),
		},
		{
			Name:        "list_dns_records",
			Description: "List all DNS records in a zone",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"zone_id": {
						"type": "string",
						"description": "The ID of the DNS zone"
					}
				},
				"required": ["zone_id"]
			}`),
		},
		{
			Name:        "get_dns_record",
			Description: "Get details of a specific DNS record",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"zone_id": {
						"type": "string",
						"description": "The ID of the DNS zone"
					},
					"record_id": {
						"type": "string",
						"description": "The ID of the DNS record"
					}
				},
				"required": ["zone_id", "record_id"]
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
