package lib

type packageContext struct {
	Name    string `macro:"name" key:"name:"`
	Version string `macro:"version" key:"version:"`
	Release string `macro:"release" key:"release:"`
	Summary string `macro:"summary" key:"summary:"`
	License string `macro:"license" key:"license:"`
	URL     string `macro:"url" key:"url:"`

	Sources []string
}
