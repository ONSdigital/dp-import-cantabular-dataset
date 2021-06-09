package event

import (
	"errors"
	"net/http"
	"github.com/ONSdigital/log.go/v2/log"
)

// statusCode is a callback function that allows you to extract
// a status code from an error, or returns 500 as a default
func statusCode(err error) int{
	var cerr coder
	if errors.As(err, &cerr){
		return cerr.Code()
	}

	return http.StatusInternalServerError
}

// logData returns logData for an error if there is any
func logData(err error) log.Data{
	var lderr dataLogger
	if errors.As(err, &lderr){
		return lderr.LogData()
	}

	return nil
}

// unwrapLogData recursively unwraps logData from an error
func unwrapLogData(err error) log.Data{
	var data []log.Data

	for err != nil && errors.Unwrap(err) != nil{
		if lderr, ok := err.(dataLogger); ok{
			if d := lderr.LogData(); d != nil{
				data = append(data, d)
			}
		}

		err = errors.Unwrap(err)
	}

	// flatten []log.Data into single log.Data with slice
	// entries for duplicate keyed values
	logData := log.Data{}
	for _, d := range data{
		for k, v := range d{
			if val, ok := logData[k]; ok{
				if s, ok := val.([]interface{}); ok {
					s = append(s, v)
					logData[k] = s
				} else {
					logData[k] = []interface{}{val, v}
				}
			} else{
				logData[k] = v
			}
		}
	}

	return logData
}
