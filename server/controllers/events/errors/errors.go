package errors

// RequestValidationError is it's own unique type
// in order to allow us to type assert and return custom http
// error codes accordingly
type RequestValidationError struct {
	Err error
}

func (e *RequestValidationError) Error() string {
	return e.Err.Error()
}

// WebhookParsingError is it's own unique type
// in order to allow us to type assert and return custom http
// error codes accordingly
type WebhookParsingError struct {
	Err error
}

func (e *WebhookParsingError) Error() string {
	return e.Err.Error()
}

// EventParsingError is it's own unique type
// in order to allow us to type assert and return custom http
// error codes accordingly
type EventParsingError struct {
	Err error
}

func (e *EventParsingError) Error() string {
	return e.Err.Error()
}

// UnsupportedEventTypeError is it's own unique type
// in order to allow us to type assert and return custom http
// error codes accordingly
type UnsupportedEventTypeError struct {
	Msg string
}

func (r *UnsupportedEventTypeError) Error() string {
	return r.Msg
}
