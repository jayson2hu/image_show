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
			Label    string `json:"label"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode prompt templates: %v", err)
	}
	seen := map[string]bool{}
	seenStyle := map[string]bool{}
	for _, item := range response.Items {
		seen[item.Category] = true
		if item.Category == "style" {
			seenStyle[item.Label] = true
		}
	}
	if !seen["style"] || !seen["sample"] || !seen["scene"] {
		t.Fatalf("expected default style, sample and scene templates, got %#v", seen)
	}
	if !seenStyle["插画"] {
		t.Fatalf("expected default illustration style, got %#v", seenStyle)
	}
}

func TestGenerationScenesReturnsDefaultScenes(t *testing.T) {
	engine := setupAuthTest(t)
	rec := adminRequest(engine, http.MethodGet, "/api/generation/scenes", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("generation scenes=%d body=%s", rec.Code, rec.Body.String())
	}

	var response struct {
		Items []struct {
			Name             string  `json:"name"`
			Icon             string  `json:"icon"`
			PromptTemplate   string  `json:"prompt_template"`
			RecommendedRatio string  `json:"recommended_ratio"`
			Description      string  `json:"description"`
			CreditCost       float64 `json:"credit_cost"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode scenes: %v", err)
	}
	if len(response.Items) != 6 {
		t.Fatalf("expected 6 scenes, got %#v", response.Items)
	}
	if response.Items[0].Name != "小红书封面" || response.Items[0].RecommendedRatio != "portrait_3_4" || response.Items[0].CreditCost != 2 {
		t.Fatalf("unexpected first scene: %#v", response.Items[0])
	}
	if response.Items[5].Name != "自由创作" || response.Items[5].PromptTemplate != "" || response.Items[5].RecommendedRatio != "square" {
		t.Fatalf("unexpected free scene: %#v", response.Items[5])
	}
}
