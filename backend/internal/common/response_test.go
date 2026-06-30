package common

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAPIResponseKeepsEmptySuccessData(t *testing.T) {
	data := []string{}
	response := APIResponse[[]string]{
		Success: true,
		Data:    &data,
	}

	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if !strings.Contains(string(raw), `"data":[]`) {
		t.Fatalf("response = %s, want explicit empty data array", raw)
	}
}

func TestAPIResponseOmitsDataForErrors(t *testing.T) {
	response := APIResponse[any]{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input",
		},
	}

	raw, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	if strings.Contains(string(raw), `"data"`) {
		t.Fatalf("response = %s, want error response without data", raw)
	}
}
