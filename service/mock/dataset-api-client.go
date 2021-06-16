// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"sync"
)

// DatasetAPIClientMock is a mock implementation of service.DatasetAPIClient.
//
// 	func TestSomethingThatUsesDatasetAPIClient(t *testing.T) {
//
// 		// make and configure a mocked service.DatasetAPIClient
// 		mockedDatasetAPIClient := &DatasetAPIClientMock{
// 			CheckerFunc: func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
// 				panic("mock out the Checker method")
// 			},
// 			PutInstanceFunc: func(contextMoqParam context.Context, s1 string, s2 string, s3 string, s4 string, updateInstance dataset.UpdateInstance) error {
// 				panic("mock out the PutInstance method")
// 			},
// 		}
//
// 		// use mockedDatasetAPIClient in code that requires service.DatasetAPIClient
// 		// and then make assertions.
//
// 	}
type DatasetAPIClientMock struct {
	// CheckerFunc mocks the Checker method.
	CheckerFunc func(contextMoqParam context.Context, checkState *healthcheck.CheckState) error

	// PutInstanceFunc mocks the PutInstance method.
	PutInstanceFunc func(contextMoqParam context.Context, s1 string, s2 string, s3 string, s4 string, updateInstance dataset.UpdateInstance) error

	// calls tracks calls to the methods.
	calls struct {
		// Checker holds details about calls to the Checker method.
		Checker []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// CheckState is the checkState argument value.
			CheckState *healthcheck.CheckState
		}
		// PutInstance holds details about calls to the PutInstance method.
		PutInstance []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
			// S3 is the s3 argument value.
			S3 string
			// S4 is the s4 argument value.
			S4 string
			// UpdateInstance is the updateInstance argument value.
			UpdateInstance dataset.UpdateInstance
		}
	}
	lockChecker     sync.RWMutex
	lockPutInstance sync.RWMutex
}

// Checker calls CheckerFunc.
func (mock *DatasetAPIClientMock) Checker(contextMoqParam context.Context, checkState *healthcheck.CheckState) error {
	if mock.CheckerFunc == nil {
		panic("DatasetAPIClientMock.CheckerFunc: method is nil but DatasetAPIClient.Checker was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}{
		ContextMoqParam: contextMoqParam,
		CheckState:      checkState,
	}
	mock.lockChecker.Lock()
	mock.calls.Checker = append(mock.calls.Checker, callInfo)
	mock.lockChecker.Unlock()
	return mock.CheckerFunc(contextMoqParam, checkState)
}

// CheckerCalls gets all the calls that were made to Checker.
// Check the length with:
//     len(mockedDatasetAPIClient.CheckerCalls())
func (mock *DatasetAPIClientMock) CheckerCalls() []struct {
	ContextMoqParam context.Context
	CheckState      *healthcheck.CheckState
} {
	var calls []struct {
		ContextMoqParam context.Context
		CheckState      *healthcheck.CheckState
	}
	mock.lockChecker.RLock()
	calls = mock.calls.Checker
	mock.lockChecker.RUnlock()
	return calls
}

// PutInstance calls PutInstanceFunc.
func (mock *DatasetAPIClientMock) PutInstance(contextMoqParam context.Context, s1 string, s2 string, s3 string, s4 string, updateInstance dataset.UpdateInstance) error {
	if mock.PutInstanceFunc == nil {
		panic("DatasetAPIClientMock.PutInstanceFunc: method is nil but DatasetAPIClient.PutInstance was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		S1              string
		S2              string
		S3              string
		S4              string
		UpdateInstance  dataset.UpdateInstance
	}{
		ContextMoqParam: contextMoqParam,
		S1:              s1,
		S2:              s2,
		S3:              s3,
		S4:              s4,
		UpdateInstance:  updateInstance,
	}
	mock.lockPutInstance.Lock()
	mock.calls.PutInstance = append(mock.calls.PutInstance, callInfo)
	mock.lockPutInstance.Unlock()
	return mock.PutInstanceFunc(contextMoqParam, s1, s2, s3, s4, updateInstance)
}

// PutInstanceCalls gets all the calls that were made to PutInstance.
// Check the length with:
//     len(mockedDatasetAPIClient.PutInstanceCalls())
func (mock *DatasetAPIClientMock) PutInstanceCalls() []struct {
	ContextMoqParam context.Context
	S1              string
	S2              string
	S3              string
	S4              string
	UpdateInstance  dataset.UpdateInstance
} {
	var calls []struct {
		ContextMoqParam context.Context
		S1              string
		S2              string
		S3              string
		S4              string
		UpdateInstance  dataset.UpdateInstance
	}
	mock.lockPutInstance.RLock()
	calls = mock.calls.PutInstance
	mock.lockPutInstance.RUnlock()
	return calls
}
