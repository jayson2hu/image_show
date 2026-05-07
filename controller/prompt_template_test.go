package controller_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestPromptTemplatesDefaultIncludesHomeCategories(t *testing.T) {
	engine := setupAuthTest(t)
	rec := adminRequest(engine, http.MethodGet, "/api/prompt-templates", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("prompt templates=%d body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		Items []struct {
			Category string `json:"category"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode prompt templates: %v", err)
	}
	seen := map[string]bool{}
	for _, item := range response.Items {
		seen[item.Category] = true
	}
	if !seen["style"] || !seen["sample"] || !seen["scenario"] {
		t.Fatalf("expected default style, sample and scenario templates, got %#v", seen)
	}
}
