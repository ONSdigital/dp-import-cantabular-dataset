// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/importapi"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	"sync"
)

// Ensure, that ImportAPIClientMock does implement handler.ImportAPIClient.
// If this is not the case, regenerate this file with moq.
var _ handler.ImportAPIClient = &ImportAPIClientMock{}

// ImportAPIClientMock is a mock implementation of handler.ImportAPIClient.
//
// 	func TestSomethingThatUsesImportAPIClient(t *testing.T) {
//
// 		// make and configure a mocked handler.ImportAPIClient
// 		mockedImportAPIClient := &ImportAPIClientMock{
// 			UpdateImportJobStateFunc: func(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error {
// 				panic("mock out the UpdateImportJobState method")
// 			},
// 		}
//
// 		// use mockedImportAPIClient in code that requires handler.ImportAPIClient
// 		// and then make assertions.
//
// 	}
type ImportAPIClientMock struct {
	// UpdateImportJobStateFunc mocks the UpdateImportJobState method.
	UpdateImportJobStateFunc func(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error

	// calls tracks calls to the methods.
	calls struct {
		// UpdateImportJobState holds details about calls to the UpdateImportJobState method.
		UpdateImportJobState []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// JobID is the jobID argument value.
			JobID string
			// ServiceToken is the serviceToken argument value.
			ServiceToken string
			// NewState is the newState argument value.
			NewState importapi.State
		}
	}
	lockUpdateImportJobState sync.RWMutex
}

// UpdateImportJobState calls UpdateImportJobStateFunc.
func (mock *ImportAPIClientMock) UpdateImportJobState(ctx context.Context, jobID string, serviceToken string, newState importapi.State) error {
	if mock.UpdateImportJobStateFunc == nil {
		panic("ImportAPIClientMock.UpdateImportJobStateFunc: method is nil but ImportAPIClient.UpdateImportJobState was just called")
	}
	callInfo := struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		NewState     importapi.State
	}{
		Ctx:          ctx,
		JobID:        jobID,
		ServiceToken: serviceToken,
		NewState:     newState,
	}
	mock.lockUpdateImportJobState.Lock()
	mock.calls.UpdateImportJobState = append(mock.calls.UpdateImportJobState, callInfo)
	mock.lockUpdateImportJobState.Unlock()
	return mock.UpdateImportJobStateFunc(ctx, jobID, serviceToken, newState)
}

// UpdateImportJobStateCalls gets all the calls that were made to UpdateImportJobState.
// Check the length with:
//     len(mockedImportAPIClient.UpdateImportJobStateCalls())
func (mock *ImportAPIClientMock) UpdateImportJobStateCalls() []struct {
	Ctx          context.Context
	JobID        string
	ServiceToken string
	NewState     importapi.State
} {
	var calls []struct {
		Ctx          context.Context
		JobID        string
		ServiceToken string
		NewState     importapi.State
	}
	mock.lockUpdateImportJobState.RLock()
	calls = mock.calls.UpdateImportJobState
	mock.lockUpdateImportJobState.RUnlock()
	return calls
}
