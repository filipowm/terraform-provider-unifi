package unifi //nolint: testpackage

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerError_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    ServerError
		expected string
	}{
		{
			name: "Error with message and validation error",
			input: ServerError{
				StatusCode:    400,
				RequestMethod: "GET",
				RequestURL:    "http://example.com",
				Message:       "Bad Request",
				Details: []ServerErrorDetails{
					{
						Message: "Invalid input",
						ValidationError: ServerValidationError{
							Field:   "username",
							Pattern: "^[a-zA-Z0-9]+$",
						},
					},
				},
			},
			expected: "Server error (400) for GET http://example.com: Bad Request\nInvalid input: field 'username' should match '^[a-zA-Z0-9]+$'",
		},
		{
			name: "Error with message only",
			input: ServerError{
				StatusCode:    404,
				RequestMethod: "POST",
				RequestURL:    "http://example.com",
				Message:       "Not Found",
			},
			expected: "Server error (404) for POST http://example.com: Not Found",
		},
		{
			name: "Error with multiple validation errors",
			input: ServerError{
				StatusCode:    422,
				RequestMethod: "PUT",
				RequestURL:    "http://example.com",
				Message:       "Unprocessable Entity",
				Details: []ServerErrorDetails{
					{
						Message: "Invalid username",
						ValidationError: ServerValidationError{
							Field:   "username",
							Pattern: "^[a-zA-Z0-9]+$",
						},
					},
					{
						Message: "Invalid email",
						ValidationError: ServerValidationError{
							Field:   "email",
							Pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
						},
					},
				},
			},
			expected: "Server error (422) for PUT http://example.com: Unprocessable Entity\nInvalid username: field 'username' should match '^[a-zA-Z0-9]+$'\nInvalid email: field 'email' should match '^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$'",
		},
		{
			name: "Error with no details",
			input: ServerError{
				StatusCode:    500,
				RequestMethod: "DELETE",
				RequestURL:    "http://example.com",
				Message:       "Internal Server Error",
			},
			expected: "Server error (500) for DELETE http://example.com: Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			a.Equal(tt.expected, tt.input.Error())
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		responseBody  string
		statusCode    int
		expectedError *ServerError
	}{
		{
			name: "V2 error with invalid fields",
			responseBody: `{
				"code": "error_code",
				"message": "error message",
				"details": {
					"invalid_fields": ["field1", "field2"]
				}
			}`,
			statusCode: 400,
			expectedError: &ServerError{
				StatusCode:    400,
				RequestMethod: "GET",
				RequestURL:    "http://example.com",
				Message:       "error message",
				ErrorCode:     "error_code",
				Details: []ServerErrorDetails{
					{ValidationError: ServerValidationError{Field: "field1"}},
					{ValidationError: ServerValidationError{Field: "field2"}},
				},
			},
		},
		{
			name: "V1 error with validation error",
			responseBody: `{
				"Meta": {"rc": "error", "msg": "meta error message"},
				"data": [{
					"Meta": {"rc": "error", "msg": "data meta error message"},
					"validationError": {"field": "field1", "pattern": "pattern1"},
					"rc": "error",
					"msg": "data error message"
				}]
			}`,
			statusCode: 400,
			expectedError: &ServerError{
				StatusCode:    400,
				RequestMethod: "GET",
				RequestURL:    "http://example.com",
				Message:       "meta error message",
				ErrorCode:     "error",
				Details: []ServerErrorDetails{
					{
						Message: "data error message",
						ValidationError: ServerValidationError{
							Field:   "field1",
							Pattern: "pattern1",
						},
					},
				},
			},
		},
		{
			name:          "No error",
			responseBody:  `{"Meta": {"rc": "ok", "msg": "success"}}`,
			statusCode:    200,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := assert.New(t)
			handler := &DefaultResponseErrorHandler{}

			// Create a new HTTP response recorder
			recorder := httptest.NewRecorder()
			recorder.WriteHeader(tt.statusCode)
			recorder.Body = bytes.NewBufferString(tt.responseBody)

			// Create a new HTTP request
			req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
			recorder.Result().Request = req

			// Call the HandleError function
			err := handler.HandleError(recorder.Result())

			if tt.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				a.EqualValues(tt.expectedError, err)
			}
		})
	}
}
