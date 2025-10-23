package vmdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type vmDBResponse struct {
	Status string       `json:"status"`
	Data   responseData `json:"data"`
	Stats  interface{}  `json:"stats"`
}

type responseData struct {
	ResultType string   `json:"resultType"`
	Result     []result `json:"result"`
}

type result struct {
	Metric interface{}   `json:"metric"`
	Value  []interface{} `json:"value"`
}

func (m *vmdbExporter) GetLastExecTimestamp() (time.Time, error) {
	ctx := context.TODO()

	urlGetLastExec, err := m.getLastExecTimestampURL()
	if err != nil {
		return time.Time{}, fmt.Errorf("creating url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlGetLastExec, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %w", ErrSendingRequest, err)
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("%d, %w", resp.StatusCode, ErrUnexpectedResponseStatusCode)
	}

	var response vmDBResponse

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %w", ErrReadingRequestBody, err)
	}

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %w", ErrUnmarshalingRequestBody, err)
	}

	return extractLastExecTimestamp(response.Data.Result)
}

func (m *vmdbExporter) PushExecTimestamp(t time.Time) error {
	metrics := Metrics{} //nolint:exhaustruct

	metrics.AddExecTimeMetric(
		uint64(t.UnixMilli()),
		uint64(t.UnixMilli()),
	)

	return nil
	//return m.PushMetrics(&metrics)
}

func (m *vmdbExporter) getLastExecTimestampURL() (string, error) {
	var (
		urlGetLastUpdate *url.URL
		err              error
	)

	urlGetLastUpdate, err = url.Parse(m.URLProvider.Get())
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrParsingVMURL, err)
	}

	query := urlGetLastUpdate.Query()
	query.Set(
		"query",
		strings.Join(
			[]string{
				"last_over_time(", execTimeMetricName, "[", m.LastExecTimestampSearchRange, "])",
			}, ""))

	urlGetLastUpdate.RawQuery = query.Encode()

	return urlGetLastUpdate.String(), nil
}

func extractLastExecTimestamp(r []result) (time.Time, error) {
	if len(r) == 0 {
		return time.Time{}, nil
	}

	lastExecStr, ok := r[0].Value[1].(string)
	if !ok {
		return time.Time{}, ErrFailedConvertExecTimestamp
	}

	lastExecInt, err := strconv.ParseInt(lastExecStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("converting last exec timestamp to uint64 : %w", err)
	}

	return time.UnixMilli(lastExecInt), nil
}
