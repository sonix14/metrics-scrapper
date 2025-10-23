package vmdb

import (
	"bytes"
	"fmt"
	"metrics-scrapper/internal/vmdb/internal/metric"
)

type Metrics struct {
	Data []metric.Metric
}

func (m *Metrics) AddPRMetric(
	name string,
	repo string,
	value any,
	timestamp uint64,
) {
	m.Data = append(
		m.Data,
		metric.Metric{
			Labels: metric.PRMetricLabels{
				Name: name,
				Repo: repo,
			},
			Values:     []any{value},
			Timestamps: []uint64{timestamp},
		},
	)
}

const execTimeMetricName = "scraper_exec_timestamp"

func (m *Metrics) AddExecTimeMetric(
	value any,
	timestamp uint64,
) {
	m.Data = append(
		m.Data,
		metric.Metric{
			Labels: metric.ExecTimeMetricLabels{
				Name: execTimeMetricName,
			},
			Values:     []any{value},
			Timestamps: []uint64{timestamp},
		},
	)
}

func (m *Metrics) ExportToJSON() ([]byte, error) {
	var exported bytes.Buffer

	for _, metric := range m.Data {
		formatted, err := metric.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("marshalling metric: %w", err)
		}

		_, err = exported.Write(formatted)
		if err != nil {
			return nil, fmt.Errorf("appending metric bytes: %w", err)
		}

		err = exported.WriteByte('\n')
		if err != nil {
			return nil, fmt.Errorf("appending delimiter: %w", err)
		}
	}

	return exported.Bytes(), nil
}
