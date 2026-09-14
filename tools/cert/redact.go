package cert

import (
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
)

// The private key and the external account binding secret are write-only in the spec
// but modelled by the SDK anyway, and a tool result is transcript the client keeps.
// Every cert handler returns through these, so a new secret field has one place to go.

func redactCertificate(c certSDK.CertificateRead) certSDK.CertificateRead {
	c.Properties.PrivateKey = ""
	return c
}

func redactCertificateList(l certSDK.CertificateReadList) certSDK.CertificateReadList {
	for i := range l.Items {
		l.Items[i] = redactCertificate(l.Items[i])
	}
	return l
}

func redactProvider(p certSDK.ProviderRead) certSDK.ProviderRead {
	if eab := p.Properties.ExternalAccountBinding; eab != nil {
		// A fresh struct: the pointer is shared with the caller's copy of the response.
		p.Properties.ExternalAccountBinding = &certSDK.ProviderExternalAccountBinding{KeyId: eab.KeyId}
	}
	return p
}

func redactProviderList(l certSDK.ProviderReadList) certSDK.ProviderReadList {
	for i := range l.Items {
		l.Items[i] = redactProvider(l.Items[i])
	}
	return l
}
