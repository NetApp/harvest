package firmware

import (
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

const firmwareVersionMatrix = "eseries_firmware_version"

type Firmware struct {
	*plugin.AbstractPlugin
	client *rest.Client
	data   *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Firmware{AbstractPlugin: p}
}

func (f *Firmware) Init(remote conf.Remote) error {
	if err := f.InitAbc(); err != nil {
		return err
	}

	clientTimeout := f.ParentParams.GetChildContentS("client_timeout")
	if clientTimeout == "" {
		clientTimeout = rest.DefaultTimeout
	}

	duration, err := time.ParseDuration(clientTimeout)
	if err != nil {
		f.SLogger.Info("Using default timeout", slog.String("timeout", rest.DefaultTimeout))
		duration, _ = time.ParseDuration(rest.DefaultTimeout)
	}

	poller, err := conf.PollerNamed(f.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, f.SLogger)

	if f.client, err = rest.New(poller, duration, credentials, ""); err != nil {
		return err
	}

	if err := f.client.Init(1, remote); err != nil {
		return err
	}

	f.data = matrix.New(f.Parent+".Firmware", firmwareVersionMatrix, firmwareVersionMatrix)
	f.data.SetExportOptions(matrix.NewExportOptionsWithLabels(
		[]string{"array_id", "code_module"},
		[]string{"array_id", "code_module", "version"},
	))

	return nil
}

func (f *Firmware) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	arrayID := f.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		f.SLogger.Warn("array_id not found in ParentParams, skipping firmware collection")
		return nil, nil, nil
	}

	query := rest.NewURLBuilder().APIPath("firmware/embedded-firmware/{array_id}/versions").ArrayID(arrayID).Build()
	results, err := f.client.Fetch(f.client.APIPath+"/"+query, nil)
	if err != nil {
		f.SLogger.Error("Failed to fetch firmware versions", slogx.Err(err))
		return nil, nil, err
	}
	if len(results) == 0 {
		f.SLogger.Warn("No firmware version data returned")
		return nil, nil, nil
	}

	f.data.PurgeInstances()
	f.data.Reset()
	f.data.SetGlobalLabels(dataMap[f.Object].GetGlobalLabels())

	f.processCodeVersions(f.data, arrayID, results[0])

	metadata := &collector.Metadata{}
	metadata.PluginInstances.Store(uint64(len(f.data.GetInstances())))

	return []*matrix.Matrix{f.data}, metadata, nil
}

func (f *Firmware) processCodeVersions(mat *matrix.Matrix, arrayID string, response gjson.Result) {
	versions := response.Get("codeVersions")
	if !versions.Exists() || !versions.IsArray() {
		f.SLogger.Debug("No codeVersions found in firmware response")
		return
	}

	for _, ver := range versions.Array() {
		codeModule := ver.Get("codeModule").ClonedString()
		versionString := ver.Get("versionString").ClonedString()

		if codeModule == "" {
			continue
		}

		key := arrayID + "_" + codeModule
		inst, err := mat.NewInstance(key)
		if err != nil {
			f.SLogger.Error("Failed to create firmware version instance", slogx.Err(err), slog.String("key", key))
			continue
		}

		inst.SetLabelTrimmed("array_id", arrayID)
		inst.SetLabelTrimmed("code_module", codeModule)
		inst.SetLabelTrimmed("version", versionString)
	}
}
