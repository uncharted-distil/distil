package task

import (
	"encoding/json"
	"io/ioutil"

    "github.com/pkg/errors"

    "github.com/unchartedsoftware/distil/api/compute"
	"github.com/unchartedsoftware/distil/api/compute/description"
    "github.com/unchartedsoftware/distil/api/pipeline"
    "github.com/unchartedsoftware/distil-ingest/rest"
)

var (
    client *compute.Client
)

func submitPrimitive(dataset string, step *pipeline.PipelineDescription) (string, error) {
    request := compute.NewExecPipelineRequest(dataset, step)

    err := request.Dispatch(client)
    if err != nil {
        return "", errors.Wrap(err, "unable to dispatch pipeline")
    }

    // listen for completion
    var errPipeline error
    var datasetURI string
    err = request.Listen(func(status compute.ExecPipelineStatus) {
        // check for error
        if status.Error != nil {
            errPipeline = status.Error
        }

        if status.Progress == pipeline.ProgressState_name[int32(pipeline.ProgressState_COMPLETED)] {
            datasetURI = status.ResultURI
        }
    })

    if err != nil {
        return "", errors.Wrap(err, "unable to listen to pipeline")
    }

    if errPipeline != nil {
        return "", errors.Wrap(errPipeline, "error executing pipeline")
    }

    return datasetURI, nil
}

// ClassifyPrimmitive will classify the dataset using a primitive.
func ClassifyPrimmitive(index string, dataset string, config *IngestTaskConfig) error {
    // create & submit the solution request
	pip, err := description.CreateSimonPipeline("says", "")
    if err != nil {
        return errors.Wrap(err, "unable to create Simon pipeline")
    }

    datasetURI, err := submitPrimitive(dataset, pip)
    if err != nil {
        return errors.Wrap(err, "unable to run Simon pipeline")
    }

    // parse primitive response (d3mIndex,probabilities,labels)
    classification := &rest.ClassificationResult{}

    // output the classification in the expected JSON format
    bytes, err := json.MarshalIndent(classification, "", "    ")
    if err != nil {
        return errors.Wrap(err, "unable to serialize classification result")
    }
    // write to file
    err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
    if err != nil {
        return errors.Wrap(err, "unable to store classification result")
    }

	return nil
}

// RankPrimmitive will rank the dataset using a primitive.
func RankPrimmitive(index string, dataset string, config *IngestTaskConfig) error {
    // create & submit the solution request
	pip, err := description.CreatePunkPipeline("harry", "")
    if err != nil {
        return errors.Wrap(err, "unable to create Punk pipeline")
    }

    datasetURI, err := submitPrimitive(dataset, pip)
    if err != nil {
        return errors.Wrap(err, "unable to run Punk pipeline")
    }

    // parse primitive response (d3mIndex,probabilities,labels)
    classification := &rest.ClassificationResult{}

    // output the classification in the expected JSON format
    bytes, err := json.MarshalIndent(classification, "", "    ")
    if err != nil {
        return errors.Wrap(err, "unable to serialize classification result")
    }
    // write to file
    err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
    if err != nil {
        return errors.Wrap(err, "unable to store classification result")
    }

	return nil
}

// SummarizePrimitive will summarize the dataset using a primitive.
func SummarizePrimitive(index string, dataset string, config *IngestTaskConfig) error {
    // create & submit the solution request
	pip, err := description.CreateDukePipeline("wellington", "")
    if err != nil {
        return errors.Wrap(err, "unable to create Duke pipeline")
    }

    datasetURI, err := submitPrimitive(dataset, pip)
    if err != nil {
        return errors.Wrap(err, "unable to run Duke pipeline")
    }

    // parse primitive response (d3mIndex,probabilities,labels)
    classification := &rest.ClassificationResult{}

    // output the classification in the expected JSON format
    bytes, err := json.MarshalIndent(classification, "", "    ")
    if err != nil {
        return errors.Wrap(err, "unable to serialize classification result")
    }
    // write to file
    err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
    if err != nil {
        return errors.Wrap(err, "unable to store classification result")
    }

	return nil
}

// FeaturizePrimitive will featurize the dataset fields using a primitive.
func FeaturizePrimitive(index string, dataset string, config *IngestTaskConfig) error {
    // create & submit the solution request
	pip, err := description.CreateCrocPipeline("leather", "")
    if err != nil {
        return errors.Wrap(err, "unable to create Croc pipeline")
    }

    datasetURI, err := submitPrimitive(dataset, pip)
    if err != nil {
        return errors.Wrap(err, "unable to run Croc pipeline")
    }

    // parse primitive response (d3mIndex,probabilities,labels)
    classification := &rest.ClassificationResult{}

    // output the classification in the expected JSON format
    bytes, err := json.MarshalIndent(classification, "", "    ")
    if err != nil {
        return errors.Wrap(err, "unable to serialize classification result")
    }
    // write to file
    err = ioutil.WriteFile(config.getTmpAbsolutePath(config.ClassificationOutputPathRelative), bytes, 0644)
    if err != nil {
        return errors.Wrap(err, "unable to store classification result")
    }

	return nil
}
