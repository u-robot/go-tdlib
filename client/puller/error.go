package puller

import (
	"errors"
)

// ErrEndOfPull is error trown on End Of Pull (EOP).
var ErrEndOfPull = errors.New("end of pull")
