package main

import (
	"fmt"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/lastmilelink/protoc-gen-jsonschema/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

const (
	TypeNameTimestamp = ".google.protobuf.Timestamp"

	PatternTimestamp = `^[0-9]{4}-[0-2][0-9]-[0-3][0-9]T[0-2][0-9]:[0-6][0-9]:[0-6][0-9].[0-9]{3}Z$`
)

var (
	enumMap = make(map[string]*descriptor.EnumDescriptorProto)
)

func convertTimestamp(desc *descriptor.FieldDescriptorProto, jsonSchemaType *jsonschema.Type) {
	if desc.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE &&
		desc.GetTypeName() == TypeNameTimestamp {

		jsonSchemaType.Type = gojsonschema.TYPE_STRING
		jsonSchemaType.Pattern = PatternTimestamp
	}
}

func findAllEnumTypes(req *plugin.CodeGeneratorRequest) {
	for _, file := range req.GetProtoFile() {
		for _, enum := range file.GetEnumType() {
			name := fmt.Sprintf(".%s.%s", file.GetPackage(), enum.GetName())
			enumMap[name] = enum
		}
	}
}

func convertEnum(desc *descriptor.FieldDescriptorProto, jsonSchemaType *jsonschema.Type) {
	if desc.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM {
		enum, ok := enumMap[desc.GetTypeName()]
		if !ok {
			logWithLevel(LOG_ERROR, "Failed to find enum type [%s]", desc.GetTypeName())
			return
		}

		var enums []interface{}
		for _, v := range enum.GetValue() {
			enums = append(enums, v.GetName())
		}

		jsonSchemaType.Type = gojsonschema.TYPE_STRING
		jsonSchemaType.Enum = enums
		jsonSchemaType.OneOf = nil
	}
}
