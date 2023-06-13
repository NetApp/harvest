package volume

import (
	"errors"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/api/ontapi/zapi"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"strconv"
)

type Volume struct {
	*plugin.AbstractPlugin
	currentVal int
	client     *zapi.Client
	aggrsMap   map[string]string // aggregate-uuid -> aggregate-name map
}

type aggrData struct {
	aggrUUID string
	aggrName string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Volume{AbstractPlugin: p}
}

func (my *Volume) Init() error {

	var err error

	if err = my.InitAbc(); err != nil {
		return err
	}

	if my.client, err = zapi.New(conf.ZapiPoller(my.ParentParams), my.Auth); err != nil {
		my.Logger.Error().Stack().Err(err).Msg("connecting")
		return err
	}

	if err = my.client.Init(5); err != nil {
		return err
	}

	my.aggrsMap = make(map[string]string)

	// Assigned the value to currentVal so that plugin would be invoked first time to populate cache.
	my.currentVal = my.SetPluginInterval()

	return nil
}

func (my *Volume) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, error) {

	data := dataMap[my.Object]
	if my.currentVal >= my.PluginInvocationRate {
		my.currentVal = 0

		// invoke disk-encrypt-get-iter zapi and populate disk info
		disks, err1 := my.getEncryptedDisks()
		// invoke aggr-status-get-iter zapi and populate aggr disk mapping info
		aggrDiskMap, err2 := my.getAggrDiskMapping()

		if err1 != nil {
			if errors.Is(err1, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err1).Msg("Failed to collect disk data")
			} else {
				my.Logger.Error().Err(err1).Msg("Failed to collect disk data")
			}
		}
		if err2 != nil {
			if errors.Is(err2, errs.ErrNoInstance) {
				my.Logger.Debug().Err(err2).Msg("Failed to collect aggregate-disk mapping data")
			} else {
				my.Logger.Error().Err(err2).Msg("Failed to collect aggregate-disk mapping data")
			}
		}
		// update aggrsMap based on disk data and addr disk mapping
		my.updateAggrMap(disks, aggrDiskMap)
	}

	// update volume instance labels
	my.updateVolumeLabels(data)

	my.currentVal++
	return nil, nil
}

func (my *Volume) updateVolumeLabels(data *matrix.Matrix) {
	for _, volume := range data.GetInstances() {
		if !volume.IsExportable() {
			continue
		}
		aggrUUID := volume.GetLabel("aggrUuid")
		_, exist := my.aggrsMap[aggrUUID]
		volume.SetLabel("isHardwareEncrypted", strconv.FormatBool(exist))
	}
}

func (my *Volume) getEncryptedDisks() ([]string, error) {
	var (
		result    []*node.Node
		diskNames []string
		err       error
	)

	request := node.NewXMLS("disk-encrypt-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	//algorithm is -- Protection mode needs to be DATA or FULL
	// Fetching rest of them and add as
	query := request.NewChildS("query", "")
	encryptInfoQuery := query.NewChildS("disk-encrypt-info", "")
	encryptInfoQuery.NewChildS("protection-mode", "open|part|miss")

	// fetching only disks whose protection-mode is open/part/miss
	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, disk := range result {
		diskName := disk.GetChildContentS("disk-name")
		diskNames = append(diskNames, diskName)
	}
	return diskNames, nil
}

func (my *Volume) updateAggrMap(disks []string, aggrDiskMap map[string]aggrData) {
	if disks != nil && aggrDiskMap != nil {
		// Clean aggrsMap map
		my.aggrsMap = make(map[string]string)

		for _, disk := range disks {
			aggr := aggrDiskMap[disk]
			my.aggrsMap[aggr.aggrUUID] = aggr.aggrName
		}
	}
}

func (my *Volume) getAggrDiskMapping() (map[string]aggrData, error) {
	var (
		result        []*node.Node
		aggrsDisksMap map[string]aggrData
		diskName      string
		err           error
	)

	request := node.NewXMLS("aggr-status-get-iter")
	request.NewChildS("max-records", collectors.DefaultBatchSize)
	aggrsDisksMap = make(map[string]aggrData)

	if result, err = my.client.InvokeZapiCall(request); err != nil {
		return nil, err
	}

	if len(result) == 0 || result == nil {
		return nil, errs.New(errs.ErrNoInstance, "no records found")
	}

	for _, aggrDiskData := range result {
		aggrUUID := aggrDiskData.GetChildContentS("aggregate-uuid")
		aggrName := aggrDiskData.GetChildContentS("aggregate")
		aggrDiskList := aggrDiskData.GetChildS("aggr-plex-list").GetChildS("aggr-plex-info").GetChildS("aggr-raidgroup-list").GetChildS("aggr-raidgroup-info").GetChildS("aggr-disk-list").GetChildren()
		for _, aggrDisk := range aggrDiskList {
			diskName = aggrDisk.GetChildContentS("disk")
			aggrsDisksMap[diskName] = aggrData{aggrUUID: aggrUUID, aggrName: aggrName}
		}
	}
	return aggrsDisksMap, nil
}
