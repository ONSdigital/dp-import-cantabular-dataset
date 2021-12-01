package event

import (
	"github.com/ONSdigital/dp-import/events"
)

// InstanceStarted provides an avro structure for an Instance Started event
type InstanceStarted events.CantabularDatasetInstanceStarted

// CategoryDimensionImport provies an avro structure for a Category Dimension Import event
type CategoryDimensionImport events.CantabularDatasetCategoryDimensionImport
