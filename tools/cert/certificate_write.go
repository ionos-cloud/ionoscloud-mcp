package cert

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterCertificateWriteTools registers the create/update/delete certificate tools.
func RegisterCertificateWriteTools(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateCertificate(server, client, scope, confirm)
	registerUpdateCertificate(server, client, scope)
	registerDeleteCertificate(server, client, scope, confirm)
}

func registerCreateCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_cert_certificate",
		Description: "Upload one SSL/TLS certificate with its chain and private key, for use by an Application Load Balancer HTTPS listener. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token and the same material to upload it. Uploads exactly one certificate per call. " +
			"The private key is write-only: it is never echoed in the preview and never returned by a read tool. For a certificate that renews itself, use create_cert_provider and create_cert_auto_certificate instead of uploading here." + createdNote + stateNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateCertCertificateInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return tools.ErrorText("name is required to create a certificate"), nil, nil
		}
		for _, f := range []struct{ field, value string }{
			{"certificate", input.Certificate},
			{"certificate_chain", input.CertificateChain},
			{"private_key", input.PrivateKey},
		} {
			if msg := validatePEM(f.field, f.value); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}
		props := certSDK.Certificate{
			Name:             name,
			Certificate:      strings.TrimSpace(input.Certificate),
			CertificateChain: strings.TrimSpace(input.CertificateChain),
			PrivateKey:       strings.TrimSpace(input.PrivateKey),
		}
		// The material is in the target: one certificate's token cannot upload another.
		target := tools.Target(req, name, materialDigest(props))

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_cert_certificate", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_cert_certificate", "name, certificate, certificate_chain and private_key", err)), nil, nil
			}
			created, _, err := client.CertificateApi.CertificatesPost(ctx).CertificateCreate(certSDK.CertificateCreate{Properties: props}).Execute()
			return tools.ToResult(redactCertificate(created), err)
		}

		token, err := confirm.Mint("create_cert_certificate", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one certificate:",
			Fields: tools.Fields(
				"name", name,
				"certificate", certificateSummary(props.Certificate),
				"certificate_chain", pemBlockCount(props.CertificateChain),
				"private_key", tools.Redacted(&props.PrivateKey),
			),
			Tool:      "create_cert_certificate",
			Replay:    tools.Fields("name", name, "certificate, certificate_chain, private_key", "(the same material, byte for byte)"),
			TokenNote: "This uploads exactly one certificate. The token authorizes uploading only this exact material",
		}.Render(token)), nil, nil
	})
}

func registerUpdateCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_cert_certificate",
		Description: "Rename a certificate." + renameNote +
			"The certificate body, chain and private key are immutable: to rotate the material, upload a new certificate with create_cert_certificate and repoint the load balancer at it, then delete the old one." + updatedNote,
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
		Description: "Delete one SSL/TLS certificate and its private key. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. This is irreversible — the private key is gone, so re-uploading needs the original key file. " +
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

// materialDigest binds a confirmation token to the exact material previewed.
func materialDigest(c certSDK.Certificate) string {
	sum := sha256.Sum256([]byte(c.Certificate + c.CertificateChain + c.PrivateKey))
	return hex.EncodeToString(sum[:])
}

// timeText renders an optional API timestamp for a preview field.
func timeText(t *certSDK.IonosTime) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
