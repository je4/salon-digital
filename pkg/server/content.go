package server

type ContentPerson struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type ContentMedia struct {
	Type      string
	Thumbnail string
	Media     string
}

type Content struct {
	Title    string                  `json:"title"`
	Subtitle string                  `json:"subtitle"`
	Persons  []ContentPerson         `json:"persons"`
	Year     string                  `json:"year"`
	Poster   string                  `json:"poster"`
	Medias   map[string]ContentMedia `json:"medias"`
}
