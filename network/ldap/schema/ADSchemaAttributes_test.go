package schema

import (
	"testing"
)

func Test_SchemaAttributeDisplayNameToGUID_In_GUIDToSchemaAttributeDisplayName(t *testing.T) {
	for schemaAttributeDisplayName, schemaAttributeGUID := range SchemaAttributeDisplayNameToGUID {
		if _, exists := GUIDToSchemaAttributeDisplayName[schemaAttributeGUID]; !exists {
			t.Errorf("Key %s from SchemaAttributeDisplayNameToGUID not found in GUIDToSchemaAttributeDisplayName", schemaAttributeDisplayName)
		}
	}
}

func Test_GUIDToSchemaAttributeDisplayName_In_SchemaAttributeDisplayNameToGUID(t *testing.T) {
	for schemaAttributeGUID, schemaAttributeDisplayName := range GUIDToSchemaAttributeDisplayName {
		if _, exists := SchemaAttributeDisplayNameToGUID[schemaAttributeDisplayName]; !exists {
			t.Errorf("Key %s from GUIDToSchemaAttributeDisplayName not found in SchemaAttributeDisplayNameToGUID", schemaAttributeGUID)
		}
	}
}
