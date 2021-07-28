package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func Run(command string, arg ...string) string {
	out, err := exec.Command(command, arg...).Output()
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		panic(err)
	}
	output := string(out[:])
	return output
}

func Exec(dir string, command string, arg ...string) string {
	cmd := exec.Command(command, arg...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		panic(err)
	}
	output := string(out[:])
	return output
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func RemoveSafely(filename string) bool {
	exist := FileExists(filename)
	if exist {
		err := os.Remove(filename)
		if err != nil {
			fmt.Println(err)
			return false
		}
		fmt.Println("File " + filename + " has been deleted.")
	}
	return true
}

func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
