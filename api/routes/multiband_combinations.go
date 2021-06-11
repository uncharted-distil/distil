//
//   Copyright © 2020 Uncharted Software Inc.
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

package routes

import (
	"net/http"

	"github.com/pkg/errors"
	api "github.com/uncharted-distil/distil/api/model"
	"github.com/uncharted-distil/distil/api/util"
)

// MultiBandCombinations provides a lit of combinations to be serialized to JSON for transport to the
// client.
type MultiBandCombinations struct {
	Combinations []*MultiBandCombinationDesc `json:"combinations"`
}

// MultiBandCombinationsHandler fetches a list of available band combination names for a given dataset.
func MultiBandCombinationsHandler(ctor api.MetadataStorageCtor) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// dataset := pat.Param(r, "dataset")
		// This currently just fetches band combos for sentinel 2 data so we don't read the dataset parameter yet.
		combinationsList := make([]*MultiBandCombinationDesc, len(util.SentinelBandCombinations))
		idx := 0
		for _, value := range util.SentinelBandCombinations {
			combinationsList[idx] = &MultiBandCombinationDesc{value.ID, value.DisplayName}
			idx++
		}
		combinations := MultiBandCombinations{combinationsList}
		err := handleJSON(w, combinations)
		if err != nil {
			handleError(w, errors.Wrap(err, "unable marshal result histogram into JSON"))
			return
		}
	}
}
