package k8s

// Every mutating Kubernetes endpoint answers 202, so the tool returns before the
// change has taken effect. Appended to each write tool's description.
const (
	asyncResourceNote = " Asynchronous (202): poll get_k8s_cluster or get_k8s_nodepool for metadata.state — ACTIVE/AVAILABLE is done, DEPLOYING/UPDATING/BUSY still working, FAILED_* failed. Leave 30s between polls; a BUSY resource queues further changes rather than rejecting them."

	asyncNodeNote = " Asynchronous (202): follow list_k8s_nodepool_nodes, whose state values are PROVISIONING, PROVISIONED, READY, TERMINATING, REBUILDING and BUSY. Leave 30s between polls."
)
