package nativepdf

import (
	"context"
	"io"
)

// UnavailableLoader is the default placeholder for real PDF parsing until a
// native backend is implemented behind the nativepdf Loader interface.
type UnavailableLoader struct{}

// NewUnavailableLoader returns the current native PDF loader placeholder.
func NewUnavailableLoader() *UnavailableLoader {
	return &UnavailableLoader{}
}

// OpenPath reports that the real native backend is not implemented yet.
func (l *UnavailableLoader) OpenPath(_ context.Context, _ string, _ OpenOptions) (DocumentHandle, error) {
	return nil, ErrBackendUnavailable
}

// OpenReader reports that the real native backend is not implemented yet.
func (l *UnavailableLoader) OpenReader(_ context.Context, _ string, _ io.Reader, _ OpenOptions) (DocumentHandle, error) {
	return nil, ErrBackendUnavailable
}
