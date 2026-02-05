package controller

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
)

type Controller struct {
	*plugin.AbstractPlugin
	client           *rest.Client
	schedule         int
	controllerLabels map[string]string // Maps controller_id -> label ("A", "B", etc.)
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Controller{AbstractPlugin: p}
}

func (c *Controller) Init(remote conf.Remote) error {
	if err := c.InitAbc(); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(c.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, c.SLogger)
	if c.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	// Skip actual API call in test mode
	if !c.Options.IsTest {
		if err := c.client.Init(1, remote); err != nil {
			return err
		}
	}

	c.controllerLabels = make(map[string]string)

	c.schedule = c.SetPluginInterval()
	return nil
}

func (c *Controller) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]

	arrayID := c.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		c.SLogger.Warn("arrayID not found in ParentParams, skipping controller labels")
		return nil, nil, nil
	}

	if c.schedule >= c.PluginInvocationRate {
		c.schedule = 0
		c.refreshControllerLabels(arrayID)
	}
	c.schedule++

	c.applyControllerLabels(data)

	return nil, nil, nil
}

func (c *Controller) refreshControllerLabels(arrayID string) {
	c.controllerLabels = make(map[string]string)

	controllerLabels, err := c.buildControllerLabelMap(arrayID)
	if err != nil {
		c.SLogger.Warn("Failed to build controller label map", slogx.Err(err))
		return
	}

	c.controllerLabels = controllerLabels
	c.SLogger.Debug("Refreshed controller labels", slog.Int("count", len(c.controllerLabels)))
}

func (c *Controller) buildControllerLabelMap(arrayID string) (map[string]string, error) {
	controllerLabels := make(map[string]string)

	apiPath := c.client.APIPath + "/storage-systems/" + arrayID + "/controllers"
	controllers, err := c.client.Fetch(apiPath, nil)
	if err != nil {
		return controllerLabels, fmt.Errorf("failed to fetch controllers: %w", err)
	}

	for _, controller := range controllers {
		controllerID := controller.Get("controllerRef").String()
		if controllerID == "" {
			controllerID = controller.Get("id").String()
		}

		label := controller.Get("physicalLocation.label").String()

		if controllerID != "" && label != "" {
			controllerLabels[controllerID] = label
		}
	}

	c.SLogger.Debug("Built controller label map", slog.Int("count", len(controllerLabels)))
	return controllerLabels, nil
}

func (c *Controller) applyControllerLabels(data *matrix.Matrix) {
	for _, instance := range data.GetInstances() {
		controllerID := instance.GetLabel("controller_id")
		if controllerID == "" {
			controllerID = instance.GetLabel("id")
		}
		if controllerID == "" {
			continue
		}

		if label, ok := c.controllerLabels[controllerID]; ok {
			instance.SetLabel("controller", label)
		} else {
			c.SLogger.Debug("Controller label not found in cache",
				slog.String("controller_id", controllerID))
		}
	}
}
