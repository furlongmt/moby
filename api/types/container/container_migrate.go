package container

// MATT ADDED THIS
type IterBody struct {

	// Warnings encountered when creating the page esrver
	// Required: true
	Warnings []string `json:"Warnings"`
}

type CreatePageServerBody struct {
	Port uint32

	// Warnings encountered when creating the page esrver
	// Required: true
	Warnings []string `json:"Warnings"`
}
