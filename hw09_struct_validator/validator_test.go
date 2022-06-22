package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Dummy struct {
		Num  []int `validate:"min:0"`
		Num1 int   `validate:"min"`
	}

	Wrong struct {
		Num int `validate:"len:1"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:    "J3QQ4-H7H2V-2HCH-M3HK8-6M8VW",
				Name:  "Billy",
				Age:   99,
				Email: "nope",
				Role:  "teapot",
				Phones: []string{
					"1234",
					"88005553535",
				},
				meta: nil,
			},
			expectedErr: ValidationErrors{
				{
					Field: "ID",
					Err:   invalidLenError,
				},
				{
					Field: "Age",
					Err:   greaterMaxError,
				},
				{
					Field: "Email",
					Err:   regexpValidationError,
				},
				{
					Field: "Role",
					Err:   notInSetError,
				},
				{
					Field: "Phones[0]",
					Err:   invalidLenError,
				},
			},
		},
		{
			in:          App{Version: "1.2.3"},
			expectedErr: nil,
		},
		{
			in: App{Version: "1.2.3 build 130622"},
			expectedErr: ValidationErrors{
				{
					Field: "Version",
					Err:   invalidLenError,
				},
			},
		},
		{
			in: Token{
				Header:    []byte{'i', 'd', 'd', 'q', 'd'},
				Payload:   nil,
				Signature: nil,
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 418,
				Body: "I'm a teapot",
			},
			expectedErr: ValidationErrors{
				{
					Field: "Code",
					Err:   notInSetError,
				},
			},
		},
		{
			in: Dummy{
				Num:  []int{1, 2, 3, 4, 5},
				Num1: 2,
			},
			expectedErr: ruleError,
		},
		{
			in:          Wrong{Num: 1},
			expectedErr: typeError,
		},
		{
			in:          "validate this",
			expectedErr: valueError,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			result := Validate(tt.in)
			if tt.expectedErr != nil {
				require.True(t, errors.Is(result, tt.expectedErr))
			} else {
				require.Nil(t, result)
			}
		})
	}
}
