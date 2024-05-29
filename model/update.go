package model

// UpdateOptions sets firmware update options for a device component
type UpdateOptions struct {
	AllowDowngrade    bool // Allow firmware to be downgraded
	InstallAll        bool // Install all available updates (vendor tooling like DSU fetches the updates and installs them)
	DownloadOnly      bool // Only download updates, skip install - Works with InstallAll (where updates are fetched by the vendor tooling)
	Serial            string
	Vendor            string
	Model             string
	Name              string
	Slug              string
	UpdateFile        string // Location of the UpdateFile to be installed
	InstallerVersion  string // The all available updates installer version (specific to dell DSU)
	RepositoryVersion string // The update repository version to activate when defined
	BaseURL           string // The BaseURL for the updates
}
