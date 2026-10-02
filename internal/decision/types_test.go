package decision

import (
	"testing"

	json "github.com/goccy/go-json"
)

func TestRequestMarshal(t *testing.T) {
	req := &Request{
		Model: "typesafe/jev-1.13",
		State: "Task: clean up inactive accounts.",
		Questions: map[string]Question{
			"safe_to_run": BoolQuestion("Is this safe?", map[string]string{
				"true":  "Reversible",
				"false": "Destructive",
			}),
			"team": ChoiceQuestion("Which team?", map[string]string{
				"billing": "Payments",
				"sales":   "Pricing",
			}),
			"urgency": ScoreQuestion("How urgent?", []string{"Calm", "Urgent"}),
		},
	}
	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	qs, ok := got["questions"].(map[string]any)
	if !ok {
		t.Fatalf("questions missing: %s", raw)
	}
	noul := qs["safe_to_run"].(map[string]any)
	if noul["type"] != "noul" {
		t.Errorf("noul type = %v", noul["type"])
	}
	if _, ok := noul["criteria"].(map[string]any); !ok {
		t.Errorf("noul criteria should be an object: %s", raw)
	}
	score := qs["urgency"].(map[string]any)
	if score["type"] != "score" {
		t.Errorf("score type = %v", score["type"])
	}
	if _, ok := score["criteria"].([]any); !ok {
		t.Errorf("score criteria should be an array: %s", raw)
	}
}

func TestAnswerUnmarshalNoul(t *testing.T) {
	var a Answer
	if err := json.Unmarshal([]byte(`{"type":"noul","noul":0.96,"confidence":0.9}`), &a); err != nil {
		t.Fatal(err)
	}
	if a.Type != QuestionNoul {
		t.Errorf("type = %q", a.Type)
	}
	if v, ok := a.NoulValue(); !ok || v != 0.96 {
		t.Errorf("noul = %v, %v", v, ok)
	}
	if v, ok := a.ConfidenceValue(); !ok || v != 0.9 {
		t.Errorf("confidence = %v, %v", v, ok)
	}
	if _, ok := a.ScoreValue(); ok {
		t.Error("score should not be set")
	}
}

func TestAnswerUnmarshalChoice(t *testing.T) {
	var a Answer
	raw := `{"type":"choice","choice":"billing","probabilities":{"sales":0,"billing":0.99,"technical":0.01},"confidence":0.99}`
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if a.Type != QuestionChoice || a.Choice != "billing" {
		t.Errorf("choice = %q, type = %q", a.Choice, a.Type)
	}
	if a.Probabilities["billing"] != 0.99 {
		t.Errorf("probabilities = %v", a.Probabilities)
	}
}

func TestAnswerUnmarshalScore(t *testing.T) {
	var a Answer
	raw := `{"type":"score","score":1.99,"legend":{"0":"Calm","1":"Urgent"},"probabilities":{"0":0.01,"1":0.99},"confidence":0.99}`
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if a.Type != QuestionScore {
		t.Errorf("type = %q", a.Type)
	}
	if v, ok := a.ScoreValue(); !ok || v != 1.99 {
		t.Errorf("score = %v, %v", v, ok)
	}
	if a.Legend["1"] != "Urgent" {
		t.Errorf("legend = %v", a.Legend)
	}
}

func TestResponseUnmarshal(t *testing.T) {
	raw := `{
		"id": "gen-dec-123",
		"model": "typesafe/jev-1.13-20260917",
		"provider": "TypeSafe",
		"answers": {
			"safe_to_run": {"type": "noul", "noul": 0.98}
		},
		"usage": {"input_tokens": 384, "output_tokens": 22, "cost": 0.000016128}
	}`
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != "gen-dec-123" || resp.Model != "typesafe/jev-1.13-20260917" {
		t.Errorf("resp = %+v", resp)
	}
	ans, ok := resp.Answers["safe_to_run"]
	if !ok {
		t.Fatal("answer missing")
	}
	if v, ok := ans.NoulValue(); !ok || v != 0.98 {
		t.Errorf("noul = %v", v)
	}
	if resp.Usage.InputTokens != 384 || resp.Usage.Cost == 0 {
		t.Errorf("usage = %+v", resp.Usage)
	}
}
