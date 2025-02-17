package auditlog

import (
	"testing"
)

// MockCacheRefresher is a mock implementation of the CacheRefresher interface
type MockCacheRefresher struct{}

func (m *MockCacheRefresher) RefreshCache() error {
	return nil
}

func init() {
	volumeCache["123e4567-e89b-12d3-a456-426614174000"] = VolumeInfo{Name: "testVolume", SVM: "testSVM"}
}

func TestVolumeWriteCreateHandler(t *testing.T) {
	handler := VolumeWriteHandler{op: "create"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		"volume create -vserver testSVM -volume testVolume -state online -policy default -size 200MB -aggregate umeng_aff300_aggr2",
		"volume create -volume testVolume -size 200MB -vserver testSVM -state online -policy default -aggregate umeng_aff300_aggr2",
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "create" {
			t.Errorf("Expected operation: create, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumeWriteModifyHandler(t *testing.T) {
	handler := VolumeWriteHandler{op: "update"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		"volume modify -vserver testSVM -volume testVolume -size 201MB -state online",
		"volume modify -vserver testSVM -volume       testVolume -size 201MB -state online",
		"volume modify -vserver testSVM -size 201MB -state online -volume testVolume",
		"volume modify -vserver testSVM -volume testVolume -state offline",
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "update" {
			t.Errorf("Expected operation: update, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumeWriteRenameHandler(t *testing.T) {
	handler := VolumeRenameHandler{op: "update"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		"volume rename -vserver testSVM -volume testVolume -newname testVolume1",
		"volume rename -vserver testSVM        -volume      testVolume      -newname       testVolume1",
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume1" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume1, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "update" {
			t.Errorf("Expected operation: update, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumeWriteDeleteHandler(t *testing.T) {
	handler := VolumeWriteHandler{op: "delete"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		"volume create -volume testVolume -vserver testSVM",
		"volume create -volume    testVolume    -vserver    testSVM",
		"volume create -volume\ttestVolume\t-vserver\ttestSVM",
		"volume create -volume\ntestVolume\n-vserver\ntestSVM",
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "delete" {
			t.Errorf("Expected operation: delete, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumePatchHandler(t *testing.T) {
	handler := VolumePatchHandler{op: "update"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		"PATCH /api/storage/volumes/123e4567-e89b-12d3-a456-426614174000 : {\"name\": \"testVolume\", \"size\": \"220MB\"}",
		"PATCH /api/storage/volumes/123e4567-e89b-12d3-a456-426614174000 : {\"name\": \"testVolume\", \"size\": \"220MB\"     }",
		"PATCH /api/storage/volumes/123e4567-e89b-12d3-a456-426614174000 : [\"X-Dot-Client-App: SMv4\"] {\"state\":\"online\"}",
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "update" {
			t.Errorf("Expected operation: update, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumePostHandler(t *testing.T) {
	handler := VolumePostHandler{op: "create"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		`POST /api/storage/volumes : {"svm":"testSVM","name":"testVolume"}`,
		`POST /api/storage/volumes : {"svm": "testSVM", "name": "testVolume"}`,
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "create" {
			t.Errorf("Expected operation: create, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumeDeleteHandler(t *testing.T) {
	handler := VolumeDeleteHandler{op: "delete"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		"DELETE /api/storage/volumes/123e4567-e89b-12d3-a456-426614174000",
		"DELETE    /api/storage/volumes/123e4567-e89b-12d3-a456-426614174000",
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "delete" {
			t.Errorf("Expected operation: delete, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumePrivateCliPostHandler(t *testing.T) {
	handler := VolumePrivateCliPostHandler{op: "create"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		`POST /api/private/cli/volume : {"vserver":"testSVM","volume":"testVolume"}`,
		`POST /api/private/cli/volume : {"vserver": "testSVM", "volume": "testVolume"}`,
		`POST /api/private/cli/volume : { ^I^I\"vserver\": \"testSVM\", ^I^I\"volume\": \"testVolume\", ^I^I\"size\": \"200MB\", ^I^I\"aggregate\": \"umeng_aff300_aggr1\" ^I}`,
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "create" {
			t.Errorf("Expected operation: create, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumePrivateCliRenameHandler(t *testing.T) {
	handler := VolumePrivateCliRenameHandler{op: "update"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		`POST /api/private/cli/volume/rename : {"vserver":"testSVM","volume":"testVolume","newname":"newTestVolume"}`,
		`POST /api/private/cli/volume/rename : {"vserver": "testSVM", "volume": "testVolume", "newname": "newTestVolume"}`,
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "newTestVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: newTestVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "update" {
			t.Errorf("Expected operation: update, got: %s", handler.GetOperation())
		}
	}
}

func TestVolumePrivateCliDeleteCliHandler(t *testing.T) {
	handler := VolumePrivateCliDeleteCliHandler{op: "delete"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		`DELETE /api/private/cli/volume : {"vserver":"testSVM","volume":"testVolume"}`,
		`DELETE /api/private/cli/volume : {"vserver": "testSVM", "volume": "testVolume"}`,
		`DELETE /api/private/cli/volume : { ^I^I\"vserver\": \"testSVM\", ^I^I\"volume\": \"testVolume\" ^I}`,
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testVolume" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testVolume, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "delete" {
			t.Errorf("Expected operation: delete, got: %s", handler.GetOperation())
		}
	}
}

func TestApplicationPostHandler(t *testing.T) {
	handler := ApplicationPostHandler{op: "create"}
	refresher := &MockCacheRefresher{}
	inputs := []string{
		`POST /api/application/applications : ["X-Dot-Client-App: SMv4"] {"name":"testApp","svm":{"name":"testSVM"},"template":{"name":"nas"}}`,
		`POST /api/application/applications : ["X-Dot-Client-App: SMv4"] {"name": "testApp", "svm": {"name": "testSVM"}, "template": {"name": "nas"}}`,
		`POST /api/application/applications : {\"name\":\"testApp\",\"smart_container\":true,\"svm\":{\"name\":\"testSVM\"},\"nas\":{\"nfs_access\":[],\"cifs_access\":[],\"application_components\":[{\"name\":\"t1\",\"total_size\":1073741824,\"share_count\":1,\"storage_service\":{\"name\":\"performance\"},\"export_policy\":{\"id\":30064771074}}],\"protection_type\":{\"remote_rpo\":\"none\"}},\"template\":{\"name\":\"nas\"}}`,
	}

	for _, input := range inputs {
		input = normalizeInput(input)
		volume, svm, _, _ := handler.ExtractNames(input, refresher)
		if volume != "testApp" || svm != "testSVM" {
			t.Errorf("Input: %s, Expected volume: testApp, svm: testSVM, got volume: %s, svm: %s", input, volume, svm)
		}
		if handler.GetOperation() != "create" {
			t.Errorf("Expected operation: create, got: %s", handler.GetOperation())
		}
	}
}
