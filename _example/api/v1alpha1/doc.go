// Package v1alpha1 contains API Schema definitions for the example v1alpha1 API group
// +k8s:deepcopy-gen=package
// +k8s:conversion-gen=github.com/sivchari/utilconversion/_example/api/v1
// +k8s:defaulter-gen=TypeMeta
// +groupName=example.utilconversion.io
package v1alpha1

import runtime "k8s.io/apimachinery/pkg/runtime"

// SchemeBuilder is used to add go types to the GroupVersionKind scheme
var (
	localSchemeBuilder = runtime.NewSchemeBuilder()
	AddToScheme        = localSchemeBuilder.AddToScheme
)
