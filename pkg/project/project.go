package project

var (
	description = "The config-controller does something."
	gitSHA      = "n/a"
	name        = "config-controller"
	source      = "https://github.com/giantswarm/config-controller"
	version     = "0.10.0"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
