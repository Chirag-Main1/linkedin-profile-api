package models

type Profile struct {
	Username        string        `json:"username"`
	FirstName       string        `json:"first_name"`
	LastName        string        `json:"last_name"`
	Headline        string        `json:"headline"`
	Location        string        `json:"location"`
	About           string        `json:"about"`
	ProfileImageURL string        `json:"profile_image_url,omitempty"`
	Experience      []Experience  `json:"experience"`
	Education       []Education   `json:"education"`
	Skills          []Skill       `json:"skills"`
	Certifications  []Certification `json:"certifications"`
	Languages       []Language    `json:"languages"`
}

type Experience struct {
	Title       string `json:"title"`
	CompanyName string `json:"company_name"`
	Location    string `json:"location,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
	Description string `json:"description,omitempty"`
}

type Education struct {
	SchoolName  string `json:"school_name"`
	Degree      string `json:"degree,omitempty"`
	FieldOfStudy string `json:"field_of_study,omitempty"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type Skill struct {
	Name string `json:"name"`
}

type Certification struct {
	Name         string `json:"name"`
	Authority    string `json:"authority,omitempty"`
	LicenseNumber string `json:"license_number,omitempty"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
}

type Language struct {
	Name        string `json:"name"`
	Proficiency string `json:"proficiency,omitempty"`
}
