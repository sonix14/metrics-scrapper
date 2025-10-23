package metric

import (
	"encoding/json"
)

type Metric struct {
	Labels     any      `json:"metric"`
	Values     []any    `json:"values"`
	Timestamps []uint64 `json:"timestamps"`
}

type PRMetricLabels struct {
	Name string `json:"__name__"` //nolint:tagliatelle
	Repo string `json:"repo"`
}

type ExecTimeMetricLabels struct {
	Name string `json:"__name__"` //nolint:tagliatelle
}

func (m Metric) ToJSON() ([]byte, error) {
	return json.Marshal(m) //nolint:wrapcheck
}
