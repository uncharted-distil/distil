package env

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/pkg/errors"

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
	csvFilename    = ""
	mu             = &sync.Mutex{}
	initializedLog = false

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

func InitializeLog(filename string, config *Config) error {

	if initializedLog {
		return errors.Errorf("d3m system log already initialized")
	}

	// write the logs to the output log directory
	csvFilename = path.Join(config.D3MOutputDir, "logs", filename)

	// initialize the log with the header
	err := util.WriteFileWithDirs(csvFilename, []byte("timestamp,feature_id,type,activity_l1,activity_l2,other"), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "unable to initialize the activity log")
	}

	initializedLog = true

	return nil
}

func LogSystemAction(feature string, activity string, subActivity string) {
	logAction(feature, "SYSTEM", activity, subActivity)
}

func LogAPIAction(method string) {
	// look up the activity and sub activity based on the grpc method
	logAction(featureMap[method], "TA2TA3", activityMap[method], subActivityMap[method])
}

func LogDatamartAction(feature string, activity string, subActivity string) {
	logAction(feature, "DATAMART", activity, subActivity)
}

func logAction(feature string, typ string, activity string, subActivity string) {
	timestamp := fmt.Sprintf(time.Now().Format(time.RFC3339))

	mu.Lock()
	defer mu.Unlock()
	f, _ := os.OpenFile(csvFilename, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	w := csv.NewWriter(f)
	for i := 0; i < 10; i++ {
		w.Write([]string{timestamp, feature, typ, activity, subActivity, "{}"})
	}
	w.Flush()
}
