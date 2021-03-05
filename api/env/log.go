//
//    Copyright © 2021 Uncharted Software Inc.
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package env

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/unchartedsoftware/plog"

	"github.com/uncharted-distil/distil/api/util"
)

const (
	searchSolutionMethod     = "/Core/SearchSolutions"
	searchSolutionGetMethod  = "/Core/GetSearchSolutionsResults"
	scoreSolutionMethod      = "/Core/ScoreSolution"
	scoreSolutionGetMethod   = "/Core/GetScoreSolutionResults"
	fitSolutionMethod        = "/Core/FitSolution"
	fitSolutionGetMethod     = "/Core/GetFitSolutionResults"
	produceSolutionMethod    = "/Core/ProduceSolution"
	produceSolutionGetMethod = "/Core/GetProduceSolutionResults"
	endSearchSolutionMethod  = "/Core/EndSearchSolutions"
)

var (
	mu     = &sync.Mutex{}
	logger *DiscoveryLogger

	featureMap = map[string]string{
		searchSolutionMethod:     "SearchSolutions",
		searchSolutionGetMethod:  "GetSearchSolutionsResults",
		scoreSolutionMethod:      "ScoreSolution",
		scoreSolutionGetMethod:   "GetScoreSolutionResults",
		fitSolutionMethod:        "FitSolution",
		fitSolutionGetMethod:     "GetFitSolutionResults",
		produceSolutionMethod:    "ProduceSolution",
		produceSolutionGetMethod: "GetProduceSolutionResults",
		endSearchSolutionMethod:  "EndSearchSolutions",
	}
	activityMap = map[string]string{
		searchSolutionMethod:     "MODEL_SELECTION",
		searchSolutionGetMethod:  "MODEL_SELECTION",
		scoreSolutionMethod:      "MODEL_SELECTION",
		scoreSolutionGetMethod:   "MODEL_SELECTION",
		fitSolutionMethod:        "MODEL_SELECTION",
		fitSolutionGetMethod:     "MODEL_SELECTION",
		produceSolutionMethod:    "MODEL_SELECTION",
		produceSolutionGetMethod: "MODEL_SELECTION",
		endSearchSolutionMethod:  "SYSTEM_ACTIVITY",
	}
	subActivityMap = map[string]string{
		searchSolutionMethod:     "MODEL_SEARCH",
		searchSolutionGetMethod:  "MODEL_SEARCH",
		scoreSolutionMethod:      "MODEL_SUMMARIZATION",
		scoreSolutionGetMethod:   "MODEL_SUMMARIZATION",
		fitSolutionMethod:        "MODEL_EXPLANATION",
		fitSolutionGetMethod:     "MODEL_EXPLANATION",
		produceSolutionMethod:    "MODEL_EXPLANATION",
		produceSolutionGetMethod: "MODEL_EXPLANATION",
		endSearchSolutionMethod:  "",
	}
)

// DiscoveryLogger logs problem discovery information.
type DiscoveryLogger struct {
	csvFilename string
	config      *Config
}

// NewDiscoveryLogger creates and initializes a discovery logger.
func NewDiscoveryLogger(filename string, config *Config) (*DiscoveryLogger, error) {
	logger = &DiscoveryLogger{
		config: config,
	}

	return logger.InitializeLog(filename)
}

// InitializeLog initializes the discovery log.
func (l *DiscoveryLogger) InitializeLog(filename string) (*DiscoveryLogger, error) {

	// write the logs to the output log directory
	csvFilename := path.Join(l.config.D3MOutputDir, "logs", filename)

	// initialize the log with the header
	err := util.WriteFileWithDirs(csvFilename, []byte("timestamp,feature_id,type,activity_l1,activity_l2,other\n"), os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize the activity log")
	}

	l.csvFilename = csvFilename

	return l, nil
}

// LogDatamartActionGlobal logs a datamart fuction call to the discovery log.
func LogDatamartActionGlobal(feature string, activity string, subActivity string) {
	logger.LogDatamartAction(feature, activity, subActivity)
}

// LogSystemAction logs a system action to the discovery log.
func (l *DiscoveryLogger) LogSystemAction(feature string, activity string, subActivity string, details string) {
	l.logAction(feature, "SYSTEM", activity, subActivity, details)
}

// LogAPIAction logs a TA2TA3 API call to the discovery log.
func (l *DiscoveryLogger) LogAPIAction(method string, params map[string]string) {
	// look up the feature, the activity and sub activity based on the grpc method
	feature := featureMap[method]
	if feature == "" {
		feature = method
	}
	l.logActionWithParams(feature, "TA23API", activityMap[method], subActivityMap[method], params)
}

// LogDatamartAction logs a datamart fuction call to the discovery log.
func (l *DiscoveryLogger) LogDatamartAction(feature string, activity string, subActivity string) {
	l.logAction(feature, "DATAMART", activity, subActivity, "")
}

func (l *DiscoveryLogger) logActionWithParams(feature string, typ string, activity string, subActivity string, params map[string]string) {
	paramsString, _ := json.Marshal(params)
	l.logAction(feature, typ, activity, subActivity, string(paramsString))
}

func (l *DiscoveryLogger) logAction(feature string, typ string, activity string, subActivity string, other string) {
	timestamp := fmt.Sprint(time.Now().Format(time.RFC3339))

	mu.Lock()
	defer mu.Unlock()
	f, err := os.OpenFile(l.csvFilename, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Errorf("unable to open discovery log: %v", err)
	}
	w := csv.NewWriter(f)

	err = w.Write([]string{timestamp, feature, typ, activity, subActivity, other})
	if err != nil {
		log.Errorf("unable to log to discovery log: %v", err)
	}

	w.Flush()
}
