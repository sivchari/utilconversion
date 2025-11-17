package v1alpha1

import (
	"testing"

	"github.com/sivchari/utilconversion"
	v1 "github.com/sivchari/utilconversion/_example/api/v1"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func TestFuzzyConversion(t *testing.T) {
	// Register our types with the scheme
	testScheme := runtime.NewScheme()

	// Register our custom types
	testScheme.AddKnownTypes(
		v1.SchemeGroupVersion,
		&v1.MyResource{},
	)
	testScheme.AddKnownTypes(
		SchemeGroupVersion,
		&MyResource{},
	)

	t.Run("v1alpha1 MyResource", utilconversion.FuzzTestFunc(&utilconversion.FuzzTestFuncInput{
		Scheme: testScheme,
		Hub:    &v1.MyResource{},
		Spoke:  &MyResource{},
		SpokeAfterMutation: func(spoke conversion.Convertible) {
			resource := spoke.(*MyResource)
			// nil → "" → &"" の変換を正規化
			if resource.Spec.OldField != nil && *resource.Spec.OldField == "" {
				resource.Spec.OldField = nil
			}
			// Remove empty annotations map if it exists
			if len(resource.GetAnnotations()) == 0 {
				resource.SetAnnotations(nil)
			}
		},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{FuzzerFuncs},
	}))
}
