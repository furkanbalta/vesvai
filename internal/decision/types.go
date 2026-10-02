package decision

type QuestionType string

const (
	QuestionNoul   QuestionType = "noul"
	QuestionChoice QuestionType = "choice"
	QuestionScore  QuestionType = "score"
)

type Question struct {
	Type         QuestionType `json:"type"`
	Instructions string       `json:"instructions,omitempty"`
	Criteria     any          `json:"criteria,omitempty"`
}

func BoolQuestion(instructions string, criteria map[string]string) Question {
	return Question{Type: QuestionNoul, Instructions: instructions, Criteria: criteria}
}

func ChoiceQuestion(instructions string, criteria map[string]string) Question {
	return Question{Type: QuestionChoice, Instructions: instructions, Criteria: criteria}
}

func ScoreQuestion(instructions string, criteria []string) Question {
	return Question{Type: QuestionScore, Instructions: instructions, Criteria: criteria}
}

type Request struct {
	Model     string              `json:"model"`
	State     string              `json:"state"`
	Questions map[string]Question `json:"questions"`
}

type Answer struct {
	Type          QuestionType       `json:"type"`
	Noul          *float64           `json:"noul,omitempty"`
	Choice        string             `json:"choice,omitempty"`
	Score         *float64           `json:"score,omitempty"`
	Probabilities map[string]float64 `json:"probabilities,omitempty"`
	Legend        map[string]string  `json:"legend,omitempty"`
	Confidence    *float64           `json:"confidence,omitempty"`
}

func (a Answer) NoulValue() (float64, bool) {
	if a.Noul == nil {
		return 0, false
	}
	return *a.Noul, true
}

func (a Answer) ScoreValue() (float64, bool) {
	if a.Score == nil {
		return 0, false
	}
	return *a.Score, true
}

func (a Answer) ConfidenceValue() (float64, bool) {
	if a.Confidence == nil {
		return 0, false
	}
	return *a.Confidence, true
}

type Usage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	Cost         float64 `json:"cost,omitempty"`
}

type Response struct {
	ID       string            `json:"id"`
	Model    string            `json:"model"`
	Provider string            `json:"provider"`
	Answers  map[string]Answer `json:"answers"`
	Usage    Usage             `json:"usage,omitempty"`
}
