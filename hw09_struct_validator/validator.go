package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

type intCheckFn func(int, int) bool

type validationFn func(ruleVal string, val reflect.Value, name string) error

var (
	ErrNotStruct   = errors.New("input is not a struct")
	ErrInvalidRule = errors.New("invalid validation rule")
	ErrWrongType   = errors.New("rule unavailable for this field type")

	ErrInvalidLen     = errors.New("string len is not valid")
	ErrNotInSet       = errors.New("value is not in validated set")
	ErrLessMin        = errors.New("value is less than minimal")
	ErrGreaterMax     = errors.New("value is greater than maximal")
	ErrRegExpMismatch = errors.New("value does not match regular expression")
)

func (v ValidationErrors) Error() string {
	var errStr string
	for i := 0; i < len(v); i++ {
		errStr += fmt.Sprintf("Field: %s, Error: %s", v[i].Field, v[i].Err) + "\n"
	}
	return errStr
}

func Validate(v interface{}) error {
	vVal := reflect.ValueOf(v)
	if vVal.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	vType := reflect.TypeOf(v)
	var errList ValidationErrors
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		rules := getValidationRules(field.Tag)
		if len(rules) > 0 {
			for ri := 0; ri < len(rules); ri++ {
				err := validateField(rules[ri], vVal.Field(i), field.Name)
				var valErr ValidationErrors
				if err != nil {
					if errors.As(err, &valErr) {
						errList = append(errList, valErr...)
					} else {
						return err
					}
				}
			}
		}
	}

	if len(errList) > 0 {
		return errList
	}
	return nil
}

func getValidationRules(tag reflect.StructTag) (rules []string) {
	validationTag, ok := tag.Lookup("validate")
	if ok {
		if validationTag != "" {
			rules = strings.Split(validationTag, "|")
		}
	}
	return
}

func validateField(rules string, val reflect.Value, name string) error {
	rulesData := strings.Split(rules, ":")
	if len(rulesData) != 2 {
		return fmt.Errorf("field %s: %w", name, ErrInvalidRule)
	}
	if rulesData[1] == "" {
		return fmt.Errorf("field %s: %w", name, ErrInvalidRule)
	}

	switch rulesData[0] {
	case "len":
		return tryLenRule(rulesData[1], val, name)
	case "regexp":
		return tryRegexpRule(rulesData[1], val, name)
	case "in":
		return tryInRule(rulesData[1], val, name)
	case "min":
		return tryMinRule(rulesData[1], val, name)
	case "max":
		return tryMaxRule(rulesData[1], val, name)
	}

	return fmt.Errorf("field %s: %w", name, ErrInvalidRule)
}

func tryLenRule(ruleVal string, val reflect.Value, name string) error {
	switch val.Type().Kind() { //nolint:exhaustive
	case reflect.Slice:
		return validateSlice(ruleVal, val, name, tryLenRule)
	case reflect.String:
		strLen, err := strconv.Atoi(ruleVal)
		if err != nil {
			return fmt.Errorf("field %s: %w caused by %s", name, ErrInvalidRule, err)
		}
		if val.Len() != strLen {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   ErrInvalidLen,
			}}
		}
	default:
		return fmt.Errorf("field %s: %w", name, ErrWrongType)
	}

	return nil
}

func tryInRule(ruleVal string, val reflect.Value, name string) error {
	switch val.Type().Kind() { //nolint:exhaustive
	case reflect.Slice:
		return validateSlice(ruleVal, val, name, tryInRule)
	case reflect.String:
		availableVals := strings.Split(ruleVal, ",")
		if !stringInSlice(val.String(), availableVals) {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   ErrNotInSet,
			}}
		}
	case reflect.Int:
		availableVals := strings.Split(ruleVal, ",")
		availableInts := make([]int, len(availableVals))
		for i := 0; i < len(availableVals); i++ {
			value, err := strconv.Atoi(availableVals[i])
			if err != nil {
				return fmt.Errorf("field %s: %w caused by %s", name, ErrInvalidRule, err)
			}
			availableInts[i] = value
		}

		if !intInSlice(int(val.Int()), availableInts) {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   ErrNotInSet,
			}}
		}
	default:
		return fmt.Errorf("field %s: %w", name, ErrWrongType)
	}

	return nil
}

func tryMinRule(ruleVal string, val reflect.Value, name string) error {
	switch val.Type().Kind() { //nolint:exhaustive
	case reflect.Slice:
		return validateSlice(ruleVal, val, name, tryMinRule)
	case reflect.Int:
		return validateInt(ruleVal, int(val.Int()), name,
			func(fVal int, checkVal int) bool {
				return fVal < checkVal
			},
			ErrLessMin)
	default:
		return fmt.Errorf("field %s: %w", name, ErrWrongType)
	}
}

func tryMaxRule(ruleVal string, val reflect.Value, name string) error {
	switch val.Type().Kind() { //nolint:exhaustive
	case reflect.Slice:
		return validateSlice(ruleVal, val, name, tryMaxRule)
	case reflect.Int:
		return validateInt(ruleVal, int(val.Int()), name,
			func(fVal int, checkVal int) bool {
				return fVal > checkVal
			},
			ErrGreaterMax)
	default:
		return fmt.Errorf("field %s: %w", name, ErrWrongType)
	}
}

func tryRegexpRule(ruleVal string, val reflect.Value, name string) error {
	switch val.Type().Kind() { //nolint:exhaustive
	case reflect.Slice:
		return validateSlice(ruleVal, val, name, tryRegexpRule)
	case reflect.String:
		rEx, err := regexp.Compile(ruleVal)
		if err != nil {
			return fmt.Errorf("field %s: %w caused by %s", name, ErrInvalidRule, err)
		}
		if !rEx.Match([]byte(val.String())) {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   ErrRegExpMismatch,
			}}
		}
	default:
		return fmt.Errorf("field %s: %w", name, ErrWrongType)
	}

	return nil
}

func validateSlice(ruleVal string, val reflect.Value, name string, validator validationFn) error {
	var errs ValidationErrors
	for i := 0; i < val.Len(); i++ {
		err := validator(ruleVal, val.Index(i), name+"["+strconv.Itoa(i)+"]")
		var valErr ValidationErrors
		if err != nil {
			if errors.As(err, &valErr) {
				errs = append(errs, valErr...)
			} else {
				return err
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func validateInt(ruleValue string, fieldValue int, fieldName string,
	check intCheckFn, possibleErr error) error {
	checkVal, err := strconv.Atoi(ruleValue)
	if err != nil {
		return fmt.Errorf("field %s: %w caused by %s", fieldName, ErrInvalidRule, err)
	}
	if check(fieldValue, checkVal) {
		return ValidationErrors{ValidationError{
			Field: fieldName,
			Err:   possibleErr,
		}}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
