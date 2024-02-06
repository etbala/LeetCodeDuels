package models

// Problem represents a problem entity as stored in your database.
type Problem struct {
	ProblemID  int    `json:"problem_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Difficulty string `json:"difficulty"`
}

// NewProblem creates a new instance of Problem.
func NewProblem(problemID int, name string, url string, difficulty string) *Problem {
	return &Problem{
		ProblemID:  problemID,
		Name:       name,
		URL:        url,
		Difficulty: difficulty,
	}
}
