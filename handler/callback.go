package handler

import (
	"github.com/ONSdigital/log.go/v2/log"
	"errors"
)

// logData returns logData for an error if there is any
func logData(err error) log.Data{
	var lderr dataLogger
	if errors.As(err, &lderr){
		return lderr.LogData()
	}

	return nil
}
