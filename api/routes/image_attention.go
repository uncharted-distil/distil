package routes

import (
	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	"goji.io/v3/pat"
	"net/http"
	"strconv"
)
// ImageAttentionHandler provides an image filter for the supplied index
func ImageAttentionHandler(pCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
	dataset := pat.Param(r, "dataset")
	resultID := pat.Param(r, "resultId")
	d3mIndex := pat.Param(r, "index")

	index, err:= strconv.Atoi(d3mIndex)
	if err != nil {
		handleError(w, err)
		return
	}
	dataStorage, err:=pCtor()
	if err != nil {
		handleError(w, err)
		return
	}
	datasetName,err := dataStorage.GetStorageName(dataset)
	if err != nil {
		handleError(w, err)
		return
	}
	// fetch data 
	data,err:=dataStorage.FetchExplainValues(dataset, datasetName, []int{index}, resultID)
	if err != nil {
		handleError(w, err)
		return
	}
	for _,v:=range data{
		scaledMatrix:=util.ScaleConfidenceMatrix(ThumbnailDimensions, ThumbnailDimensions, &v.GradCAM)
		filter:=util.ConfidenceMatrixToImage(scaledMatrix, util.ViridisColorScale, 100)
		imageBytes, err := util.ImageToJPEG(filter)
		if err != nil {
			handleError(w, err)
			return
		}
		_, err = w.Write(imageBytes)
		if err != nil {
			handleError(w, errors.Wrap(err, "failed to write image resource bytes"))
			return
		}
	}
}
}