package description

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/unchartedsoftware/distil/api/pipeline"
)

const (
	stepInputsKey     = "inputs"
	pipelineInputsKey = "inputs"
)

// Step provides data for a pipeline description step and an operation
// to create a protobuf PipelineDescriptionStep from that data.
type Step interface {
	BuildDescriptionStep() (*pipeline.PipelineDescriptionStep, error)
	GetPrimitive() *pipeline.Primitive
	GetArguments() map[string]string
	UpdateArguments(string, string)
	GetHyperparameters() map[string]interface{}
	GetOutputMethods() []string
}

// StepData contains the minimum amount of data used to describe a pipeline step
type StepData struct {
	Primitive       *pipeline.Primitive
	Arguments       map[string]string
	Hyperparameters map[string]interface{}
	OutputMethods   []string
}

// NewStepData Creates a pipeline step instance from the required field subset.
func NewStepData(primitive *pipeline.Primitive, outputMethods []string) *StepData {
	return NewStepDataWithHyperparameters(primitive, outputMethods, nil)
}

// NewStepDataWithHyperparameters creates a pipeline step instance from the required field subset.  Hyperparameters are
// optional so nil is a valid value, valid types fror hyper parameters are intXX, string, bool.
func NewStepDataWithHyperparameters(primitive *pipeline.Primitive, outputMethods []string, hyperparameters map[string]interface{}) *StepData {
	return &StepData{
		Primitive:       primitive,
		Hyperparameters: hyperparameters, // optional, nil is valid
		Arguments:       map[string]string{},
		OutputMethods:   outputMethods,
	}
}

// GetPrimitive returns a primitive definition for a pipeline step.
func (s *StepData) GetPrimitive() *pipeline.Primitive {
	return s.Primitive
}

// GetArguments returns a map of arguments that will be passed to the methods
// of the primitive step.
func (s *StepData) GetArguments() map[string]string {
	copy := map[string]string{}
	for k, v := range s.Arguments {
		copy[k] = v
	}
	return copy
}

// UpdateArguments updates the arguments map that will be passed to the methods
// of primtive step.
func (s *StepData) UpdateArguments(key string, value string) {
	s.Arguments[key] = value
}

// GetHyperparameters returns a map of arguments that will be passed to the primitive methods
// of the primitive step.  Types are currently restricted to intXX, bool, string
func (s *StepData) GetHyperparameters() map[string]interface{} {
	return s.Hyperparameters
}

// GetOutputMethods returns a list of methods that will be called to generate
// primitive output.  These feed into downstream primitives.
func (s *StepData) GetOutputMethods() []string {
	return s.OutputMethods
}

// BuildDescriptionStep creates protobuf structures from a pipeline step
// definition.
func (s *StepData) BuildDescriptionStep() (*pipeline.PipelineDescriptionStep, error) {
	// // generate arguments entries
	// arguments := map[string]*pipeline.PrimitiveStepArgument{}
	// for k, v := range s.Arguments {
	// 	arguments[k] = &pipeline.PrimitiveStepArgument{
	// 		// only handle container args rights now - extend to others if required
	// 		Argument: &pipeline.PrimitiveStepArgument_Container{
	// 			Container: &pipeline.ContainerArgument{
	// 				Data: v,
	// 			},
	// 		},
	// 	}
	// }

	// // generate arguments entries - accepted types are currently intXX, string, bool.  The underlying
	// // protobuf structure allows for others - introducing them should be a matter of expanding this
	// // list.
	// hyperparameters := map[string]*pipeline.PrimitiveStepHyperparameter{}
	// for k, v := range s.Hyperparameters {
	// 	var value *pipeline.Value
	// 	switch t := v.(type) {
	// 	case int, int8, int16, int32, int64:
	// 		value = &pipeline.Value{
	// 			Value: &pipeline.Value_Int64{
	// 				Int64: reflect.ValueOf(t).Int(),
	// 			},
	// 		}
	// 	case []int, []int8, []int16, []int32, []int64:
	// 		arr := []int64{}
	// 		s := reflect.ValueOf(t)
	// 		if s.Kind() == reflect.Slice {
	// 			for i := 0; i < s.Len(); i++ {
	// 				arr = append(arr, s.Index(i).Int())
	// 			}
	// 		}
	// 		value = &pipeline.Value{
	// 			Value: &pipeline.Value_Int64List{
	// 				Int64List: &pipeline.Int64List{
	// 					List: arr,
	// 				},
	// 			},
	// 		}
	// 	case string:
	// 		value = &pipeline.Value{
	// 			Value: &pipeline.Value_String_{
	// 				String_: t,
	// 			},
	// 		}
	// 	case []string:
	// 		value = &pipeline.Value{
	// 			Value: &pipeline.Value_StringList{
	// 				StringList: &pipeline.StringList{
	// 					List: t,
	// 				},
	// 			},
	// 		}
	// 	case bool:
	// 		value = &pipeline.Value{
	// 			Value: &pipeline.Value_Bool{
	// 				Bool: t,
	// 			},
	// 		}
	// 	case []bool:
	// 		value = &pipeline.Value{
	// 			Value: &pipeline.Value_BoolList{
	// 				BoolList: &pipeline.BoolList{
	// 					List: t,
	// 				},
	// 			},
	// 		}
	// 	default:
	// 		return nil, errors.Errorf("compile failed: unhandled type `%v` for hyperparameter `%s`", v, k)
	// 	}
	// 	hyperparameters[k] = &pipeline.PrimitiveStepHyperparameter{
	// 		// only handle value args rights now - extend to others if required
	// 		Argument: &pipeline.PrimitiveStepHyperparameter_Value{
	// 			Value: &pipeline.ValueArgument{
	// 				Data: value,
	// 			},
	// 		},
	// 	}
	// }

	// // list of methods that will generate output - order matters because the steps are
	// // numbered
	// outputMethods := []*pipeline.StepOutput{}
	// for _, outputMethod := range s.OutputMethods {
	// 	outputMethods = append(outputMethods,
	// 		&pipeline.StepOutput{
	// 			Id: outputMethod,
	// 		})
	// }

	// // create the pipeline description structure
	// return &pipeline.PipelineDescriptionStep{
	// 	Step: &pipeline.PipelineDescriptionStep_Primitive{
	// 		Primitive: &pipeline.PrimitivePipelineDescriptionStep{
	// 			Primitive:   s.Primitive,
	// 			Arguments:   arguments,
	// 			Hyperparams: hyperparameters,
	// 			Outputs:     outputMethods,
	// 		},
	// 	},
	// }, nil
	return nil, nil
}

func parseList(list []interface{}) (*pipeline.ValueRaw, error) {
	valueList := []*pipeline.ValueRaw{}
	for _, v := range list {
		switch t := v.(type) {
		case int, int8, int16, int32, int64, string, bool:
			value, err := parseTerminal(t)
			valueList = append(valueList, value)
			if err != nil {
				return nil, err
			}
		case []interface{}:
			value, err := parseList(t)
			valueList = append(valueList, value)
			return nil, err
		case map[string]interface{}:
			value, err := parseMap(t)
			valueList = append(valueList, value)
			return nil, err
		default:
			return nil, errors.Errorf("bad argument type %s", reflect.TypeOf(v))
		}
	}
	v := &pipeline.ValueRaw{
		Raw: &pipeline.ValueRaw_List{
			List: &pipeline.ValueList{
				Items: valueList,
			},
		},
	}
	return v, nil
}

func parseMap(value map[string]interface{}) (*pipeline.ValueRaw, error) {
	return nil, nil
}

func parseTerminal(value interface{}) (*pipeline.ValueRaw, error) {
	return nil, nil
}
