package api

import (
	"context"
	"encoding/json"

	dns "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// ToolHandlerParams contains all parameters passed to a tool handler.
type ToolHandlerParams struct {
	// Context for the request
	Context context.Context
	// Client is the IONOS Cloud compute API client
	Client *ionoscloud.APIClient
	// DNSClient is the IONOS Cloud DNS API client
	DNSClient *dns.APIClient
	// Arguments contains the tool input parameters
	Arguments map[string]interface{}
}

// ToolCallResult represents the result of a tool execution.
type ToolCallResult struct {
	// Content is the result content, typically JSON-formatted text
	Content string
	// IsError indicates if this is an error result
	IsError bool
}

// NewTextResult creates a successful text result.
func NewTextResult(content string) *ToolCallResult {
	return &ToolCallResult{
		Content: content,
		IsError: false,
	}
}

// NewErrorResult creates an error result.
func NewErrorResult(err error) *ToolCallResult {
	return &ToolCallResult{
		Content: err.Error(),
		IsError: true,
	}
}

// MarshalResult marshals any value to an indented JSON string result.
func MarshalResult(v interface{}, resourceName string) (*ToolCallResult, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return NewTextResult(string(data)), nil
}

// StatusResult creates a JSON result for status messages.
func StatusResult(fields map[string]string) (*ToolCallResult, error) {
	data, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return nil, err
	}
	return NewTextResult(string(data)), nil
}

// GetRequiredString extracts a required string parameter from arguments.
func GetRequiredString(args map[string]interface{}, key string) (string, bool) {
	v, ok := args[key].(string)
	return v, ok && v != ""
}

// GetOptionalString extracts an optional string parameter from arguments.
func GetOptionalString(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

// GetOptionalBool extracts an optional bool parameter from arguments.
// Returns the value and whether it was present.
func GetOptionalBool(args map[string]interface{}, key string) (bool, bool) {
	v, ok := args[key].(bool)
	return v, ok
}

// GetOptionalFloat extracts an optional float64 parameter from arguments.
// Returns the value and whether it was present.
func GetOptionalFloat(args map[string]interface{}, key string) (float64, bool) {
	v, ok := args[key].(float64)
	return v, ok
}

// GetOptionalInt32 extracts an optional int32 parameter from arguments.
// Returns the value and whether it was present.
func GetOptionalInt32(args map[string]interface{}, key string) (int32, bool) {
	v, ok := args[key].(float64)
	if ok {
		return int32(v), true
	}
	return 0, false
}

// GetStringSlice extracts a string slice from arguments.
func GetStringSlice(args map[string]interface{}, key string) []string {
	raw, ok := args[key].([]interface{})
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
