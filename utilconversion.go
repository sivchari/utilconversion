package utilconversion

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"sigs.k8s.io/randfill"
)

const (
	// DataAnnotation is the annotation that conversion webhooks
	// use to retain the data in case of down-conversion from the hub.
	DataAnnotation = "cluster.x-k8s.io/conversion-data"
)

// MarshalData stores the source object as json data in the destination object annotations map.
// It ignores the metadata of the source object.
func MarshalData(src, dst metav1.Object) error {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(src)
	if err != nil {
		return fmt.Errorf("failed to convert source to unstructured: %w", err)
	}

	delete(u, "metadata")

	data, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("failed to marshal source object: %w", err)
	}

	annotations := dst.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[DataAnnotation] = string(data)
	dst.SetAnnotations(annotations)

	return nil
}

// UnmarshalData tries to retrieve the data from the annotation and unmarshals it into the object passed as input.
func UnmarshalData(from metav1.Object, to any) (bool, error) {
	annotations := from.GetAnnotations()

	data, ok := annotations[DataAnnotation]
	if !ok {
		return false, nil
	}

	if err := json.Unmarshal([]byte(data), to); err != nil {
		return false, fmt.Errorf("failed to unmarshal data annotation: %w", err)
	}

	delete(annotations, DataAnnotation)
	from.SetAnnotations(annotations)

	return true, nil
}

// GetFuzzer returns a new fuzzer to be used for testing.
func GetFuzzer(scheme *runtime.Scheme, funcs ...fuzzer.FuzzerFuncs) *randfill.Filler {
	funcs = append([]fuzzer.FuzzerFuncs{
		metafuzzer.Funcs,
		func(_ runtimeserializer.CodecFactory) []any {
			return []any{
				func(input **metav1.Time, c randfill.Continue) {
					if c.Bool() {
						return
					}
					if c.Bool() {
						*input = &metav1.Time{}

						return
					}
					var sec, nsec uint32
					c.Fill(&sec)
					c.Fill(&nsec)
					fuzzed := metav1.Unix(int64(sec), int64(nsec)).Rfc3339Copy()
					*input = &metav1.Time{Time: fuzzed.Time}
				},
				func(in **intstr.IntOrString, c randfill.Continue) {
					if c.Bool() {
						return
					}
					if c.Bool() {
						*in = &intstr.IntOrString{}

						return
					}
					*in = ptr.To(intstr.FromInt32(c.Int31n(50)))
				},
			}
		},
	}, funcs...)

	return fuzzer.FuzzerFor(
		fuzzer.MergeFuzzerFuncs(funcs...),
		//nolint:gosec // fuzzer uses math/rand for generating random data
		rand.NewSource(rand.Int63()),
		runtimeserializer.NewCodecFactory(scheme),
	)
}

// FuzzTestFuncInput contains input parameters
// for the FuzzTestFunc function.
type FuzzTestFuncInput struct {
	Scheme *runtime.Scheme

	Hub              conversion.Hub
	HubAfterMutation func(conversion.Hub)

	Spoke                      conversion.Convertible
	SpokeAfterMutation         func(conversion.Convertible)
	SkipSpokeAnnotationCleanup bool

	FuzzerFuncs []fuzzer.FuzzerFuncs

	N *int
}

// FuzzTestFunc returns a new testing function to be used in tests to make sure conversions between
// the Hub version of an object and an older version aren't lossy.
func FuzzTestFunc(input *FuzzTestFuncInput) func(*testing.T) {
	if input.Scheme == nil {
		input.Scheme = scheme.Scheme
	}

	return func(t *testing.T) {
		t.Helper()
		t.Run("spoke-hub-spoke", func(t *testing.T) {
			g := gomega.NewWithT(t)
			fuzzer := GetFuzzer(input.Scheme, input.FuzzerFuncs...)

			for range ptr.Deref(input.N, 10000) {
				spokeBefore, ok := input.Spoke.DeepCopyObject().(conversion.Convertible)
				if !ok {
					t.Fatalf("input.Spoke does not implement conversion.Convertible")
				}

				fuzzer.Fill(spokeBefore)

				hubCopy, ok := input.Hub.DeepCopyObject().(conversion.Hub)
				if !ok {
					t.Fatalf("input.Hub does not implement conversion.Hub")
				}

				g.Expect(spokeBefore.ConvertTo(hubCopy)).To(gomega.Succeed())

				spokeAfter, ok := input.Spoke.DeepCopyObject().(conversion.Convertible)
				if !ok {
					t.Fatalf("input.Spoke does not implement conversion.Convertible")
				}

				g.Expect(spokeAfter.ConvertFrom(hubCopy)).To(gomega.Succeed())

				// Remove data annotation eventually added by ConvertFrom for avoiding data loss in hub-spoke-hub round trips
				// NOTE: There are use case when we want to skip this operation, e.g. if the spoke object does not have ObjectMeta (e.g. kubeadm types).
				if !input.SkipSpokeAnnotationCleanup {
					metaAfter := spokeAfter.(metav1.Object)
					delete(metaAfter.GetAnnotations(), DataAnnotation)
				}

				if input.SpokeAfterMutation != nil {
					input.SpokeAfterMutation(spokeAfter)
				}

				if !apiequality.Semantic.DeepEqual(spokeBefore, spokeAfter) {
					diff := cmp.Diff(spokeBefore, spokeAfter)
					g.Expect(false).To(gomega.BeTrue(), diff)
				}
			}
		})
		t.Run("hub-spoke-hub", func(t *testing.T) {
			g := gomega.NewWithT(t)
			fuzzer := GetFuzzer(input.Scheme, input.FuzzerFuncs...)

			for range ptr.Deref(input.N, 10000) {
				hubBefore, ok := input.Hub.DeepCopyObject().(conversion.Hub)
				if !ok {
					t.Fatalf("input.Hub does not implement conversion.Hub")
				}

				fuzzer.Fill(hubBefore)

				dstCopy, ok := input.Spoke.DeepCopyObject().(conversion.Convertible)
				if !ok {
					t.Fatalf("input.Spoke does not implement conversion.Convertible")
				}

				g.Expect(dstCopy.ConvertFrom(hubBefore)).To(gomega.Succeed())

				hubAfter, ok := input.Hub.DeepCopyObject().(conversion.Hub)
				if !ok {
					t.Fatalf("input.Hub does not implement conversion.Hub")
				}

				g.Expect(dstCopy.ConvertTo(hubAfter)).To(gomega.Succeed())

				if !apiequality.Semantic.DeepEqual(hubBefore, hubAfter) {
					diff := cmp.Diff(hubBefore, hubAfter)
					g.Expect(false).To(gomega.BeTrue(), diff)
				}
			}
		})
	}
}
