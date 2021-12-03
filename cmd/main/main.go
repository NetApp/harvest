package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	//p := "../../../grafana/dashboards/cmode"
	filepath.Walk("grafana/dashboards/cmode", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println(path, info.Size())
		return nil
	})
}
