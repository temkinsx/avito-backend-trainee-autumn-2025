package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-backend-trainee-autumn-2025/internal/api/dto"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRecorderWithRequest(t *testing.T, method, target string, body any) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}

	req, err := http.NewRequest(method, target, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	return w, c
}

func decodeError(t *testing.T, body *bytes.Buffer) dto.ErrorResponseDTO {
	t.Helper()
	var resp dto.ErrorResponseDTO
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return resp
}
