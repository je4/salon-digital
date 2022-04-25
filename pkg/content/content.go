package content

type Person struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type Media struct {
	Type      string
	Thumbnail string
	Media     string
}

type Content struct {
	Title    string             `json:"title"`
	Subtitle string             `json:"subtitle"`
	Persons  []Person           `json:"persons"`
	Year     string             `json:"year"`
	Poster   string             `json:"poster"`
	Medias   map[string][]Media `json:"medias"`
}
