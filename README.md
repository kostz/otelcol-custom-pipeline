# Otelcol custom pipeline

Repository contains the example of Otel collector pipeline with custom event receiver and enricher processor adding metadata to logs records.

## Overview

### Demoapp
Produces logs events and counter metrics

### Ip Resolve Service
Implements the resolving metadata by IP with caching mechanism

### Event receiver
Custom Otel Collector receiver receiving event messages, converting them to Otel and propagaing to pipeline

### Event Enrich Processor
Custom Otel Collector processor adding metadata to log records by ip address calling IP Resolve Service

## Run
```shell
make -C ipresolveservice run & make run & make -C demoapp run
```

