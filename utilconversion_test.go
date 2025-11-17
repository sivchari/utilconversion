package utilconversion

import (
	"testing"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestObject is a simple test object that implements metav1.Object.
type TestObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestObjectSpec   `json:"spec,omitempty,omitzero"`
	Status TestObjectStatus `json:"status,omitempty,omitzero"`
}

type TestObjectSpec struct {
	Field1 string `json:"field1,omitempty"`
	Field2 int    `json:"field2,omitempty"`
}

type TestObjectStatus struct {
	Ready bool `json:"ready,omitempty"`
}

func TestMarshalData(t *testing.T) {
	t.Run("should marshal source object to destination annotations", func(t *testing.T) {
		g := gomega.NewWithT(t)

		src := &TestObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-object",
				Namespace: "default",
				Labels: map[string]string{
					"label1": "value1",
				},
			},
			Spec: TestObjectSpec{
				Field1: "test-value",
				Field2: 42,
			},
			Status: TestObjectStatus{
				Ready: true,
			},
		}

		dst := &TestObject{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-object-dst",
				Namespace: "default",
			},
		}

		// Marshal data
		g.Expect(MarshalData(src, dst)).To(gomega.Succeed())

		// Check that annotations were added
		g.Expect(dst.GetAnnotations()).To(gomega.HaveKey(DataAnnotation))

		// Check that the annotation contains spec data
		annotationData := dst.GetAnnotations()[DataAnnotation]
		g.Expect(annotationData).To(gomega.ContainSubstring("test-value"))
		g.Expect(annotationData).To(gomega.ContainSubstring("42"))

		// Check that metadata is not included
		g.Expect(annotationData).ToNot(gomega.ContainSubstring("test-object"))
		g.Expect(annotationData).ToNot(gomega.ContainSubstring("label1"))
	})

	t.Run("should preserve existing annotations", func(t *testing.T) {
		g := gomega.NewWithT(t)

		src := &TestObject{
			Spec: TestObjectSpec{
				Field1: "test",
			},
		}

		dst := &TestObject{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"existing-annotation": "existing-value",
				},
			},
		}

		// Marshal data
		g.Expect(MarshalData(src, dst)).To(gomega.Succeed())

		// Check that both annotations exist
		annotations := dst.GetAnnotations()
		g.Expect(annotations).To(gomega.HaveKeyWithValue("existing-annotation", "existing-value"))
		g.Expect(annotations).To(gomega.HaveKey(DataAnnotation))
	})

	t.Run("should handle empty source object", func(t *testing.T) {
		g := gomega.NewWithT(t)

		src := &TestObject{}
		dst := &TestObject{}

		// Marshal data should succeed even with empty source
		g.Expect(MarshalData(src, dst)).To(gomega.Succeed())
		g.Expect(dst.GetAnnotations()).To(gomega.HaveKey(DataAnnotation))
	})
}

func TestUnmarshalData(t *testing.T) {
	t.Run("should return false when annotation doesn't exist", func(t *testing.T) {
		g := gomega.NewWithT(t)

		from := &TestObject{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"other-annotation": "value",
				},
			},
		}

		to := &TestObject{}

		found, err := UnmarshalData(from, to)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(found).To(gomega.BeFalse())
	})

	t.Run("should unmarshal data from annotation", func(t *testing.T) {
		g := gomega.NewWithT(t)

		// First marshal to get valid annotation data
		src := &TestObject{
			Spec: TestObjectSpec{
				Field1: "test-value",
				Field2: 99,
			},
			Status: TestObjectStatus{
				Ready: true,
			},
		}

		intermediate := &TestObject{}
		g.Expect(MarshalData(src, intermediate)).To(gomega.Succeed())

		// Now unmarshal
		to := &TestObject{}
		found, err := UnmarshalData(intermediate, to)
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(found).To(gomega.BeTrue())

		// Check that data was correctly unmarshaled
		g.Expect(to.Spec.Field1).To(gomega.Equal("test-value"))
		g.Expect(to.Spec.Field2).To(gomega.Equal(99))
		g.Expect(to.Status.Ready).To(gomega.BeTrue())
	})

	t.Run("should remove annotation after unmarshaling", func(t *testing.T) {
		g := gomega.NewWithT(t)

		src := &TestObject{
			Spec: TestObjectSpec{
				Field1: "test",
			},
		}

		intermediate := &TestObject{}
		g.Expect(MarshalData(src, intermediate)).To(gomega.Succeed())

		// Verify annotation exists before unmarshaling
		g.Expect(intermediate.GetAnnotations()).To(gomega.HaveKey(DataAnnotation))

		// Unmarshal
		to := &TestObject{}
		_, err := UnmarshalData(intermediate, to)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		// Verify annotation was removed
		g.Expect(intermediate.GetAnnotations()).ToNot(gomega.HaveKey(DataAnnotation))
	})

	t.Run("should preserve other annotations after unmarshaling", func(t *testing.T) {
		g := gomega.NewWithT(t)

		src := &TestObject{
			Spec: TestObjectSpec{
				Field1: "test",
			},
		}

		intermediate := &TestObject{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"other-annotation": "other-value",
				},
			},
		}
		g.Expect(MarshalData(src, intermediate)).To(gomega.Succeed())

		// Unmarshal
		to := &TestObject{}
		_, err := UnmarshalData(intermediate, to)
		g.Expect(err).ToNot(gomega.HaveOccurred())

		// Verify other annotation is preserved
		annotations := intermediate.GetAnnotations()
		g.Expect(annotations).To(gomega.HaveKeyWithValue("other-annotation", "other-value"))
		g.Expect(annotations).ToNot(gomega.HaveKey(DataAnnotation))
	})

	t.Run("should return error for invalid JSON", func(t *testing.T) {
		g := gomega.NewWithT(t)

		from := &TestObject{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					DataAnnotation: "invalid-json",
				},
			},
		}

		to := &TestObject{}
		found, err := UnmarshalData(from, to)
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(found).To(gomega.BeFalse())
	})
}
