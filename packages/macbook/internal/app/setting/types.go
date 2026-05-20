package setting

type RuntimeSettings struct {
	RepoRoot     string `json:"repoRoot"`
	DatabasePath string `json:"databasePath"`
}

type Check struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Path    string `json:"path"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Validation struct {
	Settings RuntimeSettings `json:"settings"`
	Checks   []Check         `json:"checks"`
	Valid    bool            `json:"valid"`
}

const (
	CheckOK    = "ok"
	CheckError = "error"
)
