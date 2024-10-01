package rest

import (
	"strconv"
	"testing"
)

func TestBuild(t *testing.T) {

	testQuery := "storage/volumes"
	testFields := []string{"name", "svm"}
	testFilter := []string{""}
	testReturnTimeout := 10
	isIgnoreUnknownFieldsEnabled := true
	expectedHrefTest1 := "api/storage/volumes?return_records=true&fields=name,svm&return_timeout=10&ignore_unknown_fields=true"
	hrefTest1 := NewHrefBuilder().
		APIPath(testQuery).
		Fields(testFields).
		Filter(testFilter).
		ReturnTimeout(&testReturnTimeout).
		IsIgnoreUnknownFieldsEnabled(isIgnoreUnknownFieldsEnabled).
		Build()

	if hrefTest1 != expectedHrefTest1 {
		t.Errorf("hrefTest1 should be %s but got %s", expectedHrefTest1, hrefTest1)
	}

	testFields = make([]string, 0)
	for i := range 100000000 {
		testFields = append(testFields, "Test"+strconv.Itoa(i))
	}

	expectedHrefTest2 := "api/storage/volumes?return_records=true&fields=*&return_timeout=10&ignore_unknown_fields=true"
	hrefTest2 := NewHrefBuilder().
		APIPath(testQuery).
		Fields(testFields).
		Filter(testFilter).
		ReturnTimeout(&testReturnTimeout).
		IsIgnoreUnknownFieldsEnabled(isIgnoreUnknownFieldsEnabled).
		Build()

	if hrefTest2 != expectedHrefTest2 {
		t.Errorf("hrefTest2 should be %s but got %s", expectedHrefTest2, hrefTest2)
	}
}
