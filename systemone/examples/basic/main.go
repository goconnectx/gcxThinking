package main

import (
	"context"
	"fmt"
	"log"

	"github.com/goconnectx/gcxThinking/systemone"
)

func main() {
	ctx := context.Background()

	client, err := systemone.NewClient(
		systemone.WithBaseURL("http://localhost:11434/v1/systemone"),
		systemone.WithAPIKey("your-api-key"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Example 1: Noul (yes/no) question
	fmt.Println("=== Noul question ===")
	resp, err := client.Evaluate(ctx, &systemone.EvaluateRequest{
		Model: "jev-latest",
		State: "Help! My payouts have been failing for 3 days.",
		Questions: map[string]systemone.Question{
			"is_urgent": systemone.NewNoulQuestionWithCriteria(
				"Does this convey urgency?",
				"Explicitly time-sensitive",
				"No urgency expressed",
			),
		},
	})
	if err != nil {
		log.Fatalf("Evaluate failed: %v", err)
	}

	urgency, err := resp.Answers["is_urgent"].AsNoul()
	if err != nil {
		log.Fatalf("Failed to get noul: %v", err)
	}
	fmt.Printf("Model: %s\n", resp.Model)
	fmt.Printf("Urgency probability: %.2f\n", urgency)
	fmt.Printf("Usage: %d input / %d output tokens\n\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)

	// Example 2: Choice question
	fmt.Println("=== Choice question ===")
	resp, err = client.Evaluate(ctx, &systemone.EvaluateRequest{
		Model: "jev-latest",
		State: "Help! My payouts have been failing for 3 days.",
		Questions: map[string]systemone.Question{
			"department": systemone.NewChoiceQuestion(
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
		log.Fatalf("Evaluate failed: %v", err)
	}

	choice, probs, confidence, err := resp.Answers["department"].AsChoice()
	if err != nil {
		log.Fatalf("Failed to get choice: %v", err)
	}
	fmt.Printf("Department: %s (confidence: %.2f)\n", choice, confidence)
	fmt.Printf("Probabilities: %v\n\n", probs)

	// Example 3: Score question
	fmt.Println("=== Score question ===")
	resp, err = client.Evaluate(ctx, &systemone.EvaluateRequest{
		Model: "jev-latest",
		State: "Help! My payouts have been failing for 3 days.",
		Questions: map[string]systemone.Question{
			"frustration": systemone.NewScoreQuestion(
				"How frustrated is the customer?",
				[]string{"Calm", "Frustrated", "Very angry"},
			),
		},
	})
	if err != nil {
		log.Fatalf("Evaluate failed: %v", err)
	}

	score, legend, _, confidence, err := resp.Answers["frustration"].AsScore()
	if err != nil {
		log.Fatalf("Failed to get score: %v", err)
	}
	fmt.Printf("Frustration score: %.2f (confidence: %.2f)\n", score, confidence)
	fmt.Printf("Legend: %v\n", legend)

	// Example 4: Multiple questions in one call
	fmt.Println("\n=== Multiple questions ===")
	resp, err = client.Evaluate(ctx, &systemone.EvaluateRequest{
		Model: "jev-latest",
		State: "I love this product! It works perfectly and the support team was amazing.",
		Questions: map[string]systemone.Question{
			"is_positive": systemone.NewNoulQuestion("Is this a positive message?"),
			"sentiment": systemone.NewScoreQuestion(
				"How positive is this message?",
				[]string{"Very negative", "Negative", "Neutral", "Positive", "Very positive"},
			),
		},
	})
	if err != nil {
		log.Fatalf("Evaluate failed: %v", err)
	}

	for id, ans := range resp.Answers {
		fmt.Printf("Answer '%s' (type: %s): ", id, ans.Type)
		switch ans.Type {
		case "noul":
			val, _ := ans.AsNoul()
			fmt.Printf("%.2f\n", val)
		case "score":
			val, _, _, _, _ := ans.AsScore()
			fmt.Printf("%.2f\n", val)
		case "choice":
			val, _, _, _ := ans.AsChoice()
			fmt.Printf("%s\n", val)
		}
	}
}
