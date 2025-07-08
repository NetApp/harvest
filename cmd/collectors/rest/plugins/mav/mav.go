package mav

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type AlertSeverity string

const (
	fieldIndex               = "index"
	fieldQuery               = "query"
	fieldOperation           = "operation"
	fieldState               = "state"
	fieldApprovedUsers       = "approved_users"
	fieldUserRequested       = "user_requested"
	fieldCreateTime          = "create_time"
	fieldApproveExpiryTime   = "approve_expiry_time"
	fieldExecutionExpiryTime = "execution_expiry_time"
	fieldApproveTime         = "approve_time"
	fieldSeqID               = "seq_id"
	fieldUserVetoed          = "user_vetoed"
	fieldLabels              = "labels"
)

type Mav struct {
	*plugin.AbstractPlugin
	client                *rest.Client
	mavData               *matrix.Matrix
	mavDataExtendedMatrix *matrix.Matrix
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Mav{AbstractPlugin: p}
}

func (m *Mav) Init(remote conf.Remote) error {

	var err error

	if err := m.InitAbc(); err != nil {
		return err
	}

	m.initMatrix()

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if m.client, err = rest.New(conf.ZapiPoller(m.ParentParams), timeout, m.Auth); err != nil {
		return err
	}

	return m.client.Init(5, remote)
}

func newExportOptions(fields ...string) *node.Node {
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	for _, f := range fields {
		instanceKeys.NewChildS("", f)
	}
	return exportOptions
}

func (m *Mav) initMatrix() {
	// Build the first matrix
	m.mavData = matrix.New(m.Parent+".MavRequest", "mav_request", "MavRequest")
	fieldsForMavData := []string{
		fieldIndex,
		fieldQuery,
		fieldOperation,
		fieldUserRequested,
		fieldSeqID,
	}
	exportOptions := newExportOptions(fieldsForMavData...)
	m.mavData.SetExportOptions(exportOptions)

	// Build the second matrix (with extra fields)
	m.mavDataExtendedMatrix = matrix.New(m.Parent+".MavRequest2", "mav_request", "MavRequest2")
	fieldsForMavDataLabel := []string{
		fieldIndex,
		fieldQuery,
		fieldOperation,
		fieldUserRequested,
		fieldSeqID,
		// add below fields which can change to this matrix
		fieldState,
		fieldApprovedUsers,
		fieldUserVetoed,
	}
	exportOptions2 := newExportOptions(fieldsForMavDataLabel...)
	m.mavDataExtendedMatrix.SetExportOptions(exportOptions2)
}

func (m *Mav) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[m.Object]
	m.client.Metadata.Reset()
	// Purge and reset data
	m.mavData.PurgeInstances()
	m.mavData.Reset()
	m.mavDataExtendedMatrix.PurgeInstances()
	m.mavDataExtendedMatrix.Reset()

	m.mavData.SetGlobalLabels(data.GetGlobalLabels())
	m.mavDataExtendedMatrix.SetGlobalLabels(data.GetGlobalLabels())

	err := m.collectMAVRequests()
	if err != nil {
		return nil, m.client.Metadata, err
	}
	return []*matrix.Matrix{m.mavData, m.mavDataExtendedMatrix}, m.client.Metadata, nil
}

func (m *Mav) collectMAVRequests() error {
	var (
		instance         *matrix.Instance
		instanceExtended *matrix.Instance
	)
	clusterTime, err := collectors.GetClusterTime(m.client, nil, m.SLogger)
	if err != nil {
		return err
	}
	// Subtract 5 minutes to capture the expired status and ensure the full lifecycle of the request is included.
	// All states excluding executed reaches expired state
	approveTimeFilter := fmt.Sprintf("%s=>=%d", fieldApproveExpiryTime, clusterTime.Add(-5*time.Minute).Unix())
	filter := []string{approveTimeFilter}

	approveRecords, err := m.getMAVRequests(filter)
	if err != nil {
		return err
	}

	expiryTimeFilter := fmt.Sprintf("%s=>=%d", fieldExecutionExpiryTime, clusterTime.Add(-5*time.Minute).Unix())
	filter = []string{expiryTimeFilter}

	records, err := m.getMAVRequests(filter)
	if err != nil {
		return err
	}

	records = append(records, approveRecords...)

	mat := m.mavData
	matExtended := m.mavDataExtendedMatrix
	for _, record := range records {
		index := record.Get(fieldIndex).ClonedString()
		if mat.GetInstance(index) != nil && matExtended.GetInstance(index) != nil {
			continue
		}
		query := record.Get(fieldQuery).ClonedString()
		operation := record.Get(fieldOperation).ClonedString()
		state := record.Get(fieldState).ClonedString()
		userRequested := record.Get(fieldUserRequested).ClonedString()
		userVetoed := record.Get(fieldUserVetoed).ClonedString()
		approvedUsersA := record.Get(fieldApprovedUsers)

		var appUserNames []string
		approvedUsersA.ForEach(func(_, value gjson.Result) bool {
			appUserNames = append(appUserNames, value.String())
			return true
		})
		approvedUsers := strings.Join(appUserNames, ", ")
		instance, err = mat.NewInstance(index)
		if err != nil {
			m.SLogger.Warn("error while creating instance", slog.String("key", index))
			continue
		}
		instance.SetLabel(fieldQuery, query)
		instance.SetLabel(fieldOperation, operation)
		instance.SetLabel(fieldState, state)
		instance.SetLabel(fieldUserRequested, userRequested)
		instance.SetLabel(fieldApprovedUsers, approvedUsers)
		instance.SetLabel(fieldIndex, index)

		var createTime, approveExpiryTime, executeExpiryTime, approveTime float64
		if createTimeStr := record.Get(fieldCreateTime).ClonedString(); createTimeStr != "" {
			createTime = collectors.HandleTimestamp(createTimeStr) * 1000
			instance.SetLabel(fieldSeqID, strconv.Itoa(int(createTime)))
			m.setMetric(m.mavData, instance, fieldCreateTime, createTime)
		}
		if approveExpiryTimeStr := record.Get(fieldApproveExpiryTime).ClonedString(); approveExpiryTimeStr != "" {
			approveExpiryTime = collectors.HandleTimestamp(approveExpiryTimeStr) * 1000
			m.setMetric(m.mavData, instance, fieldApproveExpiryTime, approveExpiryTime)
		}
		if executeExpiryTimeStr := record.Get(fieldExecutionExpiryTime).ClonedString(); executeExpiryTimeStr != "" {
			executeExpiryTime = collectors.HandleTimestamp(executeExpiryTimeStr) * 1000
			m.setMetric(m.mavData, instance, fieldExecutionExpiryTime, executeExpiryTime)
		}
		if approveTimeStr := record.Get(fieldApproveTime).ClonedString(); approveTimeStr != "" {
			approveTime = collectors.HandleTimestamp(approveTimeStr) * 1000
			m.setMetric(m.mavData, instance, fieldApproveTime, approveTime)
		}

		// set other matrix
		instanceExtended, err = matExtended.NewInstance(index)
		if err != nil {
			m.SLogger.Warn("error while creating instance", slog.String("key", index))
			continue
		}
		instanceExtended.SetLabel(fieldQuery, query)
		instanceExtended.SetLabel(fieldOperation, operation)
		instanceExtended.SetLabel(fieldState, state)
		instanceExtended.SetLabel(fieldUserRequested, userRequested)
		instanceExtended.SetLabel(fieldApprovedUsers, approvedUsers)
		instanceExtended.SetLabel(fieldUserVetoed, userVetoed)
		instanceExtended.SetLabel(fieldIndex, index)
		instanceExtended.SetLabel(fieldSeqID, strconv.Itoa(int(createTime)))
		m.setMetric(m.mavDataExtendedMatrix, instance, fieldLabels, float64(clusterTime.Unix()))
	}
	return nil
}

func (m *Mav) setMetric(mat *matrix.Matrix, instance *matrix.Instance, name string, value float64) {
	var err error
	met := mat.GetMetric(name)
	if met == nil {
		if met, err = mat.NewMetricFloat64(name); err != nil {
			m.SLogger.Warn(
				"error while creating metric",
				slogx.Err(err),
				slog.String("key", name),
			)
			return
		}
	}
	met.SetValueFloat64(instance, value)
}

func (m *Mav) getMAVRequests(filter []string) ([]gjson.Result, error) {
	fields := []string{
		fieldIndex,
		fieldQuery,
		fieldOperation,
		fieldState,
		fieldApprovedUsers,
		fieldUserVetoed,
		fieldUserRequested,
		fieldCreateTime,
		fieldApproveExpiryTime,
		fieldExecutionExpiryTime,
		fieldApproveTime,
	}
	query := "api/security/multi-admin-verify/requests"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Filter(filter).
		Build()

	fmt.Println(href)

	return collectors.InvokeRestCall(m.client, href)
}
