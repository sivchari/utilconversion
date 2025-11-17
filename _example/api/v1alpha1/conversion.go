package v1alpha1

import (
	"github.com/sivchari/utilconversion"
	v1 "github.com/sivchari/utilconversion/_example/api/v1"
	"k8s.io/apimachinery/pkg/conversion"
	ctrlconversion "sigs.k8s.io/controller-runtime/pkg/conversion"
)

func Convert_v1alpha1_MyResourceSpec_To_v1_MyResourceSpec(in *MyResourceSpec, out *v1.MyResourceSpec, s conversion.Scope) error {
	out.Name = in.Name
	out.Replicas = in.Replicas

	// OldField in v1alpha1 was renamed to NewField in v1
	if in.OldField != nil {
		out.NewField = *in.OldField
	}

	return nil
}

func Convert_v1_MyResourceSpec_To_v1alpha1_MyResourceSpec(in *v1.MyResourceSpec, out *MyResourceSpec, s conversion.Scope) error {
	out.Name = in.Name
	out.Replicas = in.Replicas

	// NewField in v1 was OldField in v1alpha1
	// Always convert, even if empty string
	out.OldField = &in.NewField

	return nil
}

// ConvertTo converts this v1alpha1 MyResource to the Hub version (v1).
func (src *MyResource) ConvertTo(dstRaw ctrlconversion.Hub) error {
	dst := dstRaw.(*v1.MyResource)

	if err := Convert_v1alpha1_MyResource_To_v1_MyResource(src, dst, nil); err != nil {
		return err
	}

	// Manually restore v1-specific fields from annotation
	restored := &v1.MyResource{}
	ok, err := utilconversion.UnmarshalData(src, restored)
	if err != nil {
		return err
	}

	if ok {
		// Annotation data was restored (hub→spoke→hub case)
		dst.Spec.NewField = restored.Spec.NewField
		dst.Status.Ready = restored.Status.Ready
		dst.Status.Conditions = restored.Status.Conditions
	}

	return nil
}

// ConvertFrom converts from the Hub version (v1) to this v1alpha1 MyResource.
func (dst *MyResource) ConvertFrom(srcRaw ctrlconversion.Hub) error {
	src := srcRaw.(*v1.MyResource)

	if err := Convert_v1_MyResource_To_v1alpha1_MyResource(src, dst, nil); err != nil {
		return err
	}

	// Preserve v1 data in annotations for the up-conversion (hub→spoke→hub)
	return utilconversion.MarshalData(src, dst)
}
