package task

import (
	"fmt"

	"github.com/unchartedsoftware/distil/api/compute/description"
)

// ClassifyPrimmitive will classify the dataset using a primitive.
func ClassifyPrimmitive(index string, dataset string, config *IngestTaskConfig) error {
	step := description.NewSimonStep()
	fmt.Printf("%v", step)

	return nil
}
