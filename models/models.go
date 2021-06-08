package models

// File copied from dp-recipe-api v2

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ONSdigital/log.go/log"
)

//RecipeResults - struct for list of recipes
type RecipeResults struct {
	Count      int       `json:"count"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	TotalCount int       `json:"total_count"`
	Items      []*Recipe `json:"items"`
}

//Recipe - struct for individual recipe
type Recipe struct {
	ID              string     `bson:"_id,omitempty" json:"id,omitempty"`
	Alias           string     `bson:"alias,omitempty" json:"alias,omitempty"`
	Format          string     `bson:"format,omitempty" json:"format,omitempty"`
	InputFiles      []file     `bson:"files,omitempty" json:"files,omitempty"`
	OutputInstances []Instance `bson:"output_instances,omitempty" json:"output_instances,omitempty"`
	CantabularBlob  string     `bson:"cantabular_blob,omitempty" json:"cantabular_blob,omitempty"`
}

//CodeList - Code lists for instance
type CodeList struct {
	ID          string `bson:"id,omitempty" json:"id,omitempty"`
	HRef        string `bson:"href,omitempty" json:"href,omitempty"`
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	IsHierarchy *bool  `bson:"is_hierarchy,omitempty" json:"is_hierarchy,omitempty"`
}

//Instance - struct for instance of recipe
type Instance struct {
	DatasetID string     `bson:"dataset_id,omitempty" json:"dataset_id,omitempty"`
	Editions  []string   `bson:"editions,omitempty" json:"editions,omitempty"`
	Title     string     `bson:"title,omitempty" json:"title,omitempty"`
	CodeLists []CodeList `bson:"code_lists,omitempty" json:"code_lists,omitempty"`
}

type file struct {
	Description string `bson:"description,omitempty" json:"description,omitempty"`
}

//HRefURL - the href of all current code lists
const HRefURL = "http://localhost:22400/code-lists/"

var (
	validFormats = map[string]bool{
		"v4":               true,
		"cantabular-blob":  true,
		"cantabular-table": true,
	}
)

//validateInstance - checks if fields of OutputInstances are not empty for ValidateAddRecipe and ValidateAddInstance
func (instance *Instance) validateInstance(ctx context.Context) (missingFields []string, invalidFields []string) {

	if instance.DatasetID == "" {
		missingFields = append(missingFields, "dataset_id")
	}

	if instance.Editions != nil && len(instance.Editions) > 0 {
		for j, edition := range instance.Editions {
			if edition == "" {
				missingFields = append(missingFields, "editions["+strconv.Itoa(j)+"]")
			}
		}
	} else {
		missingFields = append(missingFields, "editions")
	}

	if instance.Title == "" {
		missingFields = append(missingFields, "title")
	}

	if instance.CodeLists != nil && len(instance.CodeLists) > 0 {
		for i, codelist := range instance.CodeLists {
			codelistMissingFields, codelistInvalidFields := codelist.validateCodelist(ctx)

			if len(codelistMissingFields) > 0 {
				for mIndex, mField := range codelistMissingFields {
					codelistMissingFields[mIndex] = "code-lists[" + strconv.Itoa(i) + "]." + mField
				}
			}
			if len(codelistInvalidFields) > 0 {
				for iIndex, iField := range codelistInvalidFields {
					codelistInvalidFields[iIndex] = "code-lists[" + strconv.Itoa(i) + "]." + iField
				}
			}

			missingFields = append(missingFields, codelistMissingFields...)
			invalidFields = append(invalidFields, codelistInvalidFields...)
		}
	} else {
		missingFields = append(missingFields, "code-lists")
	}

	return missingFields, invalidFields
}

//validateCodelists - checks if fields of CodeList are not empty for ValidateAddRecipe, ValidateAddInstance, ValidateAddCodelist
func (codelist *CodeList) validateCodelist(ctx context.Context) (missingFields []string, invalidFields []string) {

	if codelist.ID == "" {
		missingFields = append(missingFields, "id")
	}

	if codelist.HRef == "" {
		missingFields = append(missingFields, "href")
	} else {
		if !codelist.validateCodelistHRef(ctx) {
			invalidFields = append(invalidFields, "href should be in format (URL/id)")
		}
	}

	if codelist.Name == "" {
		missingFields = append(missingFields, "name")
	}

	if codelist.IsHierarchy == nil {
		missingFields = append(missingFields, "isHierarchy")
	}

	return missingFields, invalidFields
}

//ValidateCodelistHRef - checks if the format of the codelist.href is correct
func (codelist *CodeList) validateCodelistHRef(ctx context.Context) bool {
	hRefURL, err := url.Parse(codelist.HRef)
	if err != nil {
		log.Event(ctx, "error parsing codelist.href", log.ERROR, log.Error(err))
		return false
	}
	urlPathBool := strings.Contains(hRefURL.Path, "/code-lists") && strings.Contains(hRefURL.Path, codelist.ID)
	if hRefURL.Scheme != "" && hRefURL.Host != "" && urlPathBool {
		return true
	}
	return false
}

//ValidateAddRecipe - checks if all the fields of the recipe are non-empty
func (recipe *Recipe) ValidateAddRecipe(ctx context.Context) error {
	var missingFields []string
	var invalidFields []string

	//recipe.ID generated by API if ID not given so never missing (generates a V4 UUID)

	if recipe.Alias == "" {
		missingFields = append(missingFields, "alias")
	}
	if recipe.Format == "" {
		missingFields = append(missingFields, "format")
	} else {
		if !validFormats[recipe.Format] {
			invalidFields = append(invalidFields, "format is not valid")
		}
	}

	if recipe.Format == "cantabular-blob" && recipe.CantabularBlob == "" {
		missingFields = append(missingFields, "cantabular-blob")
	} else {
		if recipe.Format == "cantabular-table" && recipe.CantabularBlob == "" {
			missingFields = append(missingFields, "cantabular-table")
		}
	}

	if recipe.Format == "v4" {
		if recipe.InputFiles != nil && len(recipe.InputFiles) > 0 {
			for i, file := range recipe.InputFiles {
				if file.Description == "" {
					missingFields = append(missingFields, "input-files["+strconv.Itoa(i)+"].description")
				}
			}
		} else {
			missingFields = append(missingFields, "input-files")
		}
	}

	if recipe.OutputInstances != nil && len(recipe.OutputInstances) > 0 {
		for i, instance := range recipe.OutputInstances {
			instanceMissingFields, instanceInvalidFields := instance.validateInstance(ctx)
			if len(instanceMissingFields) > 0 {
				for mIndex, mField := range instanceMissingFields {
					instanceMissingFields[mIndex] = "output-instances[" + strconv.Itoa(i) + "]." + mField
				}
			}
			if len(instanceInvalidFields) > 0 {
				for iIndex, iField := range instanceInvalidFields {
					instanceInvalidFields[iIndex] = "output-instances[" + strconv.Itoa(i) + "]." + iField
				}
			}
			missingFields = append(missingFields, instanceMissingFields...)
			invalidFields = append(invalidFields, instanceInvalidFields...)
		}
	} else {
		missingFields = append(missingFields, "output-instances")
	}

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil

}

//ValidateAddInstance - checks if fields of OutputInstances are not empty
func (instance *Instance) ValidateAddInstance(ctx context.Context) error {
	var missingFields []string
	var invalidFields []string

	instanceMissingField, instanceInvalidField := instance.validateInstance(ctx)
	missingFields = append(missingFields, instanceMissingField...)
	invalidFields = append(invalidFields, instanceInvalidField...)

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil
}

//ValidateAddCodelist - checks if fields of Codelist are not empty
func (codelist *CodeList) ValidateAddCodelist(ctx context.Context) error {
	var missingFields []string
	var invalidFields []string

	codelistMissingFields, codelistinvalidFields := codelist.validateCodelist(ctx)
	missingFields = append(missingFields, codelistMissingFields...)
	invalidFields = append(invalidFields, codelistinvalidFields...)

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil
}

//ValidateUpdateRecipe - checks updates of recipe for PUT request
func (recipe *Recipe) ValidateUpdateRecipe(ctx context.Context) error {
	var missingFields []string
	var invalidFields []string

	// Validation to check at least one field of the recipe is updated
	allStringFieldsEmpty := recipe.ID == "" && recipe.Format == "" && recipe.Alias == ""
	allRemainingFieldsNil := recipe.InputFiles == nil && recipe.OutputInstances == nil
	if allStringFieldsEmpty && allRemainingFieldsNil {
		invalidFields = append(invalidFields, "no recipe fields updates given")
	}

	if recipe.ID != "" {
		invalidFields = append(invalidFields, "id cannot be changed")
	}
	if recipe.Format != "" {
		if !validFormats[recipe.Format] {
			invalidFields = append(invalidFields, "format is not valid")
		}
	}

	if recipe.InputFiles != nil && len(recipe.InputFiles) > 0 {
		for i, file := range recipe.InputFiles {
			if file.Description == "" {
				invalidFields = append(invalidFields, "empty input-files["+strconv.Itoa(i)+"].description given")
			}
		}
	}
	if recipe.InputFiles != nil && len(recipe.InputFiles) == 0 {
		invalidFields = append(invalidFields, "empty input-files update given")
	}

	//When doing the update, as recipe.OutputInstances is an array, it needs to make sure that all fields of the instance are complete
	//This functionality is already available in validateInstance
	if recipe.OutputInstances != nil && len(recipe.OutputInstances) > 0 {
		for i, instance := range recipe.OutputInstances {
			instanceMissingFields, instanceInvalidFields := instance.validateInstance(ctx)
			if len(instanceMissingFields) > 0 {
				for mIndex, mField := range instanceMissingFields {
					instanceMissingFields[mIndex] = "output-instances[" + strconv.Itoa(i) + "]." + mField
				}
			}
			if len(instanceInvalidFields) > 0 {
				for iIndex, iField := range instanceInvalidFields {
					instanceInvalidFields[iIndex] = "output-instances[" + strconv.Itoa(i) + "]." + iField
				}
			}
			missingFields = append(missingFields, instanceMissingFields...)
			invalidFields = append(invalidFields, instanceInvalidFields...)
		}
	}
	if recipe.OutputInstances != nil && len(recipe.OutputInstances) == 0 {
		invalidFields = append(invalidFields, "empty output-instances update given")
	}

	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil

}

//ValidateUpdateInstance - checks fields of instance before updating the instance of the recipe
func (instance *Instance) ValidateUpdateInstance(ctx context.Context) error {
	var missingFields []string
	var invalidFields []string

	// Validation to check if at least one instance field is updated
	allStringFieldsEmpty := instance.DatasetID == "" && instance.Title == ""
	allRemainingFieldsNil := instance.Editions == nil && instance.CodeLists == nil
	if allStringFieldsEmpty && allRemainingFieldsNil {
		invalidFields = append(invalidFields, "no instance fields updates given")
	}

	if instance.Editions != nil && len(instance.Editions) > 0 {
		for j, edition := range instance.Editions {
			if edition == "" {
				missingFields = append(missingFields, "editions["+strconv.Itoa(j)+"]")
			}
		}
	}
	if instance.Editions != nil && len(instance.Editions) == 0 {
		missingFields = append(missingFields, "editions")
	}

	//When doing the update, as instance.Codelist is an array, it needs to make sure that all fields of the codelist are complete
	//This functionality is already available in validateCodelists
	if instance.CodeLists != nil && len(instance.CodeLists) > 0 {
		for i, codelist := range instance.CodeLists {
			codelistMissingFields, codelistInvalidFields := codelist.validateCodelist(ctx)
			if len(codelistMissingFields) > 0 {
				for mIndex, mField := range codelistMissingFields {
					codelistMissingFields[mIndex] = "code-lists[" + strconv.Itoa(i) + "]." + mField
				}
			}
			if len(codelistInvalidFields) > 0 {
				for iIndex, iField := range codelistInvalidFields {
					codelistInvalidFields[iIndex] = "code-lists[" + strconv.Itoa(i) + "]." + iField
				}
			}
			missingFields = append(missingFields, codelistMissingFields...)
			invalidFields = append(invalidFields, codelistInvalidFields...)
		}
	}

	// If any fields missing from the code lists of the instance
	if missingFields != nil {
		return fmt.Errorf("missing mandatory fields: %v", missingFields)
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil

}

//ValidateUpdateCodeList - checks fields of codelist before updating the codelist in instance of the recipe
func (codelist *CodeList) ValidateUpdateCodeList(ctx context.Context) error {
	var invalidFields []string

	// Validation to check if at least one codelist field is updated
	allStringFieldsEmpty := codelist.ID == "" && codelist.Name == "" && codelist.HRef == ""
	if allStringFieldsEmpty && codelist.IsHierarchy == nil {
		invalidFields = append(invalidFields, "no codelist fields updates given")
	}

	if codelist.HRef != "" {
		if !codelist.validateCodelistHRef(ctx) {
			invalidFields = append(invalidFields, "href should be in format (URL/id)")
		}
	}

	if invalidFields != nil {
		return fmt.Errorf("invalid fields: %v", invalidFields)
	}

	return nil

}