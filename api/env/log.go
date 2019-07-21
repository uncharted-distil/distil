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
	log "github.com/unchartedsoftware/plog"
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
}

// InitializeLog initializes the discovery log.
func InitializeLog(filename string, config *Config) (*DiscoveryLogger, error) {

	if logger != nil {
		return nil, errors.Errorf("d3m system log already initialized")
	}

	// write the logs to the output log directory
	csvFilename := path.Join(config.D3MOutputDir, "logs", filename)

	// initialize the log with the header
	err := util.WriteFileWithDirs(csvFilename, []byte("timestamp,feature_id,type,activity_l1,activity_l2,other"), os.ModePerm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize the activity log")
	}

	logger = &DiscoveryLogger{
		csvFilename: csvFilename,
	}

	return logger, nil
}

func LogDatamartActionGlobal(feature string, activity string, subActivity string) {
	logger.logAction(feature, "DATAMART", activity, subActivity)
}

// LogSystemAction logs a system action to the discovery log.
func (l *DiscoveryLogger) LogSystemAction(feature string, activity string, subActivity string) {
	l.logAction(feature, "SYSTEM", activity, subActivity)
}

// LogAPIAction logs a TA2TA3 API call to the discovery log.
func (l *DiscoveryLogger) LogAPIAction(method string) {
	// look up the activity and sub activity based on the grpc method
	l.logAction(featureMap[method], "TA2TA3", activityMap[method], subActivityMap[method])
}

// LogDatamartAction logs a datamart fuction call to the discovery log.
func (l *DiscoveryLogger) LogDatamartAction(feature string, activity string, subActivity string) {
	l.logAction(feature, "DATAMART", activity, subActivity)
}

func (l *DiscoveryLogger) logAction(feature string, typ string, activity string, subActivity string) {
	log.Infof("GOT ME A LOG")
	timestamp := fmt.Sprintf(time.Now().Format(time.RFC3339))

	mu.Lock()
	defer mu.Unlock()
	f, _ := os.OpenFile(l.csvFilename, os.O_WRONLY|os.O_APPEND, os.ModePerm)
	w := csv.NewWriter(f)
	for i := 0; i < 10; i++ {
		w.Write([]string{timestamp, feature, typ, activity, subActivity, "{}"})
	}
	w.Flush()
}
