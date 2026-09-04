package cert

import (
	"context"
	"fmt"
	"strings"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Shared write-side helpers. All three Certificate Manager PATCH endpoints accept
// only the spec's PatchName, so the update tools differ solely in their SDK call.

// certListLimit is the API's maximum page size. A list response carries no total
// count, so a full page means "at least this many" and the previews say so.
const certListLimit = 1000

// renameNote is appended to every update tool: a model that assumes otherwise would
// silently believe it had changed a field the endpoint cannot touch.
const renameNote = " Only the name can be changed. "

// rename validates a rename request and applies it through the caller's SDK call.
func rename[Out any](rawID, idField, rawName string, patch func(id string, props certSDK.PatchName) (Out, error)) (*mcp.CallToolResult, any, error) {
	id := strings.TrimSpace(rawID)
	if id == "" {
		return tools.ErrorText(idField + " is required"), nil, nil
	}
	name := strings.TrimSpace(rawName)
	if name == "" {
		return tools.ErrorText("name is required and cannot be blank; it is the only field this endpoint can change"), nil, nil
	}
	updated, err := patch(id, certSDK.PatchName{Name: name})
	return tools.ToResult(updated, err)
}

// notFoundText is the message for an operation on a resource that is not there.
func notFoundText(kind, id, verb string) string {
	return fmt.Sprintf("%s %s does not exist; nothing to %s", kind, id, verb)
}

// certificatesIssuedBy counts the certificates an auto-certificate has issued. A
// failure is returned rather than folded into a zero: "none" and "could not tell"
// are different claims to put in front of someone authorizing a delete.
func certificatesIssuedBy(ctx context.Context, client *certSDK.APIClient, autoCertificateID string) (n int, capped bool, err error) {
	list, _, err := client.CertificateApi.CertificatesGet(ctx).FilterAutoCertificate(autoCertificateID).Limit(certListLimit).Execute()
	if err != nil {
		return 0, false, err
	}
	return len(list.Items), len(list.Items) >= certListLimit, nil
}

// autoCertificatesUsing counts the auto-certificates configured with a provider.
// AutoCertificatesGet has no provider filter, so the match is made client-side.
func autoCertificatesUsing(ctx context.Context, client *certSDK.APIClient, providerID string) (n int, capped bool, err error) {
	list, _, err := client.AutoCertificateApi.AutoCertificatesGet(ctx).Limit(certListLimit).Execute()
	if err != nil {
		return 0, false, err
	}
	for _, ac := range list.Items {
		if ac.Properties.Provider == providerID {
			n++
		}
	}
	return n, len(list.Items) >= certListLimit, nil
}
