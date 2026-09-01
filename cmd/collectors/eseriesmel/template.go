package eseriesmel

import (
	"log/slog"
	"strconv"

	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/template"
)

func (e *EseriesMel) LoadTemplate() (string, error) {
	templateName := e.Params.GetChildS("objects").GetChildContentS(e.Object)
	if templateName == "" {
		return "", errs.New(errs.ErrMissingParam, "template for object "+e.Object)
	}

	jitter := e.Params.GetChildContentS("jitter")
	subTemplate, path, err := e.ImportSubTemplate([]string{""}, templateName, jitter, e.Remote.Version)
	if err != nil {
		return "", err
	}

	e.Params.Union(subTemplate)

	e.Logger.Debug("loaded template",
		slog.String("object", e.Object),
		slog.String("template", templateName),
		slog.String("path", path),
	)

	return path, nil
}

func (e *EseriesMel) ParseTemplate() error {
	e.Prop.Object = e.Params.GetChildContentS("object")
	e.Prop.Query = e.Params.GetChildContentS("query")

	if e.Prop.Object == "" {
		return errs.New(errs.ErrMissingParam, "object")
	}
	if e.Prop.Query == "" {
		return errs.New(errs.ErrMissingParam, "query")
	}

	if batchSize := e.Params.GetChildS("batch_size"); batchSize != nil {
		n, err := strconv.Atoi(batchSize.GetContentS())
		if err != nil || n <= 0 {
			return errs.New(errs.ErrInvalidParam, "batch_size: must be a positive integer, got "+batchSize.GetContentS())
		}
		e.Prop.BatchSize = n
	}

	if filter := e.Params.GetChildS("filter"); filter != nil {
		e.Prop.Filter = append(e.Prop.Filter, filter.GetAllChildContentS()...)
	}

	if err := e.parseEvents(); err != nil {
		return err
	}

	if err := e.parseCounters(); err != nil {
		return err
	}

	if len(e.Prop.InstanceKeys) == 0 {
		return errs.New(errs.ErrMissingParam, "counters: at least one ^^ instance key required")
	}

	return nil
}

func (e *EseriesMel) parseEvents() error {
	events := e.Params.GetChildS("events")
	if events == nil || len(events.GetChildren()) == 0 {
		return errs.New(errs.ErrMissingParam, "events")
	}

	for _, line := range events.GetChildren() {
		eventType := line.GetChildContentS(configFieldEventType)
		name := line.GetChildContentS(fieldName)
		if eventType == "" {
			return errs.New(errs.ErrInvalidParam, "events: missing event_type")
		}
		if name == "" {
			return errs.New(errs.ErrInvalidParam, "events: missing name for event_type "+eventType)
		}
		if existing, ok := e.Prop.EventCatalog[eventType]; ok {
			e.Logger.Warn("duplicate event_type in events catalog, keeping first",
				slog.String("event_type", eventType),
				slog.String("kept", existing),
				slog.String("ignored", name))
			continue
		}
		e.Prop.EventCatalog[eventType] = name
	}

	return nil
}

func (e *EseriesMel) parseCounters() error {
	counters := e.Params.GetChildS("counters")
	if counters == nil {
		return nil
	}

	for _, counter := range counters.GetChildren() {
		content := counter.GetContentS()
		if content == "" {
			continue
		}

		name, display, kind, _ := template.ParseMetric(content)

		switch kind {
		case "key", "label":
			if existing, ok := e.Prop.InstanceLabels[name]; ok {
				e.Logger.Warn("duplicate counter, keeping first",
					slog.String("counter", name),
					slog.String("kept", existing),
					slog.String("ignored", display))
				continue
			}
			e.Prop.InstanceLabels[name] = display
			if kind == "key" {
				e.Prop.InstanceKeys = append(e.Prop.InstanceKeys, name)
			}
		default:
			return errs.New(errs.ErrInvalidParam, "counters: "+content+" must be a label (^) or instance key (^^)")
		}
	}

	return nil
}
