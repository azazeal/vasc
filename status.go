package vasc

// TODO(@azazeal): document the functions better.

import (
	"errors"
	"fmt"
)

type statusCode int

// The set of response status codes.
const (
	statusSyntax        = 100
	statusUnknown       = 101
	statusUnimplemented = 102
	statusTooFew        = 104
	statusTooMany       = 105
	statusParam         = 106
	statusAuth          = 107
	statusOK            = 200
	statusContinue      = 201
	statusCannot        = 300
	statusComms         = 400
	statusClose         = 500
)

var errSyntax = errors.New("vasc: syntax error")

// IsSyntax reports whether any error in err's chain occured due to a varnish
// syntax error.
func IsSyntax(err error) bool {
	return errors.Is(err, errSyntax)
}

var errUnknown = errors.New("vasc: unknown")

// IsUnknown reports whether any error in err's chain occured due to a varnish
// unknown error.
func IsUnknown(err error) bool {
	return errors.Is(err, errUnknown)
}

var errUnimplemented = errors.New("vasc: unimplemented")

// IsUnimplemented reports whether any error in err's chain occured due to the
// varnish instance not implementing the command.
func IsUnimplemented(err error) bool {
	return errors.Is(err, errUnimplemented)
}

var errTooFew = errors.New("vasc: too few args")

// IsTooFew reports whether any error in err's chain occured due to the args
// given to varnish being too few.
func IsTooFew(err error) bool {
	return errors.Is(err, errTooFew)
}

var errTooMany = errors.New("vasc: too many args")

// IsTooMany reports whether any error in err's chain occured due to the args
// given to varnish being too many.
func IsTooMany(err error) bool {
	return errors.Is(err, errTooMany)
}

var errParam = errors.New("vasc: param")

// IsParam reports whether any error in err's chain occured due to the param
// given to varnish.
func IsParam(err error) bool {
	return errors.Is(err, errParam)
}

var errHandshakeFailed = errors.New("vasc: handshake failed")

// IsHandshakeFailed reports whether any error in err's chain occured due to the
// handshake with the Varnish instance failing.
func IsHandshakeFailed(err error) bool {
	return errors.Is(err, errHandshakeFailed)
}

// ErrTruncated is returned in case the Varnish instance truncates the response.
type ErrTruncated int

func (err ErrTruncated) Error() string {
	return fmt.Sprintf("vasc: response truncated after %d bytes", err)
}

var errCannot = errors.New("vasc: cannot comply")

// IsCannot reports whether any error in err's chain occured due to the varnish
// instance being unable to comply.
func IsCannot(err error) bool {
	return errors.Is(err, errCannot)
}

var errComms = errors.New("vasc: comms error")

// IsComms reports whether any error in err's chain occured due to the varnish
// instance being unable to communicate.
func IsComms(err error) bool {
	return errors.Is(err, errComms)
}

var errClosed = errors.New("vasc: closed")

// IsClosed reports whether any error in err's chain occured due to the varnish
// instance having closed the connection.
func IsClosed(err error) bool {
	return errors.Is(err, errClosed)
}

// ErrUnexpectedStatusCode is returned in case the Varnish instace responds with
// an unexpected status code.
type ErrUnexpectedStatusCode int

func (err ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("vasc: unexpected status code %d", err)
}

// ErrInvalidHeader is returned in case the Varnish instace responds with an
// invalid header.
type ErrInvalidHeader string

func (err ErrInvalidHeader) Error() string {
	return fmt.Sprintf("vasc: invalid header %q", string(err))
}

var errInvalidJSONResponse = errors.New("vasc: invalid JSON response")

// IsInvalidJSONResponse reports whether any error in err's chain occured due to
// the varnish instance responding with invalid JSON.
func IsInvalidJSONResponse(err error) bool {
	return errors.Is(err, errInvalidJSONResponse)
}

var errAuth = errors.New("vasc: auth")

func codeToError(code, size int) (err error) {
	switch code {
	default:
		err = ErrUnexpectedStatusCode(code)
	case statusOK:
		break
	case statusSyntax:
		err = errSyntax
	case statusUnknown:
		err = errUnknown
	case statusUnimplemented:
		err = errUnimplemented
	case statusTooFew:
		err = errTooFew
	case statusTooMany:
		err = errTooMany
	case statusParam:
		err = errParam
	case statusAuth:
		err = errAuth
	case statusContinue:
		err = ErrTruncated(size)
	case statusCannot:
		err = errCannot
	case statusComms:
		err = errComms
	case statusClose:
		err = errClosed
	}

	return
}
