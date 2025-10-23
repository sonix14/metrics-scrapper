package manager

import (
	"metrics-scrapper/internal/vmdb"
	"time"
)

type VMDBExporter interface {
	PushMetrics(collection *vmdb.Metrics) error
	PushExecTimestamp(t time.Time) error
	GetLastExecTimestamp() (time.Time, error)
}
