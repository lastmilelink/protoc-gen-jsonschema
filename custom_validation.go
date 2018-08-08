package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/lastmilelink/protoc-gen-jsonschema/jsonschema"
	"github.com/lyft/protoc-gen-validate/validate"
)

func findValidation(desc *descriptor.FieldDescriptorProto) (*validate.FieldRules, error) {
	if desc.GetOptions() == nil {
		return nil, nil
	}
	ext, err := proto.GetExtension(desc.GetOptions(), validate.E_Rules)
	if err != nil {
		return nil, err
	}
	if ext != nil {
		if v, ok := ext.(*validate.FieldRules); ok {
			return v, nil
		}
	}
	return nil, nil
}

func isRequired(desc *descriptor.FieldDescriptorProto, jsonSchemaType *jsonschema.Type) (bool, error) {
	ext, err := findValidation(desc)
	if err != nil {
		return false, err
	}

	if ext != nil {
		if r, ok := ext.Type.(*validate.FieldRules_Any); ok {
			return r.Any.GetRequired(), nil
		}

		if r, ok := ext.Type.(*validate.FieldRules_Message); ok {
			return r.Message.GetRequired(), nil
		}
	}

	return false, nil
}

func addRules(desc *descriptor.FieldDescriptorProto, jsonSchemaType *jsonschema.Type) error {
	ext, err := findValidation(desc)
	if err != nil {
		return err
	}

	if ext == nil {
		return nil
	}

	switch v := ext.Type.(type) {
	case *validate.FieldRules_Float:
		if v.Float.Gt != nil {
			jsonSchemaType.Minimum = intPnt(int(math.Round(float64(v.Float.GetGt()) - 0.5)))
			jsonSchemaType.ExclusiveMinimum = false
		} else if v.Float.Gte != nil {
			jsonSchemaType.Minimum = intPnt(int(math.Round(float64(v.Float.GetGte()) - 0.5)))
			jsonSchemaType.ExclusiveMinimum = true
		}
		if v.Float.Lt != nil {
			jsonSchemaType.Maximum = intPnt(int(math.Round(float64(v.Float.GetLt()) + 0.5)))
			jsonSchemaType.ExclusiveMaximum = false
		} else if v.Float.Lte != nil {
			jsonSchemaType.Maximum = intPnt(int(math.Round(float64(v.Float.GetLte()) + 0.5)))
			jsonSchemaType.ExclusiveMaximum = true
		}
		convertOneOfToType(jsonSchemaType)
	case *validate.FieldRules_Double:
		if v.Double.Gt != nil {
			jsonSchemaType.Minimum = intPnt(int(math.Round(v.Double.GetGt() - 0.5)))
			jsonSchemaType.ExclusiveMinimum = false
		} else if v.Double.Gte != nil {
			jsonSchemaType.Minimum = intPnt(int(math.Round(v.Double.GetGte() - 0.5)))
			jsonSchemaType.ExclusiveMinimum = true
		}
		if v.Double.Lt != nil {
			jsonSchemaType.Maximum = intPnt(int(math.Round(v.Double.GetLt() + 0.5)))
			jsonSchemaType.ExclusiveMaximum = false
		} else if v.Double.Lte != nil {
			jsonSchemaType.Maximum = intPnt(int(math.Round(v.Double.GetLte() + 0.5)))
			jsonSchemaType.ExclusiveMaximum = true
		}
		convertOneOfToType(jsonSchemaType)
	case *validate.FieldRules_Int32:
		if v.Int32.Gt != nil {
			jsonSchemaType.Minimum = intPnt(int(v.Int32.GetGt()))
			jsonSchemaType.ExclusiveMinimum = false
		} else if v.Int32.Gte != nil {
			jsonSchemaType.Minimum = intPnt(int(v.Int32.GetGte()))
			jsonSchemaType.ExclusiveMinimum = true
		}
		if v.Int32.Lt != nil {
			jsonSchemaType.Maximum = intPnt(int(v.Int32.GetLt()))
			jsonSchemaType.ExclusiveMaximum = false
		} else if v.Int32.Lte != nil {
			jsonSchemaType.Maximum = intPnt(int(v.Int32.GetLte()))
			jsonSchemaType.ExclusiveMaximum = true
		}
		convertOneOfToType(jsonSchemaType)
	case *validate.FieldRules_Int64:
		if v.Int64.Gt != nil {
			jsonSchemaType.Minimum = intPnt(int(v.Int64.GetGt()))
			jsonSchemaType.ExclusiveMinimum = false
		} else if v.Int64.Gte != nil {
			jsonSchemaType.Minimum = intPnt(int(v.Int64.GetGte()))
			jsonSchemaType.ExclusiveMinimum = true
		}
		if v.Int64.Lt != nil {
			jsonSchemaType.Maximum = intPnt(int(v.Int64.GetLt()))
			jsonSchemaType.ExclusiveMaximum = false
		} else if v.Int64.Lte != nil {
			jsonSchemaType.Maximum = intPnt(int(v.Int64.GetLte()))
			jsonSchemaType.ExclusiveMaximum = true
		}
		convertOneOfToType(jsonSchemaType)
	case *validate.FieldRules_String_:
		if v.String_.Pattern != nil {
			jsonSchemaType.Pattern = v.String_.GetPattern()
		}
		if v.String_.Prefix != nil {
			jsonSchemaType.Pattern = fmt.Sprintf(`^%s.*$`, v.String_.GetPrefix())
		}
		if v.String_.Suffix != nil {
			jsonSchemaType.Pattern = fmt.Sprintf(`^.*%s$`, v.String_.GetSuffix())
		}
		if v.String_.Contains != nil {
			jsonSchemaType.Pattern = fmt.Sprintf(`^.*%s.*$`, v.String_.GetSuffix())
		}
		if v.String_.In != nil && len(v.String_.In) > 0 {
			jsonSchemaType.Pattern = fmt.Sprintf(`^%s$`, strings.Join(v.String_.GetIn(), "|"))
		}
		if v.String_.MinLen != nil {
			jsonSchemaType.MinLength = int(v.String_.GetMinLen())
		}
		if v.String_.MaxLen != nil {
			jsonSchemaType.MaxLength = int(v.String_.GetMaxLen())
		}
		convertOneOfToType(jsonSchemaType)
	case *validate.FieldRules_Repeated:
		if v.Repeated.MinItems != nil {
			jsonSchemaType.MinItems = int(v.Repeated.GetMinItems())
		}
		if v.Repeated.MaxItems != nil {
			jsonSchemaType.MinItems = int(v.Repeated.GetMaxItems())
		}
		if v.Repeated.Unique != nil {
			jsonSchemaType.UniqueItems = v.Repeated.GetUnique()
		}
	}
	return nil
}

func intPnt(input int) *int {
	return &input
}

func convertOneOfToType(jsonSchemaType *jsonschema.Type) error {
	oneOf := jsonSchemaType.OneOf
	if len(oneOf) != 2 {
		return nil
	}
	for _, one := range oneOf {
		if one.Type != "null" {
			jsonSchemaType.Type = one.Type
			jsonSchemaType.OneOf = nil
			return nil
		}
	}
	return fmt.Errorf("Could not find valid type from oneOf, %+v", oneOf)
}
