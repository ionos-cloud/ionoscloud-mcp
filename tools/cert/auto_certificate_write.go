package cert

import (
	"context"
	"strings"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterAutoCertificateWriteTools registers the create/update/delete
// auto-certificate tools.
func RegisterAutoCertificateWriteTools(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateAutoCertificate(server, client, scope, confirm)
	registerUpdateAutoCertificate(server, client, scope)
	registerDeleteAutoCertificate(server, client, scope, confirm)
}

func registerCreateAutoCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_cert_auto_certificate",
		Description: "Create one auto-certificate: a standing instruction to issue a certificate for a DNS name through a certificate provider and renew it before expiry. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to create it. Creates exactly one auto-certificate per call. " +
			"common_name and every subject_alternative_names entry must belong to a zone hosted in IONOS Cloud DNS, because the provider is validated through a DNS challenge — a name outside IONOS DNS fails to issue. Needs a provider from create_cert_provider." + createdNote + stateNote + issuanceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateCertAutoCertificateInput) (*mcp.CallToolResult, any, error) {
		providerID := strings.TrimSpace(input.ProviderID)
		commonName := strings.TrimSpace(input.CommonName)
		name := strings.TrimSpace(input.Name)
		switch {
		case providerID == "":
			return tools.ErrorText("provider_id is required; list the available providers with list_cert_providers"), nil, nil
		case commonName == "":
			return tools.ErrorText("common_name is required, e.g. www.example.com"), nil, nil
		case name == "":
			return tools.ErrorText("name is required to create an auto-certificate"), nil, nil
		}
		algorithm, msg := normalizeKeyAlgorithm(input.KeyAlgorithm)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		props := certSDK.AutoCertificate{
			Provider:                providerID,
			CommonName:              commonName,
			KeyAlgorithm:            algorithm,
			Name:                    name,
			SubjectAlternativeNames: cleanNames(input.SubjectAlternativeNames),
		}
		target := tools.Target(req, providerID, commonName, name, algorithm, strings.Join(props.SubjectAlternativeNames, ","))

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_cert_auto_certificate", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_cert_auto_certificate", "provider_id, common_name, name, key_algorithm and subject_alternative_names", err)), nil, nil
			}
			created, _, err := client.AutoCertificateApi.AutoCertificatesPost(ctx).AutoCertificateCreate(certSDK.AutoCertificateCreate{Properties: props}).Execute()
			return tools.ToResult(created, err)
		}

		// Resolving it here names the bad field instead of 422-ing on the execute call.
		provider, _, err := client.ProviderApi.ProvidersFindById(ctx, providerID).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(notFoundText("certificate provider", providerID, "issue certificates with") +
					"; list the available providers with list_cert_providers"), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("create_cert_auto_certificate", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one auto-certificate:",
			Fields: tools.Fields(
				"name", name,
				"common_name", commonName,
				"subject_alternative_names", strings.Join(props.SubjectAlternativeNames, ", "),
				"key_algorithm", algorithm,
				"provider_id", providerID,
				"provider", provider.Properties.Name,
				"provider_server", provider.Properties.Server,
			),
			Tool:      "create_cert_auto_certificate",
			Replay:    tools.Fields("provider_id", providerID, "common_name", commonName, "name", name),
			TokenNote: "This creates exactly one auto-certificate. The token authorizes creating only this provider_id, common_name, name, key_algorithm and subject_alternative_names",
		}.Render(token)), nil, nil
	})
}

func registerUpdateAutoCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_cert_auto_certificate",
		Description: "Rename an auto-certificate." + renameNote +
			"The provider, common name, subject alternative names and key algorithm are immutable: to change what is issued, delete this auto-certificate and create a new one." + updatedNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateCertAutoCertificateInput) (*mcp.CallToolResult, any, error) {
		return rename(input.AutoCertificateID, "auto_certificate_id", input.Name,
			func(id string, props certSDK.PatchName) (certSDK.AutoCertificateRead, error) {
				updated, _, err := client.AutoCertificateApi.AutoCertificatesPatch(ctx, id).AutoCertificatePatch(certSDK.AutoCertificatePatch{Properties: props}).Execute()
				return updated, err
			})
	})
}

func registerDeleteAutoCertificate(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_cert_auto_certificate",
		Description: "Delete one auto-certificate, stopping automatic renewal for its DNS name. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. This is irreversible. " +
			"The certificates it has already issued keep working until they expire, and nothing renews them afterwards — so a load balancer using one starts serving an expired certificate on that date. Verify with list_cert_certificates once the delete is through." + deletedNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteCertAutoCertificateInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.AutoCertificateID)
		if id == "" {
			return tools.ErrorText("auto_certificate_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_cert_auto_certificate", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_cert_auto_certificate", "auto_certificate_id", err)), nil, nil
			}
			if _, err := client.AutoCertificateApi.AutoCertificatesDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("auto-certificate", id) +
				" get_cert_auto_certificate answers 404 once it is gone."), nil, nil
		}

		autoCert, _, err := client.AutoCertificateApi.AutoCertificatesFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(notFoundText("auto-certificate", id, "delete")), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		count, capped, countErr := certificatesIssuedBy(ctx, client, id)
		radius := tools.AffectedRadius()
		radius.Add("certificates it issued (they keep working until they expire, then nothing renews them)", count)

		token, mErr := confirm.Mint("delete_cert_auto_certificate", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to DELETE an auto-certificate, stopping renewal for its DNS name. This is IRREVERSIBLE." +
			tools.CappedCountNote(capped, "certificate", certListLimit)
		emptyNote := "This auto-certificate has issued no certificates yet."
		if unreadable := tools.IncompleteRadiusNote(tools.ErrLabel(countErr, "issued certificates")); unreadable != "" {
			headline += unreadable
			emptyNote = "" // an unreadable collection must not read as an empty one
		}
		cp := autoCert.Properties
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"auto_certificate_id", id,
				"name", cp.Name,
				"common_name", cp.CommonName,
				"subject_alternative_names", strings.Join(cp.SubjectAlternativeNames, ", "),
				"key_algorithm", cp.KeyAlgorithm,
				"provider_id", cp.Provider,
				"state", autoCert.Metadata.State,
				"last_issued_certificate", tools.OptStr(autoCert.Metadata.LastIssuedCertificate),
			),
			Radius:    radius,
			EmptyNote: emptyNote,
			Tool:      "delete_cert_auto_certificate",
			Replay:    tools.Fields("auto_certificate_id", id),
			TokenNote: "This token authorizes deleting ONLY this auto-certificate",
		}.Render(token)), nil, nil
	})
}
