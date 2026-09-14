package cert

import (
	"context"
	"fmt"
	"strings"
	"time"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterCertificateWriteTools(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerUpdateCertificate(server, client, scope)
	registerDeleteCertificate(server, client, scope, confirm)
}

func registerUpdateCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_cert_certificate",
		Description: "Rename a certificate." + renameNote +
			"The certificate body, chain and private key are immutable, and this server has no tool for uploading certificate material. To rotate, issue a replacement with create_cert_auto_certificate and repoint the load balancer at it, then delete the old one." + updatedNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateCertCertificateInput) (*mcp.CallToolResult, any, error) {
		return rename(input.CertificateID, "certificate_id", input.Name,
			func(id string, props certSDK.PatchName) (certSDK.CertificateRead, error) {
				updated, _, err := client.CertificateApi.CertificatesPatch(ctx, id).CertificatePatch(certSDK.CertificatePatch{Properties: props}).Execute()
				return redactCertificate(updated), err
			})
	})
}

func registerDeleteCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_cert_certificate",
		Description: "Delete one SSL/TLS certificate and its private key. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. This is irreversible, and this server cannot upload a replacement — re-creating it means issuing a new certificate with create_cert_auto_certificate, or uploading outside this server. " +
			"Any Application Load Balancer HTTPS listener still referencing this certificate stops serving TLS. Certificate Manager cannot list those references, so check the ALB forwarding rules first with list_alb_forwarding_rules." + deletedNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteCertCertificateInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.CertificateID)
		if id == "" {
			return tools.ErrorText("certificate_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_cert_certificate", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_cert_certificate", "certificate_id", err)), nil, nil
			}
			if _, err := client.CertificateApi.CertificatesDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("certificate", id) +
				" get_cert_certificate answers 404 once it is gone."), nil, nil
		}

		crt, _, err := client.CertificateApi.CertificatesFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(notFoundText("certificate", id, "delete")), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("delete_cert_certificate", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		md := crt.Metadata
		headline := "About to DELETE a certificate and its private key. This is IRREVERSIBLE."
		if md.AutoCertificate != nil {
			headline += "\nNOTE: this certificate was issued by an auto-certificate, which still exists and will issue a replacement at the next renewal. Delete that too to stop renewals."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"certificate_id", id,
				"name", crt.Properties.Name,
				"common_name", md.CommonName,
				"subject_alternative_names", strings.Join(md.SubjectAlternativeNames, ", "),
				"expires", timeText(md.NotAfter),
				"expired", fmt.Sprintf("%t", md.Expired),
				"state", md.State,
				"issued_by_auto_certificate", tools.OptStr(md.AutoCertificate),
			),
			Tool:      "delete_cert_certificate",
			Replay:    tools.Fields("certificate_id", id),
			TokenNote: "This token authorizes deleting ONLY this certificate",
		}.Render(token)), nil, nil
	})
}

// timeText renders an optional API timestamp for a preview field.
func timeText(t *certSDK.IonosTime) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
