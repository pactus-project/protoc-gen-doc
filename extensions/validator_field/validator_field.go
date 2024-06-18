package extensions

import (
	"encoding/json"
	"reflect"
	"strings"

	validator "github.com/mwitkow/go-proto-validators"
	"github.com/pactus-project/protoc-gen-doc/extensions"
)

// ValidatorRule represents a single validator rule from the (validator.field) method option extension.
type ValidatorRule struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// ValidatorExtension contains the rules set by the (validator.field) method option extension.
type ValidatorExtension struct {
	*validator.FieldValidator
	rules []ValidatorRule // memoized so that we don't have to use reflection more than we need.
}

// MarshalJSON implements the json.Marshaler interface.
func (v ValidatorExtension) MarshalJSON() ([]byte, error) { return json.Marshal(v.Rules()) }

// Rules returns all active rules
func (v ValidatorExtension) Rules() []ValidatorRule {
	if v.FieldValidator == nil {
		return nil
	}
	if v.rules != nil {
		return v.rules
	}
	vv := reflect.ValueOf(*v.FieldValidator)
	vt := vv.Type()
	for i := 0; i < vt.NumField(); i++ {
		tag, ok := vt.Field(i).Tag.Lookup("protobuf")
		if !ok {
			continue
		}
		for _, opt := range strings.Split(tag, ",") {
			if strings.HasPrefix(opt, "name=") {
				tag = strings.TrimPrefix(opt, "name=")
				break
			}
		}
		value := vv.Field(i)
		if value.IsNil() {
			continue
		}
		value = reflect.Indirect(value)
		v.rules = append(v.rules, ValidatorRule{Name: tag, Value: value.Interface()})
	}
	return v.rules
}

func init() {
	extensions.SetTransformer("validator.field", func(payload interface{}) interface{} {
		validator, ok := payload.(*validator.FieldValidator)
		if !ok {
			return nil
		}
		return ValidatorExtension{FieldValidator: validator}
	})
}
