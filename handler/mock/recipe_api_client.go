// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-api-clients-go/v2/recipe"
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
// 			GetRecipeFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
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
	GetRecipeFunc func(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error)

	// calls tracks calls to the methods.
	calls struct {
		// GetRecipe holds details about calls to the GetRecipe method.
		GetRecipe []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// UserAuthToken is the userAuthToken argument value.
			UserAuthToken string
			// ServiceAuthToken is the serviceAuthToken argument value.
			ServiceAuthToken string
			// RecipeID is the recipeID argument value.
			RecipeID string
		}
	}
	lockGetRecipe sync.RWMutex
}

// GetRecipe calls GetRecipeFunc.
func (mock *RecipeAPIClientMock) GetRecipe(ctx context.Context, userAuthToken string, serviceAuthToken string, recipeID string) (*recipe.Recipe, error) {
	if mock.GetRecipeFunc == nil {
		panic("RecipeAPIClientMock.GetRecipeFunc: method is nil but RecipeAPIClient.GetRecipe was just called")
	}
	callInfo := struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		RecipeID         string
	}{
		Ctx:              ctx,
		UserAuthToken:    userAuthToken,
		ServiceAuthToken: serviceAuthToken,
		RecipeID:         recipeID,
	}
	mock.lockGetRecipe.Lock()
	mock.calls.GetRecipe = append(mock.calls.GetRecipe, callInfo)
	mock.lockGetRecipe.Unlock()
	return mock.GetRecipeFunc(ctx, userAuthToken, serviceAuthToken, recipeID)
}

// GetRecipeCalls gets all the calls that were made to GetRecipe.
// Check the length with:
//     len(mockedRecipeAPIClient.GetRecipeCalls())
func (mock *RecipeAPIClientMock) GetRecipeCalls() []struct {
	Ctx              context.Context
	UserAuthToken    string
	ServiceAuthToken string
	RecipeID         string
} {
	var calls []struct {
		Ctx              context.Context
		UserAuthToken    string
		ServiceAuthToken string
		RecipeID         string
	}
	mock.lockGetRecipe.RLock()
	calls = mock.calls.GetRecipe
	mock.lockGetRecipe.RUnlock()
	return calls
}
