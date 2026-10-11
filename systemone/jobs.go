package systemone

import "fmt"

// EvaluateRequest is the top-level shape of every evaluation request.
type EvaluateRequest struct {
	// State is the content to evaluate. It can be a plain string or structured data.
	State interface{} `json:"state"`

	// Model is the model identifier that handles the request (e.g. "jev-latest").
	Model string `json:"model"`

	// Questions is a map of question identifiers to Question definitions.
	Questions map[string]Question `json:"questions"`
}

// Question represents a single evaluation question.
// Use NewNoulQuestion, NewChoiceQuestion, or NewScoreQuestion to construct.
type Question struct {
	// Type is one of "noul", "choice", or "score".
	Type string `json:"type"`

	// Instructions describes what the model should evaluate.
	// Can be a string, object, or array.
	Instructions interface{} `json:"instructions"`

	// Criteria provides rubric or options for the evaluation.
	// Type varies: Noul uses NoulCriteria, Choice uses map[string]interface{}, Score uses []string.
	Criteria interface{} `json:"criteria,omitempty"`
}

// NoulCriteria provides descriptions for what a yes and a no mean.
type NoulCriteria struct {
	True  interface{} `json:"true,omitempty"`
	False interface{} `json:"false,omitempty"`
}

// EvaluateResponse is the response from the evaluation endpoint.
type EvaluateResponse struct {
	// Model is the model that performed the evaluation.
	Model string `json:"model"`

	// Answers contains one answer per question, keyed by the same ids from the request.
	Answers map[string]Answer `json:"answers"`

	// Usage reports token consumption.
	Usage Usage `json:"usage"`
}

// Usage reports token consumption for a request.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Answer is the result for a single evaluated question.
type Answer struct {
	// Type matches the question type ("noul", "choice", or "score").
	Type string `json:"type"`

	// Noul is the yes/no probability on a scale from 0 (no) to 1 (yes).
	// Only present when Type is "noul".
	Noul float64 `json:"noul,omitempty"`

	// Choice is the highest-probability option selected.
	// Only present when Type is "choice".
	Choice string `json:"choice,omitempty"`

	// Score is the probability-weighted value across the levels.
	// Only present when Type is "score".
	Score float64 `json:"score,omitempty"`

	// Legend maps each level index back to its description.
	// Only present when Type is "score".
	Legend map[string]string `json:"legend,omitempty"`

	// Probabilities maps every option/level to its probability (floats that sum to 1).
	// Present for "choice" and "score" types.
	Probabilities map[string]float64 `json:"probabilities,omitempty"`

	// Confidence indicates how certain the model is, derived from probabilities.
	// Present for "choice" and "score" types.
	Confidence float64 `json:"confidence,omitempty"`
}

// AsNoul returns the Noul value if this answer is a noul type.
func (a Answer) AsNoul() (float64, error) {
	if a.Type != "noul" {
		return 0, fmt.Errorf("answer is type %q, not noul", a.Type)
	}
	return a.Noul, nil
}

// AsChoice returns the selected option and probability distribution.
func (a Answer) AsChoice() (choice string, probabilities map[string]float64, confidence float64, err error) {
	if a.Type != "choice" {
		return "", nil, 0, fmt.Errorf("answer is type %q, not choice", a.Type)
	}
	return a.Choice, a.Probabilities, a.Confidence, nil
}

// AsScore returns the score value, legend, and probability distribution.
func (a Answer) AsScore() (score float64, legend map[string]string, probabilities map[string]float64, confidence float64, err error) {
	if a.Type != "score" {
		return 0, nil, nil, 0, fmt.Errorf("answer is type %q, not score", a.Type)
	}
	return a.Score, a.Legend, a.Probabilities, a.Confidence, nil
}

// NewNoulQuestion constructs a yes/no question.
// instructions can be a string, object, or array.
func NewNoulQuestion(instructions interface{}) Question {
	return Question{
		Type:         "noul",
		Instructions: instructions,
	}
}

// NewNoulQuestionWithCriteria constructs a yes/no question with true/false descriptions.
func NewNoulQuestionWithCriteria(instructions interface{}, trueDesc, falseDesc interface{}) Question {
	return Question{
		Type:         "noul",
		Instructions: instructions,
		Criteria: NoulCriteria{
			True:  trueDesc,
			False: falseDesc,
		},
	}
}

// NewChoiceQuestion constructs a multiple-choice question.
// options maps each choice key to its description (can be string, object, array, or nil).
func NewChoiceQuestion(instructions interface{}, options map[string]interface{}) Question {
	return Question{
		Type:         "choice",
		Instructions: instructions,
		Criteria:     options,
	}
}

// NewScoreQuestion constructs a rating question.
// levels is an ordered array of level descriptions.
func NewScoreQuestion(instructions interface{}, levels []string) Question {
	return Question{
		Type:         "score",
		Instructions: instructions,
		Criteria:     levels,
	}
}
