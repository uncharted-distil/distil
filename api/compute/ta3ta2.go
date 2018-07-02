package compute

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/proto"
	protobuf "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/unchartedsoftware/distil/api/pipeline"
	log "github.com/unchartedsoftware/plog"
)

const (
	unknownAPIVersion = "unknown"
)

var (
	// cached ta3ta2 API version
	apiVersion string
)

func convertMetricsFromTA3ToTA2(metrics []string) []*pipeline.ProblemPerformanceMetric {
	var res []*pipeline.ProblemPerformanceMetric
	for _, metric := range metrics {
		var metricSet pipeline.PerformanceMetric
		metricAdjusted, ok := pipeline.PerformanceMetric_value[strings.ToUpper(metric)]
		if !ok {
			log.Warnf("undefined performance metric found ('%s') so defaulting to undefined", metric)
			metricSet = pipeline.PerformanceMetric_METRIC_UNDEFINED
		} else {
			metricSet = pipeline.PerformanceMetric(metricAdjusted)
		}
		res = append(res, &pipeline.ProblemPerformanceMetric{
			Metric: metricSet,
		})
	}
	return res
}

func convertTaskTypeFromTA3ToTA2(taskType string) pipeline.TaskType {
	return pipeline.TaskType(pipeline.TaskType_value[strings.ToUpper(taskType)])
}

func convertTaskSubTypeFromTA3ToTA2(taskSubType string) pipeline.TaskSubtype {
	if taskSubType == "" {
		return pipeline.TaskSubtype_TASK_SUBTYPE_UNDEFINED
	}
	return pipeline.TaskSubtype(pipeline.TaskSubtype_value[strings.ToUpper(taskSubType)])
}

func convertTargetFeaturesTA3ToTA2(target string, targetIndex int) []*pipeline.ProblemTarget {
	return []*pipeline.ProblemTarget{
		{
			ColumnName:  target,
			ResourceId:  defaultResourceID,
			TargetIndex: int32(targetIndex),
			ColumnIndex: int32(targetIndex), // TODO: is this correct?
		},
	}
}

func convertDatasetTA3ToTA2(dataset string) string {
	return dataset
}

// GetAPIVersion retrieves the ta3-ta2 API version embedded in the pipeline_core.proto file.  This is
// a non-trivial operation, so the value is cached for quick access.
func GetAPIVersion() string {
	if apiVersion != "" {
		return apiVersion
	}

	// Get the raw file descriptor bytes
	fileDesc := proto.FileDescriptor(pipeline.E_ProtocolVersion.Filename)
	if fileDesc == nil {
		log.Errorf("failed to find file descriptor for %v", pipeline.E_ProtocolVersion.Filename)
		return unknownAPIVersion
	}

	// Open a gzip reader and decompress
	r, err := gzip.NewReader(bytes.NewReader(fileDesc))
	if err != nil {
		log.Errorf("failed to open gzip reader: %v", err)
		return unknownAPIVersion
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Errorf("failed to decompress descriptor: %v", err)
		return unknownAPIVersion
	}

	// Unmarshall the bytes from the proto format
	fd := &protobuf.FileDescriptorProto{}
	if err := proto.Unmarshal(b, fd); err != nil {
		log.Errorf("malformed FileDescriptorProto: %v", err)
		return unknownAPIVersion
	}

	// Fetch the extension from the FileDescriptorOptions message
	ex, err := proto.GetExtension(fd.GetOptions(), pipeline.E_ProtocolVersion)
	if err != nil {
		log.Errorf("failed to fetch extension: %v", err)
		return unknownAPIVersion
	}

	apiVersion = *ex.(*string)

	return apiVersion
}
