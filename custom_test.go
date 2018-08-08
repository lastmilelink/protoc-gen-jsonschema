package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
)

const (
	ProtoFile = "otd_example.proto"

	MessageGood = `
		{
			"updatedAt": "2018-06-10T15:16:17.001Z",
			"updateBody": {
				"customerId": "hello",
				"age": 60,
				"accountId": "otd_hello",
				"subAccountId": "hello_otd",
				"income": 1500,
				"address": {
					"country": "GBR",
					"postcode": "EC2A 4PH",
					"town": "London",
					"streets": [
						"58-62 Scrutton St",
						"Shoreditch"
					]
				}, 
				"stops": [
					{
						"stopId": "stop1",
						"coordinates": {
							"longitude": 0.001,
							"latitude": 51.001
						}
					},
					{
						"stopId": "stop2",
						"coordinates": {
							"longitude": 0.002,
							"latitude": 51.002
						}
					}
				]
			}
		}
	`
	MessageNoAge = `
		{
			"updatedAt": "2018-06-10T15:16:17.001Z",
			"updateBody": {
				"customerId": "hello",
				"accountId": "otd_hello",
				"subAccountId": "hello_otd",
				"income": 1500,
				"address": {
					"country": "GBR",
					"postcode": "EC2A 4PH",
					"town": "London",
					"streets": [
						"58-62 Scrutton St",
						"Shoreditch"
					]
				},
				"stops": [
					{
						"stopId": "stop1",
						"coordinates": {
							"longitude": 0.001,
							"latitude": 51.001
						}
					},
					{
						"stopId": "stop2",
						"coordinates": {
							"longitude": 0.002,
							"latitude": 51.002
						}
					}
				]
			}
		}
	`

	MessageNoStreets = `
		{
			"updatedAt": "2018-06-10T15:16:17.001Z",
			"updateBody": {
				"customerId": "hello",
				"age": 60,
				"accountId": "otd_hello",
				"subAccountId": "hello_otd",
				"income": 1500,
				"address": {
					"country": "GBR",
					"postcode": "EC2A 4PH",
					"town": "London",
					"streets": [
						"58-62 Scrutton St",
						"Shoreditch"
					]
				},
				"stops": [
					{
						"stopId": "stop1",
						"coordinates": {
							"longitude": 0.001,
							"latitude": 51.001
						}
					},
					{
						"stopId": "stop2",
						"coordinates": {
							"longitude": 0.002,
							"latitude": 51.002
						}
					}
				]
			}
		}
	`
)

func TestCustomValidation(t *testing.T) {
	allowNullValues = true
	disallowBigIntsAsStrings = true
	disallowAdditionalProperties = true

	sampleProtos["OTD"] = SampleProto{
		AllowNullValues:    true,
		ExpectedJsonSchema: nil,
		FilesToGenerate:    []string{"otd_example.proto"},
		ProtoFileName:      "otd_example.proto",
	}

	// Set allowNullValues accordingly:
	sampleProto := sampleProtos["OTD"]

	sampleProtoFileName := fmt.Sprintf("%v/%v", sampleProtoDirectory, ProtoFile)
	protocCommand := exec.Command(
		protocBinary,
		"--descriptor_set_out=/dev/stdout",
		"--include_imports",
		fmt.Sprintf("--proto_path=%v", sampleProtoDirectory),
		fmt.Sprintf("--proto_path=%v", "vendor/github.com/gogo/protobuf/protobuf"),
		fmt.Sprintf("--proto_path=%v", "vendor/github.com/lyft/protoc-gen-validate"),
		sampleProtoFileName)
	var protocCommandOutput bytes.Buffer
	var protocCommandErr bytes.Buffer
	protocCommand.Stdout = &protocCommandOutput
	protocCommand.Stderr = &protocCommandErr

	// Run the command:
	err := protocCommand.Run()
	assert.NoError(t, err, "Unable to prepare a codeGeneratorRequest using protoc (%v) for sample proto file (%v), err = %s", protocBinary, sampleProtoFileName, protocCommandErr.String())

	// Unmarshal the output from the protoc command (should be a "FileDescriptorSet"):
	fileDescriptorSet := new(descriptor.FileDescriptorSet)
	err = proto.Unmarshal(protocCommandOutput.Bytes(), fileDescriptorSet)
	assert.NoError(t, err, "Unable to unmarshal proto FileDescriptorSet for sample proto file (%v)", sampleProtoFileName)

	// Prepare a request:
	codeGeneratorRequest := plugin.CodeGeneratorRequest{
		FileToGenerate: sampleProto.FilesToGenerate,
		ProtoFile:      fileDescriptorSet.GetFile(),
	}

	// Perform the conversion:
	response, err := convert(&codeGeneratorRequest)
	assert.NoError(t, err, "Unable to convert sample proto file (%v)", sampleProtoFileName)
	//assert.Equal(t, len(sampleProto.ExpectedJsonSchema), len(response.File), "Incorrect number of JSON-Schema files returned for sample proto file (%v)", sampleProtoFileName)
	//if len(sampleProto.ExpectedJsonSchema) != len(response.File) {
	//	t.Fail()
	//} else {
	//}

	var rawSchema *string
	var rawSchemas []string
	for _, responseFile := range response.File {
		rawSchemas = append(rawSchemas, *responseFile.Content)
		if responseFile.GetName() == "ExampleEvent.jsonschema" {
			rawSchema = responseFile.Content
		}
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(*rawSchema))
	assert.NoError(t, err, "Failed to parse json schema")

	log.Printf("schemas = %+v\n", rawSchemas)

	r, err := schema.Validate(gojsonschema.NewStringLoader(MessageGood))
	assert.NoError(t, err, "Error when validating json")
	assert.True(t, r.Valid(), fmt.Sprintf("MessageGood should be valid, %+v", r.Errors()))

	r, err = schema.Validate(gojsonschema.NewStringLoader(MessageNoAge))
	assert.NoError(t, err, "Error when validating json")
	assert.False(t, r.Valid(), "MessageNoAge should be invalid")
}
