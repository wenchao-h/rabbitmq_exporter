package main

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	RegisterExporter("memory", newExporterMemory)
}

var (
	memoryLabels = []string{"cluster", "node"}

	memoryGaugeVec = map[string]*prometheus.GaugeVec{
		"memory.allocated_unused": newGaugeVec(
			"memory_allocated_unused_bytes",
			"memory preallocated by the runtime (VM allocators) but not yet used",
			memoryLabels,
		),
		"memory.atom": newGaugeVec(
			"memory_atom_bytes",
			"memory used by atoms. Should be fairly constant",
			memoryLabels,
		),
		"memory.binary": newGaugeVec(
			"memory_binary_bytes",
			"memory used by shared binary data in the runtime. Most of this memory is message bodies and metadata.)",
			memoryLabels,
		),
		"memory.code": newGaugeVec(
			"memory_code_bytes",
			"memory used by code (bytecode, module metadata). This section is usually fairly constant and relatively small (unless the node is entirely blank and stores no data).",
			memoryLabels,
		),
		"memory.connection_channels": newGaugeVec(
			"memory_connection_channels_bytes",
			"memory used by client connections - channels",
			memoryLabels,
		),
		"memory.connection_other": newGaugeVec(
			"memory_connection_other_bytes",
			"memory used by client connection - other",
			memoryLabels,
		),
		"memory.connection_readers_bytes": newGaugeVec(
			"memory_connection_readers",
			"memory used by processes responsible for connection parser and most of connection state. Most of their memory attributes to TCP buffers",
			memoryLabels,
		),
		"memory.connection_writers": newGaugeVec(
			"memory_connection_writers_bytes",
			"memory used by processes responsible for serialization of outgoing protocol frames and writing to client connections socktes",
			memoryLabels,
		),
		"memory.metrics": newGaugeVec(
			"memory_metrics_bytes",
			"node-local metrics. The more connections, channels, queues are node hosts, the more stats there are to collect and keep",
			memoryLabels,
		),
		"memory.mgmt_db": newGaugeVec(
			"memory_mgmt_db_bytes",
			"management DB ETS tables + processes",
			memoryLabels,
		),
		"memory.mnesia": newGaugeVec(
			"memory_mnesia_bytes",
			"internal database (Mnesia) tables keep an in-memory copy of all its data (even on disc nodes)",
			memoryLabels,
		),
		"memory.msg_index": newGaugeVec(
			"memory_msg_index_bytes",
			"message index ETS + processes",
			memoryLabels,
		),
		"memory.other_ets": newGaugeVec(
			"memory_other_ets_bytes",
			"other in-memory tables besides those belonging to the stats database and internal database tables",
			memoryLabels,
		),
		"memory.other_proc": newGaugeVec(
			"memory_other_proc_bytes",
			"memory used by all other processes that RabbitMQ cannot categorise",
			memoryLabels,
		),
		"memory.other_system": newGaugeVec(
			"memory_other_system_bytes",
			"memory used by all other system that RabbitMQ cannot categorise",
			memoryLabels,
		),
		"memory.plugins": newGaugeVec(
			"memory_plugins_bytes",
			"memory used by plugins (apart from the Erlang client which is counted under Connections, and the management database which is counted separately). ",
			memoryLabels,
		),
		"memory.queue_procs": newGaugeVec(
			"memory_queue_procs_bytes",
			"memory used by class queue masters, queue indices, queue state",
			memoryLabels,
		),
		"memory.queue_slave_procs": newGaugeVec(
			"memory_queue_slave_procs_bytes",
			"memory used by class queue mirrors, queue indices, queue state",
			memoryLabels,
		),
		"memory.reserved_unallocated": newGaugeVec(
			"memory_reserved_unallocated_bytes",
			"memory preallocated/reserved by the kernel but not the runtime",
			memoryLabels,
		),
		"memory.total": newGaugeVec(
			"memory_total_bytes",
			"total memory",
			memoryLabels,
		),
	}
)

type exporterMemory struct {
	memoryMetricsGauge map[string]*prometheus.GaugeVec
}

func newExporterMemory() Exporter {
	memoryGaugeVecActual := memoryGaugeVec

	if len(config.ExcludeMetrics) > 0 {
		for _, metric := range config.ExcludeMetrics {
			if memoryGaugeVecActual[metric] != nil {
				delete(memoryGaugeVecActual, metric)
			}
		}
	}

	return exporterMemory{
		memoryMetricsGauge: memoryGaugeVecActual,
	}
}

func (e exporterMemory) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	for _, gauge := range e.memoryMetricsGauge {
		gauge.Reset()
	}
	selfNode := ""
	if n, ok := ctx.Value(nodeName).(string); ok {
		selfNode = n
	}
	cluster := ""
	if n, ok := ctx.Value(clusterName).(string); ok {
		cluster = n
	}
	rabbitMemoryResponses, err := getMetricMap(config, fmt.Sprintf("nodes/%s/memory", selfNode))
	if err != nil {
		return err
	}
	for key, gauge := range e.memoryMetricsGauge {
		if value, ok := rabbitMemoryResponses[key]; ok {
			gauge.WithLabelValues(cluster, selfNode).Set(value)
		}
	}
	for _, gauge := range e.memoryMetricsGauge {
		gauge.Collect(ch)
	}
	return nil
}

func (e exporterMemory) Describe(ch chan<- *prometheus.Desc) {
	for _, gauge := range e.memoryMetricsGauge {
		gauge.Describe(ch)
	}
}
