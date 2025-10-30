package version

var (
	Version   string
	GoVersion string
	Commit    string
	BuildTime string
)

type VersionInfo struct {
	Version   string `json:"version"`
	GoVersion string `json:"goVersion"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:   Version,
		GoVersion: GoVersion,
		Commit:    Commit,
		BuildTime: BuildTime,
	}
}
