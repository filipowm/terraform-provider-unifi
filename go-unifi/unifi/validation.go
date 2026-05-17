package unifi

import (
	"errors"
	"fmt"
	"regexp"
	"sync"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	vd "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// ValidationError is a custom error type for validation errors.
type ValidationError struct {
	Root     error
	Messages map[string]string
}

// Error returns the error message with combined all validation error messages.
func (v *ValidationError) Error() string {
	err := "validation failed: \n"
	for field, message := range v.Messages {
		err += fmt.Sprintf("%s: %s\n", field, message)
	}
	return err
}

// Validator is the interface for the validator. Use it to validate structs. You can register structure-level validations
// with RegisterStructValidation.
type Validator interface {
	// Validate validates the given struct and returns an error if the struct is not valid.
	Validate(i interface{}) error
	// RegisterStructValidation registers a structure-level validation function for a given struct type.
	RegisterStructValidation(fn vd.StructLevelFunc, i interface{})
	// RegisterTranslation registers a custom translation for a given tag.
	RegisterTranslation(tag string, registerFn vd.RegisterTranslationsFunc, translationFn vd.TranslationFunc) (err error)
	// RegisterCustomValidator registers a custom validator function with own tag and error message.
	RegisterCustomValidator(cv CustomValidator) error
}

type validator struct {
	validate *vd.Validate
	trans    ut.Translator
}

func (v *validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		var errs vd.ValidationErrors
		errors.As(err, &errs)
		messages := errs.Translate(v.trans)

		return &ValidationError{Root: err, Messages: messages}
	}
	return nil
}

func (v *validator) RegisterStructValidation(f vd.StructLevelFunc, s interface{}) {
	v.validate.RegisterStructValidation(f, s)
}

func (v *validator) RegisterTranslation(tag string, registerFn vd.RegisterTranslationsFunc, translationFn vd.TranslationFunc) error {
	return v.validate.RegisterTranslation(tag, v.trans, registerFn, translationFn)
}

func (v *validator) RegisterCustomValidator(cv CustomValidator) error {
	var err error
	if err = v.validate.RegisterValidation(cv.tag, cv.fn, false); err != nil {
		return fmt.Errorf("failed to register custom validation '%s': %w", cv.tag, err)
	}
	err = v.RegisterTranslation(cv.tag, func(ut ut.Translator) error {
		return ut.Add(cv.tag, cv.messageText, true)
	}, func(ut ut.Translator, fe vd.FieldError) string {
		t, _ := ut.T(cv.tag, append([]string{fe.Field()}, cv.params...)...)
		return t
	})
	if err != nil {
		return fmt.Errorf("failed to register custom validation '%s' translation: %w", cv.tag, err)
	}
	return nil
}

func newValidator() (*validator, error) {
	validate := vd.New(vd.WithRequiredStructEnabled())
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)
	trans, _ := uni.GetTranslator(enLocale.Locale())
	err := en_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return nil, err
	}

	v := &validator{
		validate: validate,
		trans:    trans,
	}

	for _, customValidator := range customValidators {
		if err = v.RegisterCustomValidator(customValidator); err != nil {
			return nil, err
		}
	}
	return v, nil
}

type CustomValidator struct {
	tag         string
	fn          vd.Func
	messageText string
	params      []string
}

func NewCustomRegexValidator(tag string, regex string) CustomValidator {
	cv := &CustomValidator{
		tag:         tag,
		messageText: regexValidatorMessage,
		params:      []string{regex},
	}
	crv := CustomRegexValidator{
		CustomValidator: cv,
		regex:           lazyRegexCompile(regex),
	}
	crv.fn = func(fl vd.FieldLevel) bool {
		return crv.regex().MatchString(fl.Field().String())
	}
	return *crv.CustomValidator
}

type CustomRegexValidator struct {
	*CustomValidator
	regex func() *regexp.Regexp
}

var customValidators = []CustomValidator{
	NewCustomRegexValidator("w_regex", wRegexString),
	NewCustomRegexValidator("numeric_nonzero", `^[1-9][0-9]*$`),
}

func lazyRegexCompile(str string) func() *regexp.Regexp {
	var regex *regexp.Regexp
	var once sync.Once
	return func() *regexp.Regexp {
		once.Do(func() {
			regex = regexp.MustCompile(str)
		})
		return regex
	}
}

const (
	regexValidatorMessage = "{0} must comply with the regular expression pattern '{1}'"
	wRegexString          = `^[\w]+$`
)
