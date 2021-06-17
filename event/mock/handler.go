// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"sync"
)

// Ensure, that HandlerMock does implement event.Handler.
// If this is not the case, regenerate this file with moq.
var _ event.Handler = &HandlerMock{}

// HandlerMock is a mock implementation of event.Handler.
//
// 	func TestSomethingThatUsesHandler(t *testing.T) {
//
// 		// make and configure a mocked event.Handler
// 		mockedHandler := &HandlerMock{
// 			HandleFunc: func(contextMoqParam context.Context, instanceStarted *event.InstanceStarted) error {
// 				panic("mock out the Handle method")
// 			},
// 		}
//
// 		// use mockedHandler in code that requires event.Handler
// 		// and then make assertions.
//
// 	}
type HandlerMock struct {
	// HandleFunc mocks the Handle method.
	HandleFunc func(contextMoqParam context.Context, instanceStarted *event.InstanceStarted) error

	// calls tracks calls to the methods.
	calls struct {
		// Handle holds details about calls to the Handle method.
		Handle []struct {
			// ContextMoqParam is the contextMoqParam argument value.
			ContextMoqParam context.Context
			// InstanceStarted is the instanceStarted argument value.
			InstanceStarted *event.InstanceStarted
		}
	}
	lockHandle sync.RWMutex
}

// Handle calls HandleFunc.
func (mock *HandlerMock) Handle(contextMoqParam context.Context, instanceStarted *event.InstanceStarted) error {
	if mock.HandleFunc == nil {
		panic("HandlerMock.HandleFunc: method is nil but Handler.Handle was just called")
	}
	callInfo := struct {
		ContextMoqParam context.Context
		InstanceStarted *event.InstanceStarted
	}{
		ContextMoqParam: contextMoqParam,
		InstanceStarted: instanceStarted,
	}
	mock.lockHandle.Lock()
	mock.calls.Handle = append(mock.calls.Handle, callInfo)
	mock.lockHandle.Unlock()
	return mock.HandleFunc(contextMoqParam, instanceStarted)
}

// HandleCalls gets all the calls that were made to Handle.
// Check the length with:
//     len(mockedHandler.HandleCalls())
func (mock *HandlerMock) HandleCalls() []struct {
	ContextMoqParam context.Context
	InstanceStarted *event.InstanceStarted
} {
	var calls []struct {
		ContextMoqParam context.Context
		InstanceStarted *event.InstanceStarted
	}
	mock.lockHandle.RLock()
	calls = mock.calls.Handle
	mock.lockHandle.RUnlock()
	return calls
}
