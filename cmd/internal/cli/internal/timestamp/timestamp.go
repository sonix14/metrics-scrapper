package timestamp

import (
	"errors"
	"fmt"
	"time"
)

const Layout = time.DateTime

var ErrUninitialized = errors.New("timestamp uinitialized")

type Timestamp struct{ value time.Time }

func (t *Timestamp) Type() string {
	return "timestamp"
}

func (t *Timestamp) Set(s string) error {
	if t == nil {
		return ErrUninitialized
	}

	parsed, err := time.Parse(Layout, s)
	if err != nil {
		return fmt.Errorf("parsing timestamp: %w", err)
	}

	t.value = parsed

	return nil
}

func (t *Timestamp) String() string {
	return t.value.Format(Layout)
}
