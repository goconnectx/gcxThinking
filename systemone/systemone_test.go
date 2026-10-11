package systemone

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(
		WithBaseURL("http://localhost:9999/v1/systemone"),
		WithAPIKey("test-key"),
		WithTimeout(5*time.Second),
		WithUserAgent("test-agent"),
	)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.baseURL.String() != "http://localhost:9999/v1/systemone" {
		t.Errorf("expected base URL, got %s", client.baseURL.String())
	}
	if client.apiKey != "test-key" {
		t.Error("expected API key to be set")
	}
	if client.httpClient.Timeout != 5*time.Second {
		t.Error("expected timeout to be 5s")
	}
	if client.userAgent != "test-agent" {
		t.Errorf("expected user agent test-agent, got %s", client.userAgent)
	}
}

func TestClient_Evaluate(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/v1/systemone", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected Content-Type application/json")
		}

		var req EvaluateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Model != "jev-latest" {
			t.Errorf("expected model jev-latest, got %s", req.Model)
		}
		if req.State != "Help! My payouts have been failing for 3 days." {
			t.Errorf("unexpected state: %v", req.State)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(EvaluateResponse{
			Model: "jev-1.13.0",
			Answers: map[string]Answer{
				"is_urgent": {
					Type: "noul",
					Noul: 0.95,
				},
			},
			Usage: Usage{InputTokens: 296, OutputTokens: 20},
		})
	})

	client, _ := NewClient(WithBaseURL(server.URL + "/v1/systemone"))

	resp, err := client.Evaluate(context.Background(), &EvaluateRequest{
		Model: "jev-latest",
		State: "Help! My payouts have been failing for 3 days.",
		Questions: map[string]Question{
			"is_urgent": NewNoulQuestion("Does this convey urgency?"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Model != "jev-1.13.0" {
		t.Errorf("expected model jev-1.13.0, got %s", resp.Model)
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(resp.Answers))
	}

	ans, ok := resp.Answers["is_urgent"]
	if !ok {
		t.Fatal("missing is_urgent answer")
	}
	if ans.Type != "noul" {
		t.Errorf("expected type noul, got %s", ans.Type)
	}
	if ans.Noul != 0.95 {
		t.Errorf("expected noul 0.95, got %f", ans.Noul)
	}

	val, err := ans.AsNoul()
	if err != nil {
		t.Fatalf("AsNoul failed: %v", err)
	}
	if val != 0.95 {
		t.Errorf("expected AsNoul 0.95, got %f", val)
	}
}

func TestClient_Evaluate_Choice(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/v1/systemone", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(EvaluateResponse{
			Model: "jev-1.13.0",
			Answers: map[string]Answer{
				"department": {
					Type:          "choice",
					Choice:        "billing",
					Probabilities: map[string]float64{"billing": 0.88, "technical": 0.12, "sales": 0.0},
					Confidence:    0.81,
				},
			},
			Usage: Usage{InputTokens: 318, OutputTokens: 34},
		})
	})

	client, _ := NewClient(WithBaseURL(server.URL + "/v1/systemone"))

	resp, err := client.Evaluate(context.Background(), &EvaluateRequest{
		Model: "jev-latest",
		State: "Help! My payouts have been failing for 3 days.",
		Questions: map[string]Question{
			"department": NewChoiceQuestion(
				"Which team should handle this?",
				map[string]interface{}{
					"billing":   "Payments, invoicing, refunds",
					"technical": "Bugs, outages, integrations",
					"sales":     "Pricing, upgrades, new accounts",
				},
			),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ans := resp.Answers["department"]
	choice, probs, conf, err := ans.AsChoice()
	if err != nil {
		t.Fatalf("AsChoice failed: %v", err)
	}
	if choice != "billing" {
		t.Errorf("expected choice billing, got %s", choice)
	}
	if probs["billing"] != 0.88 {
		t.Errorf("expected billing probability 0.88, got %f", probs["billing"])
	}
	if conf != 0.81 {
		t.Errorf("expected confidence 0.81, got %f", conf)
	}
}

func TestClient_Evaluate_Score(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/v1/systemone", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(EvaluateResponse{
			Model: "jev-1.13.0",
			Answers: map[string]Answer{
				"frustration": {
					Type:          "score",
					Score:         1.05,
					Legend:        map[string]string{"0": "Calm", "1": "Frustrated", "2": "Very angry"},
					Probabilities: map[string]float64{"0": 0.0, "1": 0.95, "2": 0.05},
					Confidence:    0.92,
				},
			},
			Usage: Usage{InputTokens: 304, OutputTokens: 18},
		})
	})

	client, _ := NewClient(WithBaseURL(server.URL + "/v1/systemone"))

	resp, err := client.Evaluate(context.Background(), &EvaluateRequest{
		Model: "jev-latest",
		State: "Help! My payouts have been failing for 3 days.",
		Questions: map[string]Question{
			"frustration": NewScoreQuestion(
				"How frustrated is the customer?",
				[]string{"Calm", "Frustrated", "Very angry"},
			),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ans := resp.Answers["frustration"]
	score, legend, probs, conf, err := ans.AsScore()
	if err != nil {
		t.Fatalf("AsScore failed: %v", err)
	}
	if score != 1.05 {
		t.Errorf("expected score 1.05, got %f", score)
	}
	if legend["1"] != "Frustrated" {
		t.Errorf("expected legend[1] Frustrated, got %s", legend["1"])
	}
	if probs["1"] != 0.95 {
		t.Errorf("expected probability[1] 0.95, got %f", probs["1"])
	}
	if conf != 0.92 {
		t.Errorf("expected confidence 0.92, got %f", conf)
	}
}

func TestClient_Evaluate_Error(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/v1/systemone", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(APIError{Code: "VALIDATION_ERROR", Message: "missing required field: state"})
	})

	client, _ := NewClient(WithBaseURL(server.URL + "/v1/systemone"))
	_, err := client.Evaluate(context.Background(), &EvaluateRequest{
		Model:     "jev-latest",
		Questions: map[string]Question{},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "missing required field: state" {
		t.Errorf("unexpected message: %s", apiErr.Message)
	}
	if !IsUnprocessable(err) {
		t.Error("expected IsUnprocessable to be true")
	}
}

func TestAnswer_TypeSafety(t *testing.T) {
	noulAns := Answer{Type: "noul", Noul: 0.5}
	if _, _, _, err := noulAns.AsChoice(); err == nil {
		t.Error("expected AsChoice to fail on noul answer")
	}
	if _, _, _, _, err := noulAns.AsScore(); err == nil {
		t.Error("expected AsScore to fail on noul answer")
	}

	choiceAns := Answer{Type: "choice", Choice: "a"}
	if _, err := choiceAns.AsNoul(); err == nil {
		t.Error("expected AsNoul to fail on choice answer")
	}

	scoreAns := Answer{Type: "score", Score: 1.0}
	if _, err := scoreAns.AsNoul(); err == nil {
		t.Error("expected AsNoul to fail on score answer")
	}
}

func TestIsRetryable(t *testing.T) {
	if IsRetryable(&APIError{StatusCode: 429}) != true {
		t.Error("expected 429 to be retryable")
	}
	if IsRetryable(&APIError{StatusCode: 529}) != true {
		t.Error("expected 529 to be retryable")
	}
	if IsRetryable(&APIError{StatusCode: 401}) != false {
		t.Error("expected 401 to not be retryable")
	}
	if IsRetryable(&APIError{StatusCode: 422}) != false {
		t.Error("expected 422 to not be retryable")
	}
}
