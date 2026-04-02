package errors

import (
	"errors"
	"fmt"
)

// Local XML client errors for unsupported operations

var (
	// ErrUnsupportedOperation is returned when an operation is not supported in local XML mode
	ErrUnsupportedOperation = errors.New("operation not supported in local XML mode")

	// ErrWriteNotSupported is returned when attempting write operations in local XML mode
	ErrWriteNotSupported = errors.New("write operations not supported in local XML mode")

	// ErrJobsNotSupported is returned when attempting job operations in local XML mode
	ErrJobsNotSupported = errors.New("job operations not supported in local XML mode")
)

// ErrInvalidXpath indicates XPath syntax validation failure.
// The Cause field wraps the underlying parse error from xmlquery.
//
// Example usage:
//
//	parseErr := errors.New("unexpected EOF")
//	err := NewErrInvalidXpath("/config/entry[@name=''", parseErr)
//	fmt.Println(err.Error())
//	// Output: invalid XPath syntax '/config/entry[@name=''': unexpected EOF
type ErrInvalidXpath struct {
	XPath string // The invalid XPath expression
	Cause error  // Underlying parse error from xmlquery
}

// Error implements the error interface.
func (e *ErrInvalidXpath) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("invalid XPath syntax '%s': %v", e.XPath, e.Cause)
	}
	return fmt.Sprintf("invalid XPath syntax '%s'", e.XPath)
}

// Unwrap returns the underlying cause error, supporting error unwrapping.
func (e *ErrInvalidXpath) Unwrap() error {
	return e.Cause
}

// ErrObjectNotFound indicates the target element was not found at the specified XPath.
// This matches PAN-OS API behavior when an XPath query returns no results.
//
// Example usage:
//
//	err := NewErrObjectNotFound("/config/devices/entry[@name='localhost']/address/entry[@name='missing']")
//	fmt.Println(err.Error())
//	// Output: object not found at XPath '/config/devices/entry[@name='localhost']/address/entry[@name='missing']'
type ErrObjectNotFound struct {
	XPath string // XPath where object was expected but not found
}

// Error implements the error interface.
func (e *ErrObjectNotFound) Error() string {
	return fmt.Sprintf("object not found at XPath '%s'", e.XPath)
}

// ErrOperationFailed indicates a specific operation within a MultiConfig batch failed.
// The OperationIndex identifies which operation (zero-based) caused the failure,
// and Cause contains the underlying error.
//
// Example usage:
//
//	notFoundErr := NewErrObjectNotFound("/config/.../entry[@name='missing']")
//	err := NewErrOperationFailed(3, notFoundErr)
//	fmt.Println(err.Error())
//	// Output: operation 3 failed: object not found at XPath '/config/.../entry[@name='missing']'
type ErrOperationFailed struct {
	OperationIndex int   // Zero-based index of failed operation in MultiConfig batch
	Cause          error // Underlying error (ErrInvalidXpath, ErrObjectNotFound, etc.)
}

// Error implements the error interface.
func (e *ErrOperationFailed) Error() string {
	return fmt.Sprintf("operation %d failed: %v", e.OperationIndex, e.Cause)
}

// Unwrap returns the underlying cause error, supporting error unwrapping.
func (e *ErrOperationFailed) Unwrap() error {
	return e.Cause
}

// ErrRenameConflict indicates a rename operation failed because the target name
// already exists in the parent container.
//
// Example usage:
//
//	err := NewErrRenameConflict(
//	    "/config/devices/entry[@name='localhost']/address/entry[@name='old-name']",
//	    "old-name",
//	    "new-name",
//	)
//	fmt.Println(err.Error())
//	// Output: rename conflict: target name 'new-name' already exists (source: 'old-name' at '/config/.../entry[@name='old-name']')
type ErrRenameConflict struct {
	XPath      string // XPath of source element being renamed
	SourceName string // Current name of the element
	TargetName string // Desired new name (already exists)
}

// Error implements the error interface.
func (e *ErrRenameConflict) Error() string {
	return fmt.Sprintf("rename conflict: target name '%s' already exists (source: '%s' at '%s')",
		e.TargetName, e.SourceName, e.XPath)
}

// NewErrInvalidXpath creates an XPath validation error.
// Use this when XPath syntax is malformed or cannot be parsed.
//
// Example:
//
//	err := NewErrInvalidXpath("/bad[[@xpath", parseErr)
func NewErrInvalidXpath(xpath string, cause error) *ErrInvalidXpath {
	return &ErrInvalidXpath{
		XPath: xpath,
		Cause: cause,
	}
}

// NewErrObjectNotFound creates an object not found error.
// Use this when an XPath query returns zero results.
//
// Example:
//
//	err := NewErrObjectNotFound("/config/.../entry[@name='missing']")
func NewErrObjectNotFound(xpath string) *ErrObjectNotFound {
	return &ErrObjectNotFound{
		XPath: xpath,
	}
}

// NewErrOperationFailed creates a MultiConfig operation failure error.
// The index is zero-based and indicates which operation in the batch failed.
//
// Example:
//
//	notFoundErr := NewErrObjectNotFound("/config/.../entry[@name='missing']")
//	err := NewErrOperationFailed(3, notFoundErr)
func NewErrOperationFailed(index int, cause error) *ErrOperationFailed {
	return &ErrOperationFailed{
		OperationIndex: index,
		Cause:          cause,
	}
}

// NewErrRenameConflict creates a rename conflict error.
// Use this when a rename operation fails because the target name already exists.
//
// Example:
//
//	err := NewErrRenameConflict("/config/.../entry[@name='old']", "old", "new")
func NewErrRenameConflict(xpath, sourceName, targetName string) *ErrRenameConflict {
	return &ErrRenameConflict{
		XPath:      xpath,
		SourceName: sourceName,
		TargetName: targetName,
	}
}
