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

var (
	valueError = errors.New("input is not a struct")
	ruleError  = errors.New("invalid validation rule")
	typeError  = errors.New("rule unavailable for this field type")

	invalidLenError       = errors.New("string len is not valid")
	notInSetError         = errors.New("value is not in validated set")
	lessMinError          = errors.New("value is less than minimal")
	greaterMaxError       = errors.New("value is greater than maximal")
	regexpValidationError = errors.New("value does not match regular expression")
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
		return valueError
	}
	vType := reflect.TypeOf(v)
	var errList ValidationErrors
	for i := 0; i < vType.NumField(); i++ {
		field := vType.Field(i)
		validationTag, ok := field.Tag.Lookup("validate")
		if ok {
			if validationTag != "" {
				// разобьем правила валидации - на случай объединенных через "|"
				rules := strings.Split(validationTag, "|")
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
			continue
		}
		continue
	}

	if len(errList) > 0 {
		return errList
	}
	return nil
}

func validateField(rules string, val reflect.Value, name string) error {
	rulesData := strings.Split(rules, ":")
	if len(rulesData) != 2 {
		return fmt.Errorf("field %s: %w", name, ruleError)
	}
	if rulesData[1] == "" {
		return fmt.Errorf("field %s: %w", name, ruleError)
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

	return fmt.Errorf("field %s: %w", name, ruleError)
}

func tryLenRule(ruleVal string, val reflect.Value, name string) error {
	var errs ValidationErrors
	if val.Type().Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			err := tryLenRule(ruleVal, val.Index(i), name+"["+strconv.Itoa(i)+"]")
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
	} else if val.Type().Kind() == reflect.String {
		strLen, err := strconv.Atoi(ruleVal)
		if err != nil {
			return fmt.Errorf("field %s: %w caused by %s", name, ruleError, err)
		}
		if val.Len() != strLen {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   invalidLenError,
			}}
		}
	} else {
		return fmt.Errorf("field %s: %w", name, typeError)
	}

	return nil
}

func tryInRule(ruleVal string, val reflect.Value, name string) error {
	var errs ValidationErrors
	if val.Type().Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			err := tryInRule(ruleVal, val.Index(i), name+"["+strconv.Itoa(i)+"]")
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
	} else if val.Type().Kind() == reflect.String {
		availableVals := strings.Split(ruleVal, ",")
		if !stringInSlice(val.String(), availableVals) {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   notInSetError,
			}}
		}
	} else if val.Type().Kind() == reflect.Int {
		availableVals := strings.Split(ruleVal, ",")
		availableInts := make([]int, len(availableVals))
		for i := 0; i < len(availableVals); i++ {
			value, err := strconv.Atoi(availableVals[i])
			if err != nil {
				return fmt.Errorf("field %s: %w caused by %s", name, ruleError, err)
			}
			availableInts[i] = value
		}

		if !intInSlice(int(val.Int()), availableInts) {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   notInSetError,
			}}
		}
	} else {
		return fmt.Errorf("field %s: %w", name, typeError)
	}

	return nil
}

func tryMinRule(ruleVal string, val reflect.Value, name string) error {
	var errs ValidationErrors
	if val.Type().Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			err := tryMinRule(ruleVal, val.Index(i), name+"["+strconv.Itoa(i)+"]")
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
	} else if val.Type().Kind() == reflect.Int {
		minVal, err := strconv.Atoi(ruleVal)
		if err != nil {
			return fmt.Errorf("field %s: %w caused by %s", name, ruleError, err)
		}
		if int(val.Int()) < minVal {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   lessMinError,
			}}
		}
	} else {
		return fmt.Errorf("field %s: %w", name, typeError)
	}

	return nil
}

func tryMaxRule(ruleVal string, val reflect.Value, name string) error {
	var errs ValidationErrors
	if val.Type().Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			err := tryMaxRule(ruleVal, val.Index(i), name+"["+strconv.Itoa(i)+"]")
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
	} else if val.Type().Kind() == reflect.Int {
		maxVal, err := strconv.Atoi(ruleVal)
		if err != nil {
			return fmt.Errorf("field %s: %w caused by %s", name, ruleError, err)
		}
		if int(val.Int()) > maxVal {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   greaterMaxError,
			}}
		}
	} else {
		return fmt.Errorf("field %s: %w", name, typeError)
	}

	return nil
}

func tryRegexpRule(ruleVal string, val reflect.Value, name string) error {
	var errs ValidationErrors
	if val.Type().Kind() == reflect.Slice {
		for i := 0; i < val.Len(); i++ {
			err := tryRegexpRule(ruleVal, val.Index(i), name+"["+strconv.Itoa(i)+"]")
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
	} else if val.Type().Kind() == reflect.String {
		rEx, err := regexp.Compile(ruleVal)
		if err != nil {
			return fmt.Errorf("field %s: %w caused by %s", name, ruleError, err)
		}
		if !rEx.Match([]byte(val.String())) {
			return ValidationErrors{ValidationError{
				Field: name,
				Err:   regexpValidationError,
			}}
		}
	} else {
		return fmt.Errorf("field %s: %w", name, typeError)
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
