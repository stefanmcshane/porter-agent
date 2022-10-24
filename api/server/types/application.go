package types

// Application represents a Porter application
type Application struct {

	// Name of the Application
	Name string `json:"name"`

	// Namespace that the application is running in on the cluster
	Namespace string `json:"namespace"`

	// Sync status of the application
	Status string `json:"status"`

	// Revision is the commit SHA revision that an application is currently at.
	// Set this to a specific revision to attempt to roll the application back to
	Revision string `json:"revision"`
}
