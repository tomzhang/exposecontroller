package v1

import (
	"k8s.io/kubernetes/pkg/conversion"
	"k8s.io/kubernetes/pkg/runtime"

	oapi "github.com/openshift/origin/pkg/api"
	newer "github.com/openshift/origin/pkg/template/api"
)

func convert_api_Template_To_v1_Template(in *newer.Template, out *Template, s conversion.Scope) error {
	//FIXME: DefaultConvert should not overwrite the Labels field on the
	//       the base object. This is likely a bug in the DefaultConvert
	//       code. For now, it is called before converting the labels.
	if err := s.DefaultConvert(in, out, conversion.IgnoreMissingFields); err != nil {
		return err
	}
	if err := s.Convert(&in.ObjectLabels, &out.Labels, 0); err != nil {
		return err
	}

	// if we have runtime.Unstructured objects from the Process call.  We need to encode those
	// objects using the unstructured codec BEFORE the REST layers gets its shot at encoding to avoid a layered
	// encode being done.
	for i := range in.Objects {
		if unstructured, ok := in.Objects[i].(*runtime.Unstructured); ok {
			bytes, err := runtime.Encode(runtime.UnstructuredJSONScheme, unstructured)
			if err != nil {
				return err
			}
			out.Objects[i] = runtime.RawExtension{RawJSON: bytes}
		}
	}
	return nil
}

func convert_v1_Template_To_api_Template(in *Template, out *newer.Template, s conversion.Scope) error {
	if err := s.Convert(&in.Labels, &out.ObjectLabels, 0); err != nil {
		return err
	}
	return s.DefaultConvert(in, out, conversion.IgnoreMissingFields)
}

func addConversionFuncs(scheme *runtime.Scheme) {
	err := scheme.AddConversionFuncs(
		convert_api_Template_To_v1_Template,
		convert_v1_Template_To_api_Template,
	)
	if err != nil {
		panic(err)
	}

	if err := scheme.AddFieldLabelConversionFunc("v1", "Template",
		oapi.GetFieldLabelConversionFunc(newer.TemplateToSelectableFields(&newer.Template{}), nil),
	); err != nil {
		panic(err)
	}
}
