// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/recipe"
	"github.com/ONSdigital/dp-import-cantabular-dataset/handler"
	"sync"
)

// Ensure, that RecipeAPIClientMock does implement handler.RecipeAPIClient.
// If this is not the case, regenerate this file with moq.
var _ handler.RecipeAPIClient = &RecipeAPIClientMock{}

// RecipeAPIClientMock is a mock implementation of handler.RecipeAPIClient.
//
// 	func TestSomethingThatUsesRecipeAPIClient(t *testing.T) {
//
// 		// make and configure a mocked handler.RecipeAPIClient
// 		mockedRecipeAPIClient := &RecipeAPIClientMock{
// 			GetRecipeFunc: func(contextMoqParam context.Context, s1 string, s2 string, s3 string) (*recipe.Recipe, error) {
// 				panic("mock out the GetRecipe method")
// 			},
// 		}
//
// 		// use mockedRecipeAPIClient in code that requires handler.RecipeAPIClient
// 		// and then make assertions.
//
// 	}
type RecipeAPIClientMock struct {
	// GetRecipeFunc mocks the GetRecipe method.
	GetRecipeFunc func(contextMoqParam context.Context, s1 string, s2 string, s3 string) (*recipe.Recipe, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetRecipe holds details about calls to the GetRecipe method.
		GetRecipe []struct {
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
	lockGetRecipe sync.RWMutex
}

// GetRecipe calls GetRecipeFunc.
func (mock *RecipeAPIClientMock) GetRecipe(contextMoqParam context.Context, s1 string, s2 string, s3 string) (*recipe.Recipe, error) {
	if mock.GetRecipeFunc == nil {
		panic("RecipeAPIClientMock.GetRecipeFunc: method is nil but RecipeAPIClient.GetRecipe was just called")
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
	mock.lockGetRecipe.Lock()
	mock.calls.GetRecipe = append(mock.calls.GetRecipe, callInfo)
	mock.lockGetRecipe.Unlock()
	return mock.GetRecipeFunc(contextMoqParam, s1, s2, s3)
}

// GetRecipeCalls gets all the calls that were made to GetRecipe.
// Check the length with:
//     len(mockedRecipeAPIClient.GetRecipeCalls())
func (mock *RecipeAPIClientMock) GetRecipeCalls() []struct {
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
	mock.lockGetRecipe.RLock()
	calls = mock.calls.GetRecipe
	mock.lockGetRecipe.RUnlock()
	return calls
}
