//
//   Copyright © 2021 Uncharted Software Inc.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package task

import (
	"strings"

	"github.com/pkg/errors"
	ingestMetadata "github.com/uncharted-distil/distil-compute/metadata"
	"github.com/uncharted-distil/distil-compute/model"
	"github.com/uncharted-distil/distil-compute/primitive/compute"
	"github.com/uncharted-distil/distil-compute/primitive/compute/description"
	"github.com/uncharted-distil/distil/api/env"
	apiModel "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/serialization"
)

type primitiveSubmitter interface {
	submit(datasetURIs []string, pipelineDesc *description.FullySpecifiedPipeline) (string, error)
}

// JoinSpec stores information for one side of a join operation.
type JoinSpec struct {
	DatasetID        string
	DatasetPath      string
	DatasetSource    ingestMetadata.DatasetSource
	ExistingMetadata *model.Metadata
	UpdatedVariables []*model.Variable
}

// JoinPair captures the information required for a single join relationship.
type JoinPair struct {
	Left             string
	Right            string
	Accuracy         float64
	AbsoluteAccuracy bool
}

// JoinDatamart will make all your dreams come true.
func JoinDatamart(joinLeft *JoinSpec, joinRight *JoinSpec, rightOrigin *model.DatasetOrigin) (string, *apiModel.FilteredData, error) {
	pipelineDesc, err := description.CreateDatamartAugmentPipeline("Join Preview",
		"Join to be reviewed by user", rightOrigin.SearchResult, rightOrigin.Provenance)
	if err != nil {
		return "", nil, err
	}
	datasetLeftURI := env.ResolvePath(joinLeft.DatasetSource, joinLeft.DatasetPath)

	return join(joinLeft, joinRight, pipelineDesc, []string{datasetLeftURI}, defaultSubmitter{}, false)
}

// JoinDistil will bring misery.
func JoinDistil(dataStorage apiModel.DataStorage, joinLeft *JoinSpec, joinRight *JoinSpec, joinPairs []*JoinPair, returnRaw bool) (string, *apiModel.FilteredData, error) {
	isKey := false
	varsLeftMapUpdated := mapDistilJoinVars(joinLeft.UpdatedVariables)
	varsRightMapUpdated := mapDistilJoinVars(joinRight.UpdatedVariables)
	joins := make([]*description.Join, len(joinPairs))
	rightVars := make([]*model.Variable, len(joinPairs))
	for i := range joinPairs {
		joins[i] = &description.Join{
			Left:     varsLeftMapUpdated[joinPairs[i].Left],
			Right:    varsRightMapUpdated[joinPairs[i].Right],
			Accuracy: joinPairs[i].Accuracy,
			Absolute: joinPairs[i].AbsoluteAccuracy,
		}
		rightVars[i] = varsRightMapUpdated[joinPairs[i].Right]

		// assume groupings are valid keys for the join
		if joins[i].Right.IsGrouping() {
			isKey = true
		}
	}
	var err error
	if !isKey {
		isKey, err = dataStorage.IsKey(joinRight.DatasetID, joinRight.ExistingMetadata.StorageName, rightVars)
		if err != nil {
			return "", nil, err
		}
	}
	if !isKey {
		return "", nil, errors.Errorf("specified right join columns do not specify a unique key")
	}

	rightExcludes := generateRightExcludes(joinLeft.UpdatedVariables, joinRight.UpdatedVariables, joinPairs)
	joinInfo := &description.JoinDescription{
		Joins:          joins,
		LeftExcludes:   []*model.Variable{},
		LeftVariables:  joinLeft.UpdatedVariables,
		RightExcludes:  rightExcludes,
		RightVariables: joinRight.UpdatedVariables,
		Type:           description.JoinTypeLeft,
	}
	pipelineDesc, err := description.CreateJoinPipeline("Joiner", "Join existing data", joinInfo)
	if err != nil {
		return "", nil, err
	}

	datasetLeftURI := joinLeft.DatasetPath
	datasetRightURI := joinRight.DatasetPath

	return join(joinLeft, joinRight, pipelineDesc, []string{datasetLeftURI, datasetRightURI}, defaultSubmitter{}, returnRaw)
}

func join(joinLeft *JoinSpec, joinRight *JoinSpec, pipelineDesc *description.FullySpecifiedPipeline,
	datasetURIs []string, submitter primitiveSubmitter, returnRaw bool) (string, *apiModel.FilteredData, error) {

	// returns a URI pointing to the merged CSV file
	resultURI, err := submitter.submit(datasetURIs, pipelineDesc)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to run join pipeline")
	}

	// create a new dataset from the merged CSV file
	csvFilename := strings.TrimPrefix(resultURI, "file://")
	leftName := joinLeft.DatasetID
	rightName := joinRight.DatasetID
	datasetName := strings.Join([]string{leftName, rightName}, "-")
	storageName := model.NormalizeDatasetID(datasetName)
	mergedVariables, err := createDatasetFromCSV(csvFilename, datasetName, storageName, joinLeft, joinRight)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create dataset from result CSV")
	}

	// return some of the data for the client to preview
	data, err := createFilteredData(csvFilename, mergedVariables, returnRaw, 100)
	if err != nil {
		return "", nil, err
	}

	return env.ResolvePath(ingestMetadata.Augmented, datasetName), data, nil
}

type defaultSubmitter struct{}

func (defaultSubmitter) submit(datasetURIs []string, pipelineDesc *description.FullySpecifiedPipeline) (string, error) {
	return submitPipeline(datasetURIs, pipelineDesc, true)
}

func createDatasetFromCSV(csvFile string, datasetName string, storageName string, joinLeft *JoinSpec, joinRight *JoinSpec) ([]*model.Variable, error) {
	inputData, err := serialization.ResultToInputCSV(csvFile)
	if err != nil {
		return nil, err
	}

	metadata := model.NewMetadata(datasetName, datasetName, datasetName, storageName)
	dataResource := model.NewDataResource(compute.DefaultResourceID, compute.D3MResourceType, map[string][]string{compute.D3MResourceFormat: {"csv"}})

	mergedVariables, referencedResources := joinMetadataVariables(inputData[0], joinLeft.ExistingMetadata, joinRight.ExistingMetadata)
	dataResource.Variables = mergedVariables
	inputData[0] = dataResource.GenerateHeader()

	metadata.DataResources = append(referencedResources, dataResource)

	outputPath := env.ResolvePath(ingestMetadata.Augmented, datasetName)

	rawDataset := &serialization.RawDataset{
		Name:     metadata.Name,
		ID:       metadata.ID,
		Metadata: metadata,
		Data:     inputData,
	}

	err = serialization.WriteDataset(outputPath, rawDataset)
	if err != nil {
		return nil, err
	}

	return mergedVariables, nil
}

func createFilteredData(csvFile string, variables []*model.Variable, returnRaw bool, lineCount int) (*apiModel.FilteredData, error) {
	datasetStorage := serialization.GetStorage(csvFile)
	inputData, err := datasetStorage.ReadData(csvFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read joined data")
	}

	return apiModel.CreateFilteredData(inputData, variables, returnRaw, lineCount)
}

func denormVariableName(variable *model.Variable) string {
	// The denorm primitive renames columns that refer to resources to "filename".
	// The maps should account for this.
	if variable.RefersTo != nil {
		return "filename"
	}

	return variable.HeaderName
}

func joinMetadataVariables(headerNames []string, leftMetadata *model.Metadata, rightMetadata *model.Metadata) ([]*model.Variable, []*model.DataResource) {
	// map the variables using the header names
	leftMap := apiModel.MapVariables(leftMetadata.GetMainDataResource().Variables, denormVariableName)
	rightMap := apiModel.MapVariables(rightMetadata.GetMainDataResource().Variables, denormVariableName)

	// left dataset takes priority in case of conflict
	mergedVariables := make([]*model.Variable, len(headerNames))
	mergedResources := []*model.DataResource{}
	for i, varName := range headerNames {
		v, ok := leftMap[varName]
		if ok && v.RefersTo != nil {
			mergedResources = append(mergedResources, getDataResource(leftMetadata, v.RefersTo["resID"].(string)))
		} else if !ok {
			v, ok = rightMap[varName]
			if ok && v.RefersTo != nil {
				mergedResources = append(mergedResources, getDataResource(rightMetadata, v.RefersTo["resID"].(string)))
			} else if !ok {
				// variable is probably an aggregation
				// create a new variable and default type to string
				// ingest process should be able to provide better info
				v = model.NewVariable(i, varName, varName, varName, varName, model.UnknownType,
					model.UnknownType, "", []string{"attribute"}, "data", nil, mergedVariables, false)
			}
		}

		v.Index = i
		mergedVariables[i] = v
	}

	return mergedVariables, mergedResources
}

func generateRightExcludes(leftVariables []*model.Variable, rightVariables []*model.Variable, joins []*JoinPair) []*model.Variable {
	// There is only allowed to be one set of geo coords after a join.  This is a constraint
	// driven by the UI, as having multiple bounds columns isn't properly handled by our
	// mapping approach.
	toRemove := map[string]bool{}
	for _, leftVar := range leftVariables {
		if leftVar.IsGrouping() && model.IsGeoBounds(leftVar.Type) {
			for _, rightVar := range rightVariables {
				if rightVar.IsGrouping() && model.IsGeoBounds(rightVar.Type) {
					gb := rightVar.Grouping.(*model.GeoBoundsGrouping)
					// don't need to remove polygon col here - we force the multiband image data to be the left
					// part of the join, and that's the only time the polygon col will be present
					toRemove[gb.CoordinatesCol] = true
					break
				}
			}
			break
		}
	}

	// right join columns should NOT be excluded since they are needed for the join itself
	for _, j := range joins {
		if toRemove[j.Right] {
			toRemove[j.Right] = false
		}
	}

	rightExcludes := []*model.Variable{}
	for _, v := range rightVariables {
		if toRemove[v.Key] {
			rightExcludes = append(rightExcludes, v)
		}
	}
	return rightExcludes
}

func mapDistilJoinVars(variables []*model.Variable) map[string]*model.Variable {
	varsMapped := apiModel.MapVariables(variables, func(variable *model.Variable) string { return variable.Key })

	// geobounds group should map the coordinates field to the grouping columns
	for _, g := range variables {
		if g.IsGrouping() && model.IsGeoBounds(g.Type) {
			geoGrouping := g.Grouping.(*model.GeoBoundsGrouping)
			varsMapped[geoGrouping.CoordinatesCol] = g
		}
	}

	return varsMapped
}
