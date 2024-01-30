package featureserver

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/TheAschr/arcgis"
)

func ValidateFeature(feature interface{}, info Info) error {
	fType := reflect.TypeOf(feature)

	if fType.Kind() != reflect.Struct {
		return fmt.Errorf("feature must be a struct")
	}

	structFieldNames := make(map[string]string)

	for i := 0; i < fType.NumField(); i++ {
		f := fType.Field(i)
		jsonTag := strings.Split(f.Tag.Get("json"), ",")[0]
		if jsonTag == "" {
			return fmt.Errorf("missing json tag for '%s'", f.Name)
		}
		structFieldNames[jsonTag] = f.Name
	}

	if structFieldNames["attributes"] == "" {
		return fmt.Errorf("missing attributes field")
	}
	attributes, ok := fType.FieldByName(structFieldNames["attributes"])
	if !ok {
		return fmt.Errorf("missing attributes field")
	}

	if structFieldNames["geometry"] == "" {
		return fmt.Errorf("missing geometry field")
	}
	geometry, ok := fType.FieldByName(structFieldNames["geometry"])
	if !ok {
		return fmt.Errorf("missing geometry field")
	}

	var fields []FieldInfo
	expectGeometryType := GeometryTypeNone

	switch info := info.(type) {
	case FeatureLayerInfo:
		fields = info.Fields
		expectGeometryType = info.GeometryType
	case TableInfo:
		fields = info.Fields
	default:
		return fmt.Errorf("unhandled info type: %T", info)
	}

	// Validate geometry

	if geometry.Type.Kind() != reflect.Struct {
		return fmt.Errorf("geometry must be a struct but is %s", geometry.Type.Kind())
	}

	geometryType := geometry.Type.String()
	switch expectGeometryType {
	case GeometryTypeNone:
		if geometryType != "featureserver.GeometryNone" {
			return fmt.Errorf("geometry type is none but geometry is not")
		}
	case GeometryTypePoint:
		if geometryType != "featureserver.GeometryPoint" {
			return fmt.Errorf("geometry type is point but geometry is not")
		}
	case GeometryTypeMultiPoint:
		if geometryType != "featureserver.GeometryMultiPoint" {
			return fmt.Errorf("invalid geometry, expected '%s' but got '%s'", expectGeometryType, geometryType)
		}
	default:
		return fmt.Errorf("unhandled geometry type: %s", geometryType)
	}

	// Validate attributes

	structFieldNames = make(map[string]string)

	for i := 0; i < attributes.Type.NumField(); i++ {
		f := attributes.Type.Field(i)
		jsonTag := strings.Split(f.Tag.Get("json"), ",")[0]
		if jsonTag == "" {
			return fmt.Errorf("missing json tag for '%s'", f.Name)
		}
		structFieldNames[jsonTag] = f.Name
	}

	for _, field := range fields {
		structFieldName, ok := structFieldNames[field.Name]
		if !ok {
			// Skip fields checking fields that are not in the struct
			continue
		}
		f, ok := attributes.Type.FieldByName(structFieldName)
		if !ok {
			return fmt.Errorf("missing field '%s'", field.Name)
		}

		var expectType any

		switch field.Type {
		case FieldTypeOID:
			expectType = int32(0)
		case FieldTypeSmallInt:
			expectType = int16(0)
		case FieldTypeInt:
			expectType = int32(0)
		case FieldTypeFloat:
			expectType = float32(0)
		case FieldTypeDouble:
			expectType = float64(0)
		case FieldTypeString:
			expectType = ""
		case FieldTypeDate:
			expectType = arcgis.Date{}
		default:
			return fmt.Errorf("unhandled field type: %s", field.Type)
		}

		expectReflectType := reflect.TypeOf(expectType)
		if field.Nullable {
			expectReflectType = reflect.PointerTo(expectReflectType)
		}

		if f.Type != expectReflectType {
			return fmt.Errorf("field '%s' has type '%s' but expected type '%s'", field.Name, f.Type, expectReflectType)
		}
	}

	return nil
}
