package models

type Problem struct {
	ProblemID  int    `json:"problem_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Difficulty string `json:"difficulty"`
}

func NewProblem(problemID int, name string, url string, difficulty string) *Problem {
	return &Problem{
		ProblemID:  problemID,
		Name:       name,
		URL:        url,
		Difficulty: difficulty,
	}
}
