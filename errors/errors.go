// Copyright 2019 Cloudbase Solutions Srl
// All Rights Reserved.

package errors

import "fmt"

var (
	// ErrNoInfo is returned when no info could be found about a resource
	ErrNoInfo = fmt.Errorf("no info available")
)

// ErrInvalidDevice is returned when a device does not meet the
// required criteria to be considered valid.
type ErrInvalidDevice struct {
	message string
}

func (e ErrInvalidDevice) Error() string {
	return e.message
}

// NewInvalidDeviceErr returns a new ErrInvalidDevice
func NewInvalidDeviceErr(msg string) error {
	return &ErrInvalidDevice{
		message: msg,
	}
}

// IsInvalidDevice checks if the supplied error is of type ErrInvalidDevice
func IsInvalidDevice(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(*ErrInvalidDevice)
	return ok
}

// ErrVolumeNotFound is returned when a particular volume was not found
type ErrVolumeNotFound struct {
	message string
}

func (e ErrVolumeNotFound) Error() string {
	return e.message
}

// NewVolumeNotFoundErr returns a new ErrVolumeNotFound
func NewVolumeNotFoundErr(msg string) error {
	return &ErrVolumeNotFound{
		message: msg,
	}
}

// IsVolumeNotFound checks if the supplied error is of type ErrVolumeNotFound
func IsVolumeNotFound(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(*ErrVolumeNotFound)
	return ok
}

// ErrOperationInterrupted is returned when an operation is interrupted
type ErrOperationInterrupted struct {
	message string
}

func (e ErrOperationInterrupted) Error() string {
	return e.message
}

// NewOperationInterruptedErr returns a new ErrOperationInterrupted error
func NewOperationInterruptedErr(msg string) error {
	return &ErrOperationInterrupted{
		message: msg,
	}
}

// IsOperationInterrupted checks if the supplied error is of type ErrOperationInterrupted
func IsOperationInterrupted(err error) bool {
	if err == nil {
		return false
	}

	_, ok := err.(*ErrOperationInterrupted)
	return ok
}
