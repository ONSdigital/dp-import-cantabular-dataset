// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-import-cantabular-dataset/config"
	"github.com/ONSdigital/dp-import-cantabular-dataset/event"
	"sync"
)

var (
	lockHandlerMockHandle sync.RWMutex
)

// Ensure, that HandlerMock does implement event.Handler.
// If this is not the case, regenerate this file with moq.
var _ event.Handler = &HandlerMock{}

// HandlerMock is a mock implementation of event.Handler.
//
//     func TestSomethingThatUsesHandler(t *testing.T) {
//
//         // make and configure a mocked event.Handler
//         mockedHandler := &HandlerMock{
//             HandleFunc: func(ctx context.Context, cfg *config.Config, helloCalled *event.HelloCalled) error {
// 	               panic("mock out the Handle method")
//             },
//         }
//
//         // use mockedHandler in code that requires event.Handler
//         // and then make assertions.
//
//     }
type HandlerMock struct {
	// HandleFunc mocks the Handle method.
	HandleFunc func(ctx context.Context, cfg *config.Config, helloCalled *event.HelloCalled) error

	// calls tracks calls to the methods.
	calls struct {
		// Handle holds details about calls to the Handle method.
		Handle []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// HelloCalled is the helloCalled argument value.
			HelloCalled *event.HelloCalled
		}
	}
}

// Handle calls HandleFunc.
func (mock *HandlerMock) Handle(ctx context.Context, cfg *config.Config, helloCalled *event.HelloCalled) error {
	if mock.HandleFunc == nil {
		panic("HandlerMock.HandleFunc: method is nil but Handler.Handle was just called")
	}
	callInfo := struct {
		Ctx         context.Context
		Cfg         *config.Config
		HelloCalled *event.HelloCalled
	}{
		Ctx:         ctx,
		Cfg:         cfg,
		HelloCalled: helloCalled,
	}
	lockHandlerMockHandle.Lock()
	mock.calls.Handle = append(mock.calls.Handle, callInfo)
	lockHandlerMockHandle.Unlock()
	return mock.HandleFunc(ctx, cfg, helloCalled)
}

// HandleCalls gets all the calls that were made to Handle.
// Check the length with:
//     len(mockedHandler.HandleCalls())
func (mock *HandlerMock) HandleCalls() []struct {
	Ctx         context.Context
	Cfg         *config.Config
	HelloCalled *event.HelloCalled
} {
	var calls []struct {
		Ctx         context.Context
		Cfg         *config.Config
		HelloCalled *event.HelloCalled
	}
	lockHandlerMockHandle.RLock()
	calls = mock.calls.Handle
	lockHandlerMockHandle.RUnlock()
	return calls
}
