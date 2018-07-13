package description

import (
	"fmt"
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

	// generate arguments entries - accepted types are currently intXX, string, bool, as well as list, map[string]
	// of those types.  The underlying protobuf structure allows for others that can be handled here as needed.
	hyperparameters := map[string]*pipeline.PrimitiveStepHyperparameter{}
	for k, v := range s.Hyperparameters {
		rawValue, err := parseValue(v)
		if err != nil {
			return nil, errors.Errorf("compile failed: hyperparameter `%s` - %s", k, err.Error())
		}

		hyperparameters[k] = &pipeline.PrimitiveStepHyperparameter{
			// only handle value args rights now - extend to others if required
			Argument: &pipeline.PrimitiveStepHyperparameter_Value{
				Value: &pipeline.ValueArgument{
					Data: &pipeline.Value{
						Value: &pipeline.Value_Raw{
							Raw: rawValue,
						},
					},
				},
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

func parseList(v interface{}) (*pipeline.ValueRaw, error) {
	// parse list contents as a list, map, or value
	valueList := []*pipeline.ValueRaw{}
	var value *pipeline.ValueRaw
	var err error

	// type switches to work well with generic arrays/maps so we have to revert to using reflection
	refValue := reflect.ValueOf(v)
	if refValue.Kind() != reflect.Slice {
		return nil, errors.Errorf("unexpected parameter %s", refValue.Kind())
	}
	for i := 0; i < refValue.Len(); i++ {
		refElement := refValue.Index(i)
		switch refElement.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.String, reflect.Bool:
			value, err = parseValue(refElement.Interface())
		case reflect.Slice:
			value, err = parseList(refElement.Interface())
		case reflect.Map:
			value, err = parseMap(refElement.Interface())
		default:
			err = errors.Errorf("unhandled list arg type %s", refElement.Kind())
		}

		if err != nil {
			return nil, err
		}

		valueList = append(valueList, value)
	}
	rawValue := &pipeline.ValueRaw{
		Raw: &pipeline.ValueRaw_List{
			List: &pipeline.ValueList{
				Items: valueList,
			},
		},
	}
	return rawValue, nil
}

func parseMap(vmap interface{}) (*pipeline.ValueRaw, error) {
	// parse map contents as list, map or value
	valueMap := map[string]*pipeline.ValueRaw{}
	var value *pipeline.ValueRaw
	var err error

	// type switches to work well with generic arrays/maps so we have to revert to using reflection
	refValue := reflect.ValueOf(vmap)
	if refValue.Kind() != reflect.Map {
		return nil, errors.Errorf("unexpected parameter %s", refValue.Kind())
	}
	keys := refValue.MapKeys()
	for _, key := range keys {
		refElement := refValue.MapIndex(key)
		switch refElement.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.String, reflect.Bool:
			value, err = parseValue(refElement.Interface)
		case reflect.Slice:
			value, err = parseList(refElement.Interface)
		case reflect.Map:
			value, err = parseMap(refElement.Interface)
		default:
			err = errors.Errorf("unhandled map arg type %s", refElement.Kind())
		}

		if err != nil {
			return nil, err
		}
		refValue.SetMapIndex(key, reflect.ValueOf(value))
	}

	v := &pipeline.ValueRaw{
		Raw: &pipeline.ValueRaw_Dict{
			Dict: &pipeline.ValueDict{
				Items: valueMap,
			},
		},
	}
	return v, nil
}

func parseValue(v interface{}) (*pipeline.ValueRaw, error) {
	refValue := reflect.ValueOf(v)
	switch refValue.Kind() {
	// parse a numeric, string or boolean value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &pipeline.ValueRaw{
			Raw: &pipeline.ValueRaw_Int64{
				Int64: refValue.Int(),
			},
		}, nil
	case reflect.String:
		return &pipeline.ValueRaw{
			Raw: &pipeline.ValueRaw_String_{
				String_: refValue.String(),
			},
		}, nil
	case reflect.Bool:
		return &pipeline.ValueRaw{
			Raw: &pipeline.ValueRaw_Bool{
				Bool: refValue.Bool(),
			},
		}, nil
	case reflect.Slice:
		fmt.Printf("%v\n", v)
		return parseList(v)
	case reflect.Map:
		return parseMap(v)
	default:
		return nil, errors.Errorf("unhandled value arg type %s", refValue.Kind())
	}
}
