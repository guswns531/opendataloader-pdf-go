package nativepdf

import "errors"

// ErrBackendUnavailable marks the current native PDF backend as intentionally
// unavailable so callers can fall back to the temporary bridge.
var ErrBackendUnavailable = errors.New("native pdf backend unavailable")
