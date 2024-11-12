package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/navarroslab/whisperer/internal/validator"
)

// Define an envelope type.
type envelope map[string]any

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder and call the DisallowUnknownFields() method on it
	// before decoding. This means that if the JSON form the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return
	// an error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err == nil { // if err IS nil
		return nil
	}

	// if there is an error during decoding, start the triage...
	var (
		syntaxError           *json.SyntaxError
		unmarshallTypeError   *json.UnmarshalTypeError
		invalidUnmarshalError *json.InvalidUnmarshalError
	)

	switch {
	// Use the errors.As() function to check whether the error has the type *json.SyntaxError
	// If so, return an plain-english message which includes location of the problem
	case errors.As(err, &syntaxError):
		return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
	// In some circumstances Decode() can return an io.ErrUnexpectedEOF error
	// for syntax errors in JSON. We check for this using errors.Is()
	case errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("body contains badly-formed JSON")
	// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the JSON value is the wrong type
	// for the target destination. If the error relates to a specific field, then we include that in our error message
	case errors.As(err, &unmarshallTypeError):
		if unmarshallTypeError.Field != "" {
			return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshallTypeError.Field)
		}

		return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshallTypeError.Offset)
	// An io.EOF error will be returned by Decode() if the request body is empty.
	// We check this with errors.Is() and return a plain-english error message.
	case errors.Is(err, io.EOF):
		return errors.New("body must not be empty")
	// A json.InvalidUnmarshalError error will be returned if we pass something
	// that is not a non-nil pointer to Decode(). We catch this and panic rather
	// than returning an error to our handler. At the end of this chapter we'll talk about
	// panicking versus returning errors, and discuss why it's an appropriate thing to do
	// in this specific situation.
	case errors.As(err, &invalidUnmarshalError):
		panic(err)
	// For anything else, return the error message as-is.
	default:
		return err
	}
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	// At this point, let's add the headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readIDParam(r *http.Request) (int, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil {
		return 0, errors.New("invalid id parameter")
	}

	return int(id), nil
}

// The readString() helper returns a string value from the query string, or the provided default value.
func (app *application) readString(qs url.Values, key, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

// The readCSV helper reads a string value from the query string and then splits it into a slice of the comma character.
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// The readInt() helper reads a string value from the query string and converts it to an integer before returning.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// The background() helper accepts an arbitrary function as parameter.
func (app *application) background(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}
