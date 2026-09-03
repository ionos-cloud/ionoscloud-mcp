package cert

// Completion semantics differ per operation: POST answers 201 and PATCH 200 with
// the stored resource, while DELETE answers 202 and finishes in the background.
const (
	createdNote = " Synchronous (201): the returned body is the stored resource."

	updatedNote = " Synchronous (200): the returned body is the updated resource."

	deletedNote = " Asynchronous (202): the API has accepted the request; the resource is gone once its get_ tool answers 404."

	// stateNote names the field that says whether the resource is usable yet.
	stateNote = " Check metadata.state: AVAILABLE is ready, PROVISIONING still working, FAILED failed with the reason in metadata.message."

	// issuanceNote: a 201 means the renewal config exists, not that a cert was issued.
	issuanceNote = " Issuing the certificate happens after this call returns: poll get_cert_auto_certificate until metadata.lastIssuedCertificate names a certificate, then read it with get_cert_certificate."
)
