package conf

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
	HasREST         bool
	IsClustered     bool
}

func (r Remote) IsZero() bool {
	return r.Name == "" && r.Model == "" && r.UUID == ""
}

func (r Remote) IsKeyPerf() bool {
	return r.IsDisaggregated && !r.IsSanOptimized
}
