package dns

// Completion semantics differ per operation, so there is no single note.
const (
	asyncZoneNote = " Asynchronous (202): poll get_dns_zone, get_dns_secondary_zone or get_dns_record for metadata.state — AVAILABLE is done, PROVISIONING/DESTROYING still working, FAILED failed. Leave a few seconds between polls."

	asyncTransferNote = " Asynchronous (202): follow get_dns_secondary_zone_axfr, which reports the transfer status of each primary IP separately along with an errorMessage when one fails."

	acceptedNoPollNote = " Asynchronous (202): the API has accepted the request. A reverse record carries no provisioning state, so there is nothing to poll — confirm with list_dns_reverse_records."

	syncNote = " Synchronous: the returned record is the final state, so no polling is needed."
)
