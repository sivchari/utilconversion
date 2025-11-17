package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// GroupName is the group name for this API
	GroupName = "example.utilconversion.io"

	// Version is the version for this API
	Version = "v1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}
