package tag

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	goversion "github.com/netapp/harvest/v2/third_party/go-version"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"sort"
	"strings"
	"time"
)

type Tag struct {
	*plugin.AbstractPlugin
	schedule    int
	client      *rest.Client
	volumeCache map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Tag{AbstractPlugin: p}
}

func (t *Tag) Init(remote conf.Remote) error {
	var err error

	if err := t.InitAbc(); err != nil {
		return err
	}

	t.volumeCache = make(map[string]string)

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if t.client, err = rest.New(conf.ZapiPoller(t.ParentParams), timeout, t.Auth); err != nil {
		return err
	}
	t.schedule = t.SetPluginInterval()
	exportOptions := t.ParentParams.GetChildS("export_options")
	if exportOptions != nil {
		instanceKeys := exportOptions.GetChildS("instance_keys")
		if instanceKeys != nil {
			tags := instanceKeys.GetChildS("tags")
			if tags == nil {
				instanceKeys.NewChildS("", "tags")
			}
		}
	}
	return t.client.Init(5, remote)
}

func (t *Tag) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[t.Object]

	if len(data.GetInstances()) == 0 {
		return nil, nil, nil
	}

	clusterVersion := t.client.Remote().Version
	ontapVersion, err := goversion.NewVersion(clusterVersion)
	if err != nil {
		t.SLogger.Error("Failed to parse version", slogx.Err(err), slog.String("version", clusterVersion))
		return nil, nil, nil
	}
	version9131 := "9.13.1"
	version9131P, err := goversion.NewVersion(version9131)
	if err != nil {
		t.SLogger.Error("Failed to parse version", slogx.Err(err), slog.String("version", version9131))
		return nil, nil, nil
	}

	if ontapVersion.LessThan(version9131P) {
		return nil, nil, nil
	}

	if t.schedule >= t.PluginInvocationRate {
		t.schedule = 0
		err = t.populateVolumeCache()
		if err != nil {
			return nil, nil, err
		}
	}
	t.schedule++

	if len(t.volumeCache) == 0 {
		return nil, nil, nil
	}

	for _, instance := range data.GetInstances() {
		svm := instance.GetLabel("svm")
		volume := instance.GetLabel("volume")
		tags, exists := t.volumeCache[svm+":"+volume]
		if exists {
			instance.SetLabel("tags", tags)
		}
	}

	return nil, nil, nil
}

// populateVolumeCache fetches volume tags from ONTAP's private CLI API and builds a cache.
// The cache maps a combination of SVM (vserver) and volume names to comma-separated tags.
// It queries the API endpoint "api/private/cli/volume" to retrieve volumes with non-empty tags.
func (t *Tag) populateVolumeCache() error {
	// Clear the existing volume cache and reinitialize it before repopulating
	t.volumeCache = make(map[string]string)

	query := "api/private/cli/volume"
	href := rest.NewHrefBuilder().
		APIPath(query).
		MaxRecords(collectors.DefaultBatchSize).
		Fields([]string{"tags", "volume", "vserver"}).
		Filter([]string{"tags=!\"\""}).
		Build()

	records, err := collectors.InvokeRestCall(t.client, href)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	for _, record := range records {
		if !record.IsObject() {
			t.SLogger.Warn("volume is not object, skipping", slog.String("type", record.Type.String()))
			continue
		}
		volume := record.Get("volume").ClonedString()
		tagsA := record.Get("tags")
		vserver := record.Get("vserver").ClonedString()

		var tags []string
		tagsA.ForEach(func(_, value gjson.Result) bool {
			tags = append(tags, value.ClonedString())
			return true
		})
		sort.Strings(tags)
		tag := strings.Join(tags, ",")
		t.volumeCache[vserver+":"+volume] = tag
	}
	return nil
}
