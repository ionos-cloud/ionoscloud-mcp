package tools

// Storage performance classes: ESSENTIAL, BALANCED and PERFORMANCE replace the
// legacy HDD, SSD Standard and SSD Premium, which stay valid but are deprecated.
const (
	StorageEssential   = "ESSENTIAL"
	StorageBalanced    = "BALANCED"
	StoragePerformance = "PERFORMANCE"
)

// VolumeStorageTypes is VolumeProperties.type minus ISO, which is not a type you
// create a volume as.
var VolumeStorageTypes = []string{
	StorageEssential, StorageBalanced, StoragePerformance,
	"HDD", "SSD", "SSD Standard", "SSD Premium", "DAS",
}

// K8sStorageTypes is what the node pool API accepts. Verified live: it answers 422
// for BALANCED and for SSD Premium, which the published spec wrongly lists.
var K8sStorageTypes = []string{
	StorageEssential, StoragePerformance, "HDD", "SSD",
}

const (
	VolumeStorageTypeList = "ESSENTIAL, BALANCED or PERFORMANCE (the performance classes), or the legacy HDD, SSD, SSD Standard, SSD Premium"
	K8sStorageTypeList    = "ESSENTIAL or PERFORMANCE (the performance classes), or the legacy HDD, SSD"
)
