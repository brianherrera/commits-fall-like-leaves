package haiku

type Mood string

const (
	MoodHumerous   Mood = "humorous"
	MoodReflective Mood = "reflective"
	MoodTechnical  Mood = "technical"
)

type HaikuCommitRequest struct {
	CommitMessage string `json:"commitMessage" binding:"required"`
	Mood          Mood   `json:"mood,omitempty"`
}

type HaikuCommitResponse struct {
	Haiku string `json:"haiku"`
}

func (m Mood) IsValid() bool {
	switch m {
	case MoodHumerous, MoodReflective, MoodTechnical:
		return true
	}
	return false
}
