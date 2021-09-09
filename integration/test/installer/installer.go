package installer

type Installer interface {
	Install() bool
	Upgrade() bool
}
