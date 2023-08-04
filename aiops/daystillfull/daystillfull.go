package daystillfull

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"math"
	"time"
)

const (
	// lamInv is 1/lam. lam is the forgetting factor, typically close to 1. we use 0.98
	lamInv                = 1.02040816327
	NanosInDay            = float32(24 * 3600 * 1e9)
	defaultPluginInterval = 24 * time.Hour
)

type DaysTillFull struct {
	*plugin.AbstractPlugin
	dtfRules  []dtfRule
	dtfByUUID map[string]*dtf
	calls     int
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &DaysTillFull{AbstractPlugin: p}
}

func (d *DaysTillFull) Init() error {

	var (
		err error
	)

	if err = d.AbstractPlugin.Init(); err != nil {
		return err
	}

	d.parseRules()

	if len(d.dtfRules) == 0 {
		return errs.New(errs.ErrMissingParam, "valid rules")
	} else {
		d.Logger.Debug().Int("count", len(d.dtfRules)).Msg("parsed rules")
	}

	d.dtfByUUID = make(map[string]*dtf)
	d.SetPluginInterval(defaultPluginInterval)

	return nil
}

type dtfRule struct {
	used   string
	total  string
	metric string
}

// parse rules from plugin parameters
func (d *DaysTillFull) parseRules() {
	var (
		usedName  string
		totalName string
		metric    string
	)

	s := d.Params.GetChildS("counters")
	if s == nil {
		return
	}
	for _, c := range s.GetChildren() {
		name := c.GetNameS()
		if name == "used" {
			usedName = c.GetContentS()
		} else if name == "total" {
			totalName = c.GetContentS()
		} else if name == "metric" {
			metric = c.GetContentS()
		}

		if usedName == "" || totalName == "" || metric == "" {
			continue
		}

		rule := dtfRule{
			used:   usedName,
			total:  totalName,
			metric: metric,
		}

		d.dtfRules = append(d.dtfRules, rule)
		usedName = ""
		totalName = ""
		metric = ""
	}
}

func (d *DaysTillFull) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {
	var err error
	// Using the plugin's schedule, calculate the PluginInvocationRate and add observations every PluginInvocationRate
	if d.calls%d.PluginInvocationRate == 0 {
		err = d.addObservations(dataMap)
	}
	d.calls++
	return nil, err
}

func (d *DaysTillFull) addObservations(dataMap map[string]*matrix.Matrix) error {
	var (
		metric *matrix.Metric
		err    error
	)

	data := dataMap[d.Object]
	now := time.Now()
	shouldExport := d.calls > 0

	d.Logger.Debug().Bool("shouldExport", shouldExport).Int("calls", d.calls).Send()

	for _, rule := range d.dtfRules {
		usedMetric := data.DisplayMetric(rule.used)
		totalMetric := data.DisplayMetric(rule.total)
		dtfMetric := rule.metric

		if usedMetric == nil || totalMetric == nil {
			d.Logger.Warn().Str("used", rule.used).Str("total", rule.total).Msg("metric does not exist")
			continue
		}

		for uuid, volume := range data.GetInstances() {
			if !volume.IsExportable() {
				continue
			}
			used, ok := usedMetric.GetValueFloat64(volume)
			if !ok {
				continue
			}
			total, ok := totalMetric.GetValueFloat64(volume)
			if !ok {
				continue
			}

			ttf, ok := d.dtfByUUID[uuid]
			if !ok {
				ttf = newDtf(2, now)
				d.dtfByUUID[uuid] = ttf
			}
			currentDayIndex := getDayIndex(now, ttf)
			ttf.addObs(currentDayIndex, float32(used))

			if !shouldExport {
				continue
			}

			// Predict the day index when storage is filled to max capacity
			predictDayIndex := maxOf(0.0, ttf.predict(float32(total)))
			daysToFull := int(predictDayIndex - currentDayIndex)
			daysToFull = constrain(daysToFull, 0.0, 365.0)

			// create new metric or update existing one
			if metric = data.GetMetric(dtfMetric); metric == nil {
				if metric, err = data.NewMetricFloat64(dtfMetric); err != nil {
					return fmt.Errorf("failed to create dtf metric=%s err=%w", dtfMetric, err)
				}
			}

			err = metric.AddValueFloat64(volume, float64(daysToFull))
			if err != nil {
				d.Logger.Error().Err(err).Str("metric", dtfMetric).Msg("Unable to add dtf")
			}
		}
	}

	return nil
}

func constrain(x int, lower int, upper int) int {
	if x < lower {
		return lower
	}
	if x > upper {
		return upper
	}
	return x
}

func maxOf(a float32, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func getDayIndex(ts time.Time, dtf *dtf) float32 {
	diff := float32(ts.UnixNano()-dtf.initTime.UnixNano()) / NanosInDay
	return float32(math.Round(float64(diff*100)) / 100)
}

type dtf struct {
	initTime time.Time
	a        *Mat // 2x2 matrix
	w        *Mat // 2x1 matrix
}

func (t *dtf) addObs(dayIndex float32, val float32) {
	x := NewMatrix(1, 2)
	x.Set(0, 0, 1.0)
	x.Set(0, 1, dayIndex)
	t.addValue(x.T(), val)
}

func (t *dtf) addValue(x *Mat, value float32) {
	// z = self.lam_inv * self.A * x
	scale := t.a.Scale(lamInv)
	z := scale.DotProduct(x)

	// alpha = float((1 + x.T*z)**(-1))
	xT := x.T()
	multiple := xT.DotProduct(z)
	multiple = multiple.Add(1.0)
	multiple = multiple.Reciprocal()
	alpha := multiple.flt()

	// Decompose: self.w = self.w + (t-alpha * float(x.T*(self.w+t*z))) * z
	// tz:                                                       t*z
	// plus:                                              self.w+t*z
	// mm:                                           x.T * plus
	// alphaTerm:                      alpha * float(mm)
	// tMinus:                       t-alphaTerm
	// prod:                         tMinus * z
	tz := z.Scale(value)
	plus := t.w.Plus(tz)
	mm := xT.DotProduct(plus)
	alphaTerm := alpha * mm.flt()
	tMinus := value - alphaTerm
	prod := z.Scale(tMinus)
	t.w = t.w.Plus(prod)

	// self.A = self.A - alpha*z*z.T
	right := z.Scale(alpha).DotProduct(z.T())
	t.a = t.a.Minus(right)
}

// Predict x given t based on t=Wx or t = mx + b
func (t *dtf) predict(total float32) float32 {
	//intercept, slope = self.w[0], self.w[1]
	intercept := (*t.w)[0][0]
	slope := (*t.w)[1][0]

	predDayIndex := (total - intercept) / slope
	predDayIndex64 := float64(predDayIndex)
	if math.IsInf(predDayIndex64, 0) || math.IsNaN(predDayIndex64) {
		return 365
	}
	return predDayIndex
}

func newDtf(numVars int, when time.Time) *dtf {
	return &dtf{
		initTime: when,
		a:        Identity(numVars),
		w:        NewMatrix(numVars, 1),
	}
}
