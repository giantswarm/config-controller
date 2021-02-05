package meta

var (
	Annotation annotation
	Label      label
)

type annotation struct {
	ConfigVersion
	XApp
	XCreator
	XInstallation
	XObjectHash
	XPreviousConfig
	XProjectVersion
}

type label struct {
}
