# Architecture

This document describes the high-level architecture of Harvest.
If you want to familiarize yourself with the code base, you're in the right place!

## Bird's Eye View

## Code Map

This section describes the important directories and data structures.

### Matrix


### Collectors 

Collectors are responsible for collecting metrics from data sources
* ZAPI collectors for: 
  * Performance
  * Shelf
  * Capacity
  * Process and Hardware monitors

### Plug-Ins

Plug-Ins post-process metrics from collectors, aggregate new counters, or trigger actions. 

They allow you to extend Harvest without writing a new collector, instead you can augment an existing one.

### Exporters 

Exporters encode metrics and send then to a time-series database. Out of the box
Harvest includes Prometheus and InfluxDB exporters.



## Request Flows

## Plugins

## Life-cycle
