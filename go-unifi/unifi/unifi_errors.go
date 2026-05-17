package unifi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrNotFound = errors.New("not found")

type Meta struct {
	RC      string `json:"rc"`
	Message string `json:"msg"`
}

func (m *Meta) error() error {
	if m.RC != "ok" {
		return &ServerError{
			ErrorCode: m.RC,
			Message:   m.Message,
		}
	}

	return nil
}

type DefaultResponseErrorHandler struct{}

type apiV2ResponseError struct {
	Code      string                    `json:"code"`
	ErrorCode int                       `json:"errorCode"`
	Message   string                    `json:"message"`
	Details   apiV2ResponseErrorDetails `json:"details"`
}

type apiV2ResponseErrorDetails struct {
	// probably there are more fields, but I didn't get any response with more fields
	InvalidFields []string `json:"invalid_fields"`
}

type apiV1ResponseError struct {
	Meta Meta                     `json:"Meta"`
	Data []apiV1ResponseErrorData `json:"data"`
}

type apiV1ResponseErrorData struct {
	Meta            Meta                  `json:"Meta"`
	ValidationError ServerValidationError `json:"validationError"`
	RC              string                `json:"rc"`
	Message         string                `json:"msg"`
}

type apiResponseError struct {
	apiV1ResponseError
	apiV2ResponseError
}

type ServerValidationError struct {
	Field   string `json:"field"`
	Pattern string `json:"pattern"`
}

type ServerErrorDetails struct {
	Message         string
	ValidationError ServerValidationError
}

type ServerError struct {
	StatusCode    int
	RequestMethod string
	RequestURL    string
	Message       string
	ErrorCode     string
	Details       []ServerErrorDetails
}

func (s *ServerError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Server error (%d) for %s %s: %s", s.StatusCode, s.RequestMethod, s.RequestURL, s.Message))
	for _, detail := range s.Details {
		b.WriteString("\n")
		if detail.Message != "" {
			b.WriteString(detail.Message + ": ")
		}
		if detail.ValidationError.Field != "" && detail.ValidationError.Pattern != "" {
			b.WriteString(fmt.Sprintf("field '%s' should match '%s'", detail.ValidationError.Field, detail.ValidationError.Pattern))
		} else if detail.ValidationError.Field != "" {
			b.WriteString(fmt.Sprintf("field '%s' is invalid", detail.ValidationError.Field))
		} else if detail.ValidationError.Pattern != "" {
			b.WriteString(fmt.Sprintf("field should match '%s'", detail.ValidationError.Pattern))
		}
	}
	return b.String()
}

func parseApiV2Error(err apiV2ResponseError, serverError *ServerError) {
	serverError.Message = err.Message
	serverError.ErrorCode = err.Code
	for _, field := range err.Details.InvalidFields {
		details := ServerErrorDetails{}
		details.ValidationError.Field = field
		serverError.Details = append(serverError.Details, details)
	}
}

func parseApiV1Error(err apiV1ResponseError, serverError *ServerError) {
	for _, d := range err.Data {
		if d.Meta.RC == "error" || d.RC == "error" {
			details := ServerErrorDetails{}
			details.Message = d.Message
			if details.Message == "" {
				details.Message = d.Meta.Message
			}
			if d.ValidationError.Field != "" || d.ValidationError.Pattern != "" {
				details.ValidationError = d.ValidationError
			}
			serverError.Details = append(serverError.Details, details)
		}
	}
	if serverError.Message == "" {
		serverError.Message = err.Meta.Message
	}
	serverError.ErrorCode = err.Meta.RC
}

func (d *DefaultResponseErrorHandler) HandleError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	var errBody apiResponseError
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		return err
	}
	serverError := ServerError{
		StatusCode:    resp.StatusCode,
		RequestMethod: resp.Request.Method,
		RequestURL:    resp.Request.URL.String(),
	}
	if errBody.apiV2ResponseError.Code != "" || errBody.apiV2ResponseError.Message != "" {
		parseApiV2Error(errBody.apiV2ResponseError, &serverError)
	} else {
		parseApiV1Error(errBody.apiV1ResponseError, &serverError)
	}

	return &serverError
}
