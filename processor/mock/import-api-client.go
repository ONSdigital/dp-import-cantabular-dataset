// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-import-cantabular-dataset/processor"
	"sync"
)

// Ensure, that ImportAPIClientMock does implement processor.ImportAPIClient.
// If this is not the case, regenerate this file with moq.
var _ processor.ImportAPIClient = &ImportAPIClientMock{}

// ImportAPIClientMock is a mock implementation of processor.ImportAPIClient.
//
// 	func TestSomethingThatUsesImportAPIClient(t *testing.T) {
//
// 		// make and configure a mocked processor.ImportAPIClient
// 		mockedImportAPIClient := &ImportAPIClientMock{
// 			UpdateImportJobStateFunc: func(contextMoqParam context.Context, s1 string, s2 string, s3 string) error {
// 				panic("mock out the UpdateImportJobState method")
// 			},
// 		}
//
// 		// use mockedImportAPIClient in code that requires processor.ImportAPIClient
// 		// and then make assertions.
//
// 	}
type ImportAPIClientMock struct {
	// UpdateImportJobStateFunc mocks the UpdateImportJobState method.
	UpdateImportJobStateFunc func(contextMoqParam context.Context, s1 string, s2 string, s3 string) error

	// calls tracks calls to the methods.
	calls struct {
		// UpdateImportJobState holds details about calls to the UpdateImportJobState method.
		UpdateImportJobState []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
			// S3 is the s3 argument value.
			S3 string
		}
	}
	lockUpdateImportJobState sync.RWMutex
}

// UpdateImportJobState calls UpdateImportJobStateFunc.
func (mock *ImportAPIClientMock) UpdateImportJobState(contextMoqParam context.Context, s1 string, s2 string, s3 string) error {
	if mock.UpdateImportJobStateFunc == nil {
		panic("ImportAPIClientMock.UpdateImportJobStateFunc: method is nil but ImportAPIClient.UpdateImportJobState was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		S1              string
		S2              string
		S3              string
	}{
		ContextMoqParam: contextMoqParam,
		S1:              s1,
		S2:              s2,
		S3:              s3,
	}
	mock.lockUpdateImportJobState.Lock()
	mock.calls.UpdateImportJobState = append(mock.calls.UpdateImportJobState, callInfo)
	mock.lockUpdateImportJobState.Unlock()
	return mock.UpdateImportJobStateFunc(contextMoqParam, s1, s2, s3)
}

// UpdateImportJobStateCalls gets all the calls that were made to UpdateImportJobState.
// Check the length with:
//     len(mockedImportAPIClient.UpdateImportJobStateCalls())
func (mock *ImportAPIClientMock) UpdateImportJobStateCalls() []struct {
	ContextMoqParam context.Context
	S1              string
	S2              string
	S3              string
} {
	var calls []struct {
		ContextMoqParam context.Context
		S1              string
		S2              string
		S3              string
	}
	mock.lockUpdateImportJobState.RLock()
	calls = mock.calls.UpdateImportJobState
	mock.lockUpdateImportJobState.RUnlock()
	return calls
}