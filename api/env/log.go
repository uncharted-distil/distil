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

	l.csvFilename = filename

	return l, nil
}

// LogDatamartActionGlobal logs a datamart fuction call to the discovery log.
func LogDatamartActionGlobal(feature string, activity string, subActivity string) {
	logger.LogDatamartAction(feature, activity, subActivity)
}

// LogSystemAction logs a system action to the discovery log.
func (l *DiscoveryLogger) LogSystemAction(feature string, activity string, subActivity string) {
	l.logAction(feature, "SYSTEM", activity, subActivity, nil)
}

// LogAPIAction logs a TA2TA3 API call to the discovery log.
func (l *DiscoveryLogger) LogAPIAction(method string, params map[string]string) {
	// look up the feature, the activity and sub activity based on the grpc method
	feature := featureMap[method]
	if feature == "" {
		feature = method
	}
	l.logAction(feature, "TA2TA3", activityMap[method], subActivityMap[method], params)
}

// LogDatamartAction logs a datamart fuction call to the discovery log.
func (l *DiscoveryLogger) LogDatamartAction(feature string, activity string, subActivity string) {
	l.logAction(feature, "DATAMART", activity, subActivity, nil)
}

func (l *DiscoveryLogger) logAction(feature string, typ string, activity string, subActivity string, params map[string]string) {
	timestamp := fmt.Sprintf(time.Now().Format(time.RFC3339))
	paramsString, _ := json.Marshal(params)

	mu.Lock()
	defer mu.Unlock()
	f, _ := os.OpenFile(l.csvFilename, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	w := csv.NewWriter(f)
	w.Write([]string{timestamp, feature, typ, activity, subActivity, string(paramsString)})
	w.Flush()
}
