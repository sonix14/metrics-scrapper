package vmdb

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/exp/slog"
)

const vmImportPath = "/api/v1/import"

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type vmdbExporter struct {
	Client                       Client
	URLProvider                  URLProvider
	Logger                       *slog.Logger
	LastExecTimestampSearchRange string
}

func NewVMDBExporter(
	client Client,
	urlProvider URLProvider,
	logger *slog.Logger,
	lastExecTimestampSearchRange string,
) *vmdbExporter {
	return &vmdbExporter{
		Client:                       client,
		URLProvider:                  urlProvider,
		Logger:                       logger,
		LastExecTimestampSearchRange: lastExecTimestampSearchRange,
	}
}

func (m *vmdbExporter) PushMetrics(metrics *Metrics) error {
	var (
		vmImportEndpoint *url.URL
		err              error
	)

	exported, err := metrics.ExportToJSON()
	if err != nil {
		return err
	}

	ctx := context.TODO()

	vmImportEndpoint, err = url.Parse(m.URLProvider.Post())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrParsingVMURL, err)
	}

	vmImportEndpoint.Path = vmImportPath

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, vmImportEndpoint.String(), bytes.NewReader(exported))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := m.Client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendingRequest, err)
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("%d, %w", resp.StatusCode, ErrUnexpectedResponseStatusCode)
	}

	return nil
}
