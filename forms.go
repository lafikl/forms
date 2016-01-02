package forms

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// validator "github.com/asaskevich/govalidator"

// FieldError is a field validation error
type FieldError struct {
	Name  string
	Error error
}

// NewFieldError creates a new instance of FieldError
func NewFieldError(name string, err error) *FieldError {
	return &FieldError{
		name,
		err,
	}
}

// Input is for creating generic input fields
// <input class="" id="" type="" name="" value="" placeholder="">
type Input struct {
	// Defaults to "text"
	typ      string
	name     string
	value    string
	classes  []string
	min      int
	max      int
	required bool
	// a map of attributes for the field
	attrs      map[string]string
	validators []func(v *Input) error
}

// Field is the common interface between all the fields
type Field interface {
	AddValidator(v func(*Input) error) *Input
	Validate() *FieldError
	Name() string
	SetName(name string)
	Value() string
	SetValue(v string)
	String() string
}

// NewInput creates a new text input field
func NewInput() *Input {
	i := new(Input)
	i.classes = []string{}
	i.attrs = map[string]string{}
	i.typ = "text"
	i.validators = []func(v *Input) error{}
	return i
}

// Name returns the name of the field
func (i *Input) Name() string {
	return i.name
}

// SetName sets the name of the field
func (i *Input) SetName(n string) {
	i.name = n
}

// AddClass appends the given class to the slice
func (i *Input) AddClass(class string) *Input {
	i.classes = append(i.classes, class)
	return i
}

// AddAttr appends the given attribute to the slice
func (i *Input) AddAttr(key, value string) *Input {
	i.attrs[key] = value
	return i
}

// AddValidator appends the given attribute to the slice
func (i *Input) AddValidator(v func(*Input) error) *Input {
	i.validators = append(i.validators, v)
	return i
}

func (i *Input) String() string {
	frmt := "<input type='%s' name='%s' value='%s' class='%s' %s>"
	input := fmt.Sprintf(frmt, i.typ, i.name, i.value, strings.Join(i.classes, " "), i.FmtAttrs())
	return input
}

// Value sets the given value to the field
func (i *Input) Value() string {
	return i.value
}

// SetValue sets the given value to the field
func (i *Input) SetValue(v string) {
	i.value = v
}

// FmtAttrs formats the attributes to `key=value` format
func (i *Input) FmtAttrs() string {
	attrs := []string{}
	frmt := "%s='%s'"
	for k, v := range i.attrs {
		attrs = append(attrs, fmt.Sprintf(frmt, k, v))
	}
	return strings.Join(attrs, " ")
}

// Validate loops through the validation funcs and stores the errors
func (i *Input) Validate() *FieldError {
	var fe *FieldError
	for _, f := range i.validators {
		err := f(i)
		if err != nil {
			fe = NewFieldError(i.Name(), err)
			break
		}
	}
	return fe
}

// TextInput is for creating inputs of type text
type TextInput struct {
	Input
}

// IntegerInput is for creating inputs of type text
type IntegerInput struct {
	Input
}

var integerValidators = []func(*Input) error{
	isInteger,
	integerBound,
}

func integerBound(i *Input) error {
	// the validity of the integer is checked before by isInteger
	v, _ := strconv.ParseInt(i.value, 10, 32)
	frmt := "%s is %s than "
	if int(v) < i.min {
		return fmt.Errorf(frmt, i.name, "less", i.min)
	}
	if int(v) > i.max {
		return fmt.Errorf(frmt, i.name, "more", i.max)
	}
	return nil
}

func isInteger(i *Input) error {
	if _, err := strconv.ParseInt(i.value, 10, 32); err != nil {
		return errors.New("Not a valid integer.")
	}
	return nil
}

// NewIntegerInput creates a new integer type input
func NewIntegerInput() *IntegerInput {
	input := new(IntegerInput)
	for _, v := range integerValidators {
		input.validators = append(input.validators, v)
	}
	return input
}

// Form creates a form out of the given fields
type Form struct {
	action string
	method string
	fields map[string]Field
}

// NewForm creates a new form
func NewForm() *Form {
	form := new(Form)
	form.method = "GET"
	form.action = ""
	form.fields = map[string]Field{}
	return form
}

// AddInput adds the given input to the form
func (f *Form) AddInput(field Field) *Form {
	f.fields[field.Name()] = field
	return f
}

// SetAction changes the action attribute to the given action
func (f *Form) SetAction(action string) *Form {
	f.action = action
	return f
}

// Action returns the action attribute
func (f *Form) Action() string {
	return f.action
}

// SetMethod changes the method attribute to the given method
func (f *Form) SetMethod(method string) *Form {
	f.method = method
	return f
}

// Method returns the method attribute
func (f *Form) Method() string {
	return f.method
}

// HTML returns a safe HTML code of the form
func (f *Form) HTML() template.HTML {
	form := fmt.Sprintf("<form action='%s' method='%s'>", f.action, f.method)
	for _, field := range f.fields {
		form += field.String()
	}
	form += "</form>"
	html := template.HTML(form)
	return html
}

// Load the submitted form
func (f *Form) Load(r *http.Request) *Form {
	if r == nil {
		return f
	}
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return f
	}
	for name, field := range f.fields {
		if f.method == "GET" {
			field.SetValue(r.FormValue(name))
			continue
		}
		field.SetValue(r.PostFormValue(name))
	}
	return f
}

// FormErrors is a slice of errors for every field
type FormErrors map[string]*FieldError

// Validate the submitted form
func (f *Form) Validate() FormErrors {
	errs := FormErrors{}
	for k, field := range f.fields {
		if err := field.Validate(); err != nil {
			errs[k] = err
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Values returns a map of the values
func (f *Form) Values() map[string]string {
	values := map[string]string{}
	for k, field := range f.fields {
		values[k] = field.Value()
	}
	return values
}
