package recruitment

type Recruiter struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Personality string `json:"personality"`
}

type Prospect struct {
	Name              string `json:"name"`
	Email             string `json:"email"`
	Organization      string `json:"organization"`
	AssignedRecruiter string `json:"assigned_recruiter"`
	AlreadySent       bool   `json:"already_sent"`

	Extra map[string]interface{} `json:"-"` // for unmarshalling
}
