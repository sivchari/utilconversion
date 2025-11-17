package v1alpha1

import (
	v1 "github.com/sivchari/utilconversion/_example/api/v1"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/randfill"
)

// Fuzzers returns the list of custom fuzzers for testing.
func Fuzzers() []any {
	return []any{
		// Fuzzer for v1alpha1.MyResourceSpec
		func(in *MyResourceSpec, c randfill.Continue) {
			// Fill common fields first
			c.Fill(&in.Name)
			c.Fill(&in.Replicas)

			// Ensure OldField is sometimes set, sometimes nil
			// Never generate empty string to avoid nil vs &"" mismatch
			if c.Bool() {
				var str string
				c.Fill(&str)
				if str != "" {
					in.OldField = &str
				} else {
					in.OldField = nil
				}
			} else {
				in.OldField = nil
			}
		},

		// Fuzzer for v1.MyResourceSpec
		func(in *v1.MyResourceSpec, c randfill.Continue) {
			// Fill all fields
			c.Fill(&in.Name)
			c.Fill(&in.Replicas)
			c.Fill(&in.NewField)
		},

		// Fuzzer for v1alpha1.MyResourceStatus
		func(in *MyResourceStatus, c randfill.Continue) {
			c.Fill(in)

			// Ensure Ready is set to a valid bool value
			in.Ready = c.Bool()
		},

		// Fuzzer for v1.MyResourceStatus
		func(in *v1.MyResourceStatus, c randfill.Continue) {
			c.Fill(in)

			// Ensure Ready is sometimes set, sometimes nil
			if c.Bool() {
				in.Ready = ptr.To(c.Bool())
			} else {
				in.Ready = nil
			}
		},
	}
}

// FuzzerFuncs returns fuzzer functions for the v1alpha1 API group.
func FuzzerFuncs(_ runtimeserializer.CodecFactory) []any {
	return Fuzzers()
}
