package cert

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterProviderWriteTools registers the create/update/delete provider tools.
func RegisterProviderWriteTools(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateProvider(server, client, scope, confirm)
	registerUpdateProvider(server, client, scope)
	registerDeleteProvider(server, client, scope, confirm)
}

func registerCreateProvider(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_cert_provider",
		Description: "Register one ACME certificate provider (a certificate authority) that auto-certificates can issue and renew certificates through. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to create it. Creates exactly one provider per call. " +
			"server is the CA's ACME directory URL, e.g. https://acme-v02.api.letsencrypt.org/directory. Supply external_account_binding only for a CA that requires a pre-registered account (ZeroSSL, Google Trust Services); its secret is write-only and is never echoed in a preview or returned by a read tool." + createdNote + stateNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateCertProviderInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return tools.ErrorText("name is required to create a certificate provider"), nil, nil
		}
		email, msg := validateEmail(input.Email)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		acmeServer, msg := validateACMEServer(input.Server)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		props := certSDK.Provider{Name: name, Email: email, Server: acmeServer}
		if eab := input.ExternalAccountBinding; eab != nil {
			keyID, keySecret := strings.TrimSpace(eab.KeyID), strings.TrimSpace(eab.KeySecret)
			if keyID == "" || keySecret == "" {
				return tools.ErrorText("external_account_binding needs both key_id and key_secret; omit the whole object for a provider that does not use one"), nil, nil
			}
			props.ExternalAccountBinding = &certSDK.ProviderExternalAccountBinding{KeyId: &keyID, KeySecret: &keySecret}
		}
		var eabKeyID, eabKeySecret *string
		if eab := props.ExternalAccountBinding; eab != nil {
			eabKeyID, eabKeySecret = eab.KeyId, eab.KeySecret
		}
		// external_account_binding determines which CA account issuance is billed/
		// authorized against, so the token must bind it too
		var eabSecretDigest string
		if eabKeySecret != nil {
			sum := sha256.Sum256([]byte(*eabKeySecret))
			eabSecretDigest = hex.EncodeToString(sum[:])
		}
		target := tools.Target(req, name, acmeServer, email, tools.OptStr(eabKeyID), eabSecretDigest)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_cert_provider", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_cert_provider", "name, email, server and external_account_binding", err)), nil, nil
			}
			created, _, err := client.ProviderApi.ProvidersPost(ctx).ProviderCreate(certSDK.ProviderCreate{Properties: props}).Execute()
			return tools.ToResult(redactProvider(created), err)
		}

		token, err := confirm.Mint("create_cert_provider", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one certificate provider:",
			Fields: tools.Fields(
				"name", name,
				"email", email,
				"server", acmeServer,
				"external_account_binding.key_id", tools.OptStr(eabKeyID),
				"external_account_binding.key_secret", tools.Redacted(eabKeySecret),
			),
			Tool:      "create_cert_provider",
			Replay:    tools.Fields("name", name, "email", email, "server", acmeServer),
			TokenNote: "This creates exactly one provider. The token authorizes creating only this name, email, server and external_account_binding",
		}.Render(token)), nil, nil
	})
}

func registerUpdateProvider(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_cert_provider",
		Description: "Rename a certificate provider." + renameNote +
			"The email, the ACME directory URL and the external account binding are immutable: to change any of them, create a new provider, repoint the auto-certificates at it, and delete this one." + updatedNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateCertProviderInput) (*mcp.CallToolResult, any, error) {
		return rename(input.ProviderID, "provider_id", input.Name,
			func(id string, props certSDK.PatchName) (certSDK.ProviderRead, error) {
				updated, _, err := client.ProviderApi.ProvidersPatch(ctx, id).ProviderPatch(certSDK.ProviderPatch{Properties: props}).Execute()
				return redactProvider(updated), err
			})
	})
}

func registerDeleteProvider(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_cert_provider",
		Description: "Delete one ACME certificate provider. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. This is irreversible, and the external account binding secret cannot be recovered. " +
			"Auto-certificates issue through a provider, so any that name this one lose the ability to renew. The preview counts them; move them to another provider first by recreating them." + deletedNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteCertProviderInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ProviderID)
		if id == "" {
			return tools.ErrorText("provider_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_cert_provider", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_cert_provider", "provider_id", err)), nil, nil
			}
			if _, err := client.ProviderApi.ProvidersDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("certificate provider", id) +
				" get_cert_provider answers 404 once it is gone."), nil, nil
		}

		provider, _, err := client.ProviderApi.ProvidersFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(notFoundText("certificate provider", id, "delete")), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		count, capped, countErr := autoCertificatesUsing(ctx, client, id)
		radius := tools.AffectedRadius()
		radius.Add("auto-certificates issuing through it (they stop being able to renew)", count)

		token, mErr := confirm.Mint("delete_cert_provider", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to DELETE a certificate provider. This is IRREVERSIBLE." +
			tools.CappedCountNote(capped, "auto-certificate", certListLimit)
		emptyNote := "No auto-certificate issues through this provider."
		switch {
		case countErr != nil:
			headline += tools.IncompleteRadiusNote(tools.ErrLabel(countErr, "auto-certificates"))
			emptyNote = "" // an unreadable collection must not read as an empty one
		case capped && count == 0:
			headline += "\nWARNING: the first page of auto-certificates had no match for this provider, but more auto-certificates exist beyond it (the API has no provider filter to check them) -- this may affect more auto-certificates than shown."
			emptyNote = ""
		}
		cp := provider.Properties
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"provider_id", id,
				"name", cp.Name,
				"email", cp.Email,
				"server", cp.Server,
				"state", provider.Metadata.State,
			),
			Radius:    radius,
			EmptyNote: emptyNote,
			Tool:      "delete_cert_provider",
			Replay:    tools.Fields("provider_id", id),
			TokenNote: "This token authorizes deleting ONLY this provider",
		}.Render(token)), nil, nil
	})
}
