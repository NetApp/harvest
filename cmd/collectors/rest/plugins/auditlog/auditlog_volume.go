package auditlog

import (
	"encoding/json"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var (
	replacer = strings.NewReplacer("^I", " ", `\"`, `"`)
)

type VolumeHandler interface {
	ExtractNames(string, *AuditLog) (string, string, string, error)
	GetOperation() string
	GetRefreshCache() bool
}

var (
	volumeNameRe             = regexp.MustCompile(`(?i)-volume\s+(\S+)`)
	svmNameRe                = regexp.MustCompile(`(?i)-vserver\s+(\S+)`)
	newNameRe                = regexp.MustCompile(`(?i)-newname\s+(\S+)`)
	patchVolumeRe            = regexp.MustCompile(`PATCH\s*/api/storage/volumes/([a-f0-9-]+)\s*:\s*.*?\{.*}`)
	postVolumeRe             = regexp.MustCompile(`POST\s*/api/storage/volumes\s*:\s*(\{.*})`)
	deleteVolumeRe           = regexp.MustCompile(`DELETE\s*/api/storage/volumes/([a-f0-9-]+)`)
	postPrivateCliVolumeRe   = regexp.MustCompile(`POST\s*/api/private/cli/volume/?\s*:\s*(\{.*})`)
	postPrivateCliRenameRe   = regexp.MustCompile(`POST\s*/api/private/cli/volume/rename/?\s*:\s*(\{.*})`)
	deletePrivateCliVolumeRe = regexp.MustCompile(`DELETE\s*/api/private/cli/volume/?\s*:\s*(\{.*})`)
	postApplicationRe        = regexp.MustCompile(`POST\s*/api/application/applications\s*:\s*(?:\[\s*.*\s*])?\s*(\{.*})`)
)

// VolumeWriteHandler handles volume write operations
type VolumeWriteHandler struct {
	op           string
	refreshCache bool
}

func (v VolumeWriteHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	volumeName := ""
	svmName := ""

	volumeMatches := volumeNameRe.FindStringSubmatch(input)
	if len(volumeMatches) > 1 {
		volumeName = volumeMatches[1]
	}

	svmMatches := svmNameRe.FindStringSubmatch(input)
	if len(svmMatches) > 1 {
		svmName = svmMatches[1]
	}

	return volumeName, svmName, "", nil
}

func (v VolumeWriteHandler) GetOperation() string {
	return v.op
}

func (v VolumeWriteHandler) GetRefreshCache() bool {
	return v.refreshCache
}

type VolumeRenameHandler struct {
	op           string
	refreshCache bool
}

func (v VolumeRenameHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}

	svmName := ""
	newName := ""

	svmMatches := svmNameRe.FindStringSubmatch(input)
	if len(svmMatches) > 1 {
		svmName = svmMatches[1]
	}

	newNameMatches := newNameRe.FindStringSubmatch(input)
	if len(newNameMatches) > 1 {
		newName = newNameMatches[1]
	}

	return newName, svmName, "", nil
}

func (v VolumeRenameHandler) GetOperation() string {
	return v.op
}

func (v VolumeRenameHandler) GetRefreshCache() bool {
	return v.refreshCache
}

// VolumePatchHandler handles PATCH /api/storage/volumes
type VolumePatchHandler struct {
	op           string
	refreshCache bool
}

func (v VolumePatchHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := patchVolumeRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	uuid := matches[1]
	volumeInfo, exists := a.GetVolumeInfo(uuid)
	if !exists {
		return "", "", uuid, nil
	}

	return volumeInfo.name, volumeInfo.svm, uuid, nil
}

func (v VolumePatchHandler) GetOperation() string {
	return v.op
}

func (v VolumePatchHandler) GetRefreshCache() bool {
	return v.refreshCache
}

// VolumePostHandler handles POST /api/storage/volumes
type VolumePostHandler struct {
	op           string
	refreshCache bool
}

func (v VolumePostHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := postVolumeRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	var payload struct {
		SVM  string `json:"svm"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &payload); err != nil {
		return "", "", "", err
	}

	return payload.Name, payload.SVM, "", nil
}

func (v VolumePostHandler) GetOperation() string {
	return v.op
}

func (v VolumePostHandler) GetRefreshCache() bool {
	return v.refreshCache
}

// VolumeDeleteHandler handles DELETE /api/storage/volumes
type VolumeDeleteHandler struct {
	op           string
	refreshCache bool
}

func (v VolumeDeleteHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := deleteVolumeRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	uuid := matches[1]
	volumeInfo, exists := a.GetVolumeInfo(uuid)
	if !exists {
		return "", "", uuid, nil
	}

	return volumeInfo.name, volumeInfo.svm, uuid, nil
}

func (v VolumeDeleteHandler) GetOperation() string {
	return v.op
}

func (v VolumeDeleteHandler) GetRefreshCache() bool {
	return v.refreshCache
}

// VolumePrivateCliPostHandler handles POST /api/private/cli/volume
type VolumePrivateCliPostHandler struct {
	op           string
	refreshCache bool
}

func (v VolumePrivateCliPostHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := postPrivateCliVolumeRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	var payload struct {
		Vserver string `json:"vserver"`
		Volume  string `json:"volume"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &payload); err != nil {
		return "", "", "", err
	}

	return payload.Volume, payload.Vserver, "", nil
}

func (v VolumePrivateCliPostHandler) GetOperation() string {
	return v.op
}

func (v VolumePrivateCliPostHandler) GetRefreshCache() bool {
	return v.refreshCache
}

// VolumePrivateCliRenameHandler handles POST /api/private/cli/volume/rename
type VolumePrivateCliRenameHandler struct {
	op           string
	refreshCache bool
}

func (v VolumePrivateCliRenameHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := postPrivateCliRenameRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	var payload struct {
		Vserver string `json:"vserver"`
		Volume  string `json:"volume"`
		NewName string `json:"newname"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &payload); err != nil {
		return "", "", "", err
	}

	return payload.NewName, payload.Vserver, "", nil
}

func (v VolumePrivateCliRenameHandler) GetOperation() string {
	return v.op
}

func (v VolumePrivateCliRenameHandler) GetRefreshCache() bool {
	return v.refreshCache
}

// VolumePrivateCliDeleteCliHandler handles DELETE /api/private/cli/volume
type VolumePrivateCliDeleteCliHandler struct {
	op           string
	refreshCache bool
}

func (v VolumePrivateCliDeleteCliHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(v.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := deletePrivateCliVolumeRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	var payload struct {
		Vserver string `json:"vserver"`
		Volume  string `json:"volume"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &payload); err != nil {
		return "", "", "", err
	}

	return payload.Volume, payload.Vserver, "", nil
}

func (v VolumePrivateCliDeleteCliHandler) GetOperation() string {
	return v.op
}

func (v VolumePrivateCliDeleteCliHandler) GetRefreshCache() bool {
	return v.refreshCache
}

type ApplicationPostHandler struct {
	op           string
	refreshCache bool
}

func (ap ApplicationPostHandler) ExtractNames(input string, a *AuditLog) (string, string, string, error) {
	err := a.RefreshVolumeCache(ap.GetRefreshCache())
	if err != nil {
		return "", "", "", err
	}
	matches := postApplicationRe.FindStringSubmatch(input)
	if len(matches) < 2 {
		return "", "", "", nil
	}

	var payload struct {
		Name string `json:"name"`
		SVM  struct {
			Name string `json:"name"`
		} `json:"svm"`
		Template struct {
			Name string `json:"name"`
		} `json:"template"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &payload); err != nil {
		return "", "", "", err
	}

	if payload.Template.Name != "nas" {
		return "", "", "", nil
	}

	return payload.Name, payload.SVM.Name, "", nil
}

func (ap ApplicationPostHandler) GetOperation() string {
	return ap.op
}

func (ap ApplicationPostHandler) GetRefreshCache() bool {
	return ap.refreshCache
}

var volumeInputHandlers = map[*regexp.Regexp]VolumeHandler{
	regexp.MustCompile(`^volume create\s+`): VolumeWriteHandler{op: "create", refreshCache: true},
	regexp.MustCompile(`^volume modify\s+`): VolumeWriteHandler{op: "update"},
	regexp.MustCompile(`^volume delete\s+`): VolumeWriteHandler{op: "delete"},
	regexp.MustCompile(`^volume rename\s+`): VolumeRenameHandler{op: "rename"},
	postVolumeRe:                            VolumePostHandler{op: "create", refreshCache: true},
	patchVolumeRe:                           VolumePatchHandler{op: "update"},
	deleteVolumeRe:                          VolumeDeleteHandler{op: "delete"},
	postPrivateCliVolumeRe:                  VolumePrivateCliPostHandler{op: "create", refreshCache: true},
	postPrivateCliRenameRe:                  VolumePrivateCliRenameHandler{op: "update"},
	deletePrivateCliVolumeRe:                VolumePrivateCliDeleteCliHandler{op: "delete"},
	postApplicationRe:                       ApplicationPostHandler{op: "create", refreshCache: true},
}

func (a *AuditLog) parseVolumeRecords(response []gjson.Result) {
	mat := a.data
	object := "volume"
	for _, result := range response {
		timestamp := result.Get("timestamp")
		var auditTimeStamp int64
		if timestamp.Exists() {
			t, err := time.Parse(time.RFC3339, timestamp.ClonedString())
			if err != nil {
				a.SLogger.Error(
					"Failed to load cluster date",
					slogx.Err(err),
					slog.String("date", timestamp.ClonedString()),
				)
				continue
			}
			auditTimeStamp = t.Unix()
		}
		application := result.Get("application").ClonedString()
		location := result.Get("location").ClonedString()
		user := result.Get("user").ClonedString()
		input := normalizeInput(result.Get("input").ClonedString())
		for re, handler := range volumeInputHandlers {
			if !re.MatchString(input) {
				continue
			}
			volume, svm, uuid, err := handler.ExtractNames(input, a)
			if err != nil {
				a.SLogger.Warn("error while parsing audit log", slogx.Err(err), slog.String("key", input))
				continue
			}
			if volume == "" || svm == "" {
				if uuid == "" {
					a.SLogger.Warn("failed to extract data", slog.String("input", input))
					continue
				}
			}
			instanceKey := application + location + user + svm + volume + uuid + handler.GetOperation() + object
			if instance := mat.GetInstance(instanceKey); instance != nil {
				a.setLogMetric(mat, instance, float64(auditTimeStamp))
				if err != nil {
					a.SLogger.Warn(
						"Unable to set value on metric",
						slogx.Err(err),
						slog.String("metric", "log"),
					)
					continue
				}
			} else {
				instance, err = mat.NewInstance(instanceKey)
				if err != nil {
					a.SLogger.Warn("error while creating instance", slog.String("key", instanceKey))
					continue
				}
				if volume == "" && svm == "" {
					instance.SetLabel("uuid", uuid)
				}
				instance.SetLabel("application", application)
				instance.SetLabel("location", location)
				instance.SetLabel("user", user)
				instance.SetLabel("op", handler.GetOperation())
				instance.SetLabel("object", object)
				instance.SetLabel("volume", volume)
				instance.SetLabel("svm", svm)
				a.setLogMetric(mat, instance, float64(auditTimeStamp))
				if err != nil {
					a.SLogger.Warn("error while setting metric value", slogx.Err(err))
					return
				}
			}
		}
	}
}

func normalizeInput(input string) string {
	return replacer.Replace(input)
}
