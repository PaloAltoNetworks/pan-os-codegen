package errors

import "fmt"

type SchemaError struct {
	Message string
	Err     error
}

func NewSchemaError(msg string) *SchemaError {
	return &SchemaError{Message: msg}
}

func (o *SchemaError) Error() string {
	return o.Message
}

func (o *SchemaError) Unwrap() error {
	return o.Err
}

type MarshallerError struct {
	Err     error
	Message string
}

func NewMarshallerError(msg string) *MarshallerError {
	return &MarshallerError{Message: msg}
}

func NewWrappedMarshallerError(err error) *MarshallerError {
	return &MarshallerError{
		Err: err,
	}
}

func (o *MarshallerError) Error() string {
	if o.Err != nil {
		return fmt.Sprintf("error while marshalling specs: %s", o.Err)
	}
	return o.Message
}

func (o *MarshallerError) Unwrap() error {
	return o.Err
}
