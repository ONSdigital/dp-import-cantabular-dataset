package handler

import (
	"errors"

	"github.com/ONSdigital/log.go/v2/log"
)

// logData returns logData for an error if there is any. This is used
// to extract log.Data embedded in an error if it implements the dataLogger
// interface
func logData(err error) log.Data {
	var lderr dataLogger
	if errors.As(err, &lderr) {
		return lderr.LogData()
	}

	return nil
}

// instanceCompleted returns whether or not the instance was successfully completed
// before the error was thrown
func instanceCompleted(err error) bool {
	var icerr instanceCompleteder
	if errors.As(err, &icerr) {
		return icerr.InstanceCompleted()
	}

	return false
}
