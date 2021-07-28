package installer

import "fmt"

type Docker struct {
	path string
}

func (d *Docker) Init(path string) {
	d.path = path
}

func (d *Docker) Install() bool {
	fmt.Printf("Docker: %s", d.path)
	return true
}
