package routes

import (
	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	"goji.io/v3/pat"
	"net/http"
	"net/url"
	"strconv"
)

// ImageAttentionHandler provides an image filter for the supplied index
func ImageAttentionHandler(solutionCtor api.SolutionStorageCtor, metaCtor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dataset := pat.Param(r, "dataset")
		d3mIndex := pat.Param(r, "index")
		resultID, err := url.PathUnescape(pat.Param(r, "resultId"))
		if err != nil {
			handleError(w, err)
			return
		}
		index, err := strconv.Atoi(d3mIndex)
		if err != nil {
			handleError(w, err)
			return
		}
		dataStorage, err := solutionCtor()
		if err != nil {
			handleError(w, err)
			return
		}
		meta, err := metaCtor()
		if err != nil {
			handleError(w, err)
			return
		}

		ds, err := meta.FetchDataset(dataset, false, false)
		if err != nil {
			handleError(w, err)
			return
		}
		storageName := ds.StorageName
		// fetch data
		data, err := dataStorage.FetchExplainValues(dataset, storageName, []int{index}, resultID)
		if err != nil {
			handleError(w, err)
			return
		}
		for _, v := range data {
			scaledMatrix := util.ScaleConfidenceMatrix(ThumbnailDimensions, ThumbnailDimensions, &v.GradCAM)
			filter := util.ConfidenceMatrixToImage(scaledMatrix, util.ViridisColorScale, 100)
			imageBytes, err := util.ImageToPNG(filter)
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
