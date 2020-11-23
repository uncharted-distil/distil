package routes

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/uncharted-distil/distil/api/env"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
	// "github.com/uncharted-distil/distil-compute/model"
	"goji.io/v3/pat"
	"net/http"
	"path"
	// "image/color"
	"strconv"
	"strings"
	"fmt"
)

const (
	ThumbnailDimensions=125
)

func ImageAttentionHandler(ctor api.MetadataStorageCtor, pCtor api.DataStorageCtor) func(http.ResponseWriter, *http.Request) {
	dataset := pat.Param(r, "dataset")
	resultID := pat.Param(r, "resultId")
	d3mIndex := pat.Param(r, "index")
	isThumbnail, err := strconv.ParseBool(pat.Param(r, "is-thumbnail"))
	if err != nil {
		handleError(w, err)
		return
	}
	storage, err:=ctor()
	if err != nil {
		handleError(w, err)
		return
	}
}