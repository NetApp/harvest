package rest

import (
	"github.com/google/go-cmp/cmp"
	"strconv"
	"testing"
)

func TestBuild(t *testing.T) {
	testQuery := "storage/volumes"
	testFields := []string{"name", "svm"}
	testHiddenFields := []string{"statistics"}
	testFilter := []string{""}
	testReturnTimeout := 10
	expectedHrefTest1 := "api/storage/volumes?return_records=true&fields=name,statistics,svm&return_timeout=10&ignore_unknown_fields=true"
	hrefTest1 := NewHrefBuilder().
		APIPath(testQuery).
		Fields(testFields).
		HiddenFields(testHiddenFields).
		Filter(testFilter).
		ReturnTimeout(&testReturnTimeout).
		IsIgnoreUnknownFieldsEnabled(true).
		Build()

	if hrefTest1 != expectedHrefTest1 {
		t.Errorf("hrefTest1 should be %s but got %s", expectedHrefTest1, hrefTest1)
	}

	testFields = make([]string, 0)
	for i := range URLMaxLimit / len("Test") {
		testFields = append(testFields, "Test"+strconv.Itoa(i))
	}

	expectedHrefTest2 := "api/storage/volumes?return_records=true&fields=*,statistics&return_timeout=10&ignore_unknown_fields=true"
	hrefTest2 := NewHrefBuilder().
		APIPath(testQuery).
		Fields(testFields).
		HiddenFields(testHiddenFields).
		Filter(testFilter).
		ReturnTimeout(&testReturnTimeout).
		IsIgnoreUnknownFieldsEnabled(true).
		Build()

	if hrefTest2 != expectedHrefTest2 {
		t.Errorf("hrefTest2 should be %s but got %s", expectedHrefTest2, hrefTest2)
	}
}

func TestFields(t *testing.T) {
	tests := []struct {
		name           string
		fields         []string
		hiddenFields   []string
		expectedResult []string
	}{
		{
			name: "Test with fields and no hidden fields",
			fields: []string{
				"uuid",
				"block_storage.primary.disk_type",
				"block_storage.primary.raid_type",
			},
			expectedResult: []string{
				"block_storage.primary.disk_type",
				"block_storage.primary.raid_type",
				"uuid",
			},
		},
		{
			name: "Test with fields and hidden fields",
			fields: []string{
				"uuid",
				"block_storage.primary.disk_type",
				"block_storage.primary.raid_type",
			},
			hiddenFields: []string{
				"hidden_field1",
				"hidden_field2",
			},
			expectedResult: []string{
				"block_storage.primary.disk_type",
				"block_storage.primary.raid_type",
				"hidden_field1",
				"hidden_field2",
				"uuid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hBuilder := NewHrefBuilder()
			hBuilder.Fields(tt.fields).HiddenFields(tt.hiddenFields).Build()
			diff := cmp.Diff(hBuilder.fields, tt.expectedResult)
			if diff != "" {
				t.Errorf("Mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
