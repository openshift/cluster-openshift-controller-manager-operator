package workload

import (
	"errors"
	"fmt"
)

// manageError append possible errors using the informed name for identification. When modified, it
// appends a ErrForceRollout as well.
func manageError(errSlice []error, name string, err error, modified bool) []error {
	if err != nil {
		errSlice = append(errSlice, fmt.Errorf("%q: %v", name, err))
	}
	if modified {
		errSlice = append(errSlice, fmt.Errorf("%w: %q", ErrForceRollout, name))
	}
	return errSlice
}

// hasError assert the presence of ErrForceRollout in the slice of errors informed.
func hasError(errSlice []error, wanted error) bool {
	for _, err := range errSlice {
		if errors.Is(err, wanted) {
			return true
		}
	}
	return false
}
