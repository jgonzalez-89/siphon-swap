package parser

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/gorilla/schema"
)

var (
	decoder         = schema.NewDecoder()
	schemaValidator = validator.New()
)

func init() {
	decoder.IgnoreUnknownKeys(true)
}

func Unmarshal[T any](r *http.Request, dest *T) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := decoder.Decode(dest, r.Form); err != nil {
		return err
	}

	if err := schemaValidator.Struct(dest); err != nil {
		return err
	}

	return nil
}
