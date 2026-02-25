package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// SchemaValidator validates POST/PUT request bodies against JSON schemas.
type SchemaValidator struct {
	schemas map[string]*jsonschema.Schema
}

func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		schemas: make(map[string]*jsonschema.Schema),
	}
}

// RegisterSchema compiles and stores a JSON schema for a route pattern.
func (sv *SchemaValidator) RegisterSchema(pattern string, schemaJSON string) error {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(pattern, strings.NewReader(schemaJSON)); err != nil {
		return err
	}
	schema, err := compiler.Compile(pattern)
	if err != nil {
		return err
	}
	sv.schemas[pattern] = schema
	return nil
}

// Middleware returns the validation middleware for a specific route pattern.
func (sv *SchemaValidator) Middleware(pattern string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST and PUT
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				next.ServeHTTP(w, r)
				return
			}

			schema, ok := sv.schemas[pattern]
			if !ok {
				// No schema registered, pass through
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				model.WriteError(w, model.ErrValidation, "Failed to read request body", nil)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))

			if err := schema.Validate(bytes.NewReader(body)); err != nil {
				model.WriteError(w, model.ErrValidation, "Request body validation failed", map[string]string{
					"error": err.Error(),
				})
				return
			}

			// Reset body for downstream handlers
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
