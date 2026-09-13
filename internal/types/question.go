package types

type Question struct {
	ID       string   `json:"id"`
	Text     string   `json:"text"`
	Options  []string `json:"options"`
	Correct  string   `json:"correct"`
	Category string   `json:"category"`
}
