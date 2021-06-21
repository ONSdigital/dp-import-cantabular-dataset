// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/cantabular"
	"sync"
)

// CantabularClientMock is a mock implementation of handler.CantabularClient.
//
// 	func TestSomethingThatUsesCantabularClient(t *testing.T) {
//
// 		// make and configure a mocked handler.CantabularClient
// 		mockedCantabularClient := &CantabularClientMock{
// 			GetCodebookFunc: func(contextMoqParam context.Context, getCodebookRequest cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
// 				panic("mock out the GetCodebook method")
// 			},
// 		}
//
// 		// use mockedCantabularClient in code that requires handler.CantabularClient
// 		// and then make assertions.
//
// 	}
type CantabularClientMock struct {
	// GetCodebookFunc mocks the GetCodebook method.
	GetCodebookFunc func(contextMoqParam context.Context, getCodebookRequest cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetCodebook holds details about calls to the GetCodebook method.
		GetCodebook []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// GetCodebookRequest is the getCodebookRequest argument value.
			GetCodebookRequest cantabular.GetCodebookRequest
		}
	}
	lockGetCodebook sync.RWMutex
}

// GetCodebook calls GetCodebookFunc.
func (mock *CantabularClientMock) GetCodebook(contextMoqParam context.Context, getCodebookRequest cantabular.GetCodebookRequest) (*cantabular.GetCodebookResponse, error) {
	if mock.GetCodebookFunc == nil {
		panic("CantabularClientMock.GetCodebookFunc: method is nil but CantabularClient.GetCodebook was just called")
	}
	callInfo := struct {
		ContextMoqParam    context.Context
		GetCodebookRequest cantabular.GetCodebookRequest
	}{
		ContextMoqParam:    contextMoqParam,
		GetCodebookRequest: getCodebookRequest,
	}
	mock.lockGetCodebook.Lock()
	mock.calls.GetCodebook = append(mock.calls.GetCodebook, callInfo)
	mock.lockGetCodebook.Unlock()
	return mock.GetCodebookFunc(contextMoqParam, getCodebookRequest)
}

// GetCodebookCalls gets all the calls that were made to GetCodebook.
// Check the length with:
//     len(mockedCantabularClient.GetCodebookCalls())
func (mock *CantabularClientMock) GetCodebookCalls() []struct {
	ContextMoqParam    context.Context
	GetCodebookRequest cantabular.GetCodebookRequest
} {
	var calls []struct {
		ContextMoqParam    context.Context
		GetCodebookRequest cantabular.GetCodebookRequest
	}
	mock.lockGetCodebook.RLock()
	calls = mock.calls.GetCodebook
	mock.lockGetCodebook.RUnlock()
	return calls
}
