package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FibrinLab/open-nucleus/internal/handler"
	"github.com/FibrinLab/open-nucleus/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStubHandler_Returns501(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.StubHandler()(rr, req)

	assert.Equal(t, http.StatusNotImplemented, rr.Code)

	var env model.Envelope
	err := json.NewDecoder(rr.Body).Decode(&env)
	require.NoError(t, err)
	assert.Equal(t, "error", env.Status)
	assert.Equal(t, model.ErrNotImplemented, env.Error.Code)
}
