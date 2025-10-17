package conf

import "github.com/netapp/harvest/v2/third_party/tidwall/gjson"

type Remote struct {
	Name            string
	Model           string
	UUID            string
	Version         string
	Release         string
	Serial          string
	IsSanOptimized  bool
	IsDisaggregated bool
	ZAPIsExist      bool
	ZAPIsChecked    bool
	HasREST         bool
	IsClustered     bool
}

const (
	ASAr2 = "asar2"
	CDOT  = "cdot"
)

func (r Remote) IsZero() bool {
	return r.Name == "" && r.Model == "" && r.UUID == ""
}

func (r Remote) IsKeyPerf() bool {
	return r.IsDisaggregated
}

func (r Remote) IsAFX() bool {
	return r.IsDisaggregated && !r.IsSanOptimized
}

func (r Remote) IsASAr2() bool {
	return r.Model == ASAr2
}

func NewRemote(results gjson.Result) Remote {
	var remote Remote
	remote.Name = results.Get("name").ClonedString()
	remote.UUID = results.Get("uuid").ClonedString()
	remote.Version = results.Get("version.generation").ClonedString() + "." +
		results.Get("version.major").ClonedString() + "." +
		results.Get("version.minor").ClonedString()
	remote.Release = results.Get("version.full").ClonedString()
	remote.IsSanOptimized = results.Get("san_optimized").Bool()
	remote.IsDisaggregated = results.Get("disaggregated").Bool()
	remote.IsClustered = true
	remote.HasREST = true
	remote.Model = CDOT
	if remote.IsDisaggregated && remote.IsSanOptimized {
		remote.Model = ASAr2
	}

	return remote
}
