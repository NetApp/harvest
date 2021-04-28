//
// Copyright NetApp Inc, 2021 All rights reserved
//
// Package Description:
//
// Examples:
//
package logger

import (
	"fmt"
	"goharvest2/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	LOG_FLAGS  int         = log.Ldate | log.Ltime | log.Lmsgprefix
	FILE_FLAGS int         = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	FILE_PERM  os.FileMode = 0644
	DIR_PERM   os.FileMode = 0755
)

var file *os.File

var levels = [6]string{
	" (trace  ) %-12s: ",
	" (debug  ) %-12s: ",
	" (info   ) %-12s: ",
	" (warning) %-12s: ",
	" (error  ) %-12s: ",
	" (fatal  ) %-12s: ",
}

var level = 2

func OpenFileOutput(dirpath, filename string) error {
	var info os.FileInfo
	var err error

	info, err = os.Stat(dirpath)
	if err != nil || !info.IsDir() {
		err = os.Mkdir(dirpath, DIR_PERM)
	}
	if err == nil || os.IsExist(err) {

		file, err = os.OpenFile(path.Join(dirpath, filename), FILE_FLAGS, FILE_PERM)
		if err == nil {
			log.SetOutput(file)
		}
	}
	return err
}

func CloseFileOutput() error {
	return file.Close()
}

func Rotate(dirpath, filename string, maxfiles int) error {
	var (
		files                       []os.FileInfo
		rotated                     []string
		err                         error
		curr_filepath, new_filepath string
	)

	curr_filepath = path.Join(dirpath, filename)
	new_filepath = path.Join(dirpath, filename+"."+"1")

	// list files in log folder, to rename older files
	if files, err = ioutil.ReadDir(dirpath); err != nil {
		return err
	}

	// rotate already existing backups
	rotated = make([]string, maxfiles) // not really necessary, only max index should be enough...

	for _, f := range files {
		if !f.IsDir() && strings.HasPrefix(f.Name(), filename+".") {
			if i, err := strconv.Atoi(strings.TrimPrefix(f.Name(), filename+".")); err == nil {
				// keep
				if i < maxfiles {
					rotated[i] = path.Join(dirpath, f.Name())
					// delete if index is higher than maxfiles
				} else {
					os.Remove(path.Join(dirpath, f.Name())) // ignore errs, not critical
				}
			}
		}
	}

	// rotate older files, starting from highest index
	for i := maxfiles - 1; i > 0; i -= 1 {
		if fp := rotated[i]; fp != "" {
			os.Rename(fp, path.Join(dirpath, filename+"."+strconv.Itoa(i+1)))
		}
	}

	// close current file
	if err = CloseFileOutput(); err != nil {
		return err
	}

	// send messages to void until we reopen file
	if err = OpenFileOutput("dev", "null"); err != nil {
		return err
	}

	os.Rename(curr_filepath, new_filepath) // catch and return err does not makes sense, probably we should panic

	CloseFileOutput()

	return OpenFileOutput(dirpath, filename)
}

func SetLevel(l int) error {
	var err error
	if l >= 0 && l < len(levels) {
		level = l
	} else {
		err = errors.New(errors.INVALID_PARAM, "level "+strconv.Itoa(l))
	}
	return err
}

func Log(lvl int, prefix, format string, vars ...interface{}) {
	log.Printf(fmt.Sprintf(levels[lvl], prefix) + fmt.Sprintf(format, vars...))
}

func Trace(prefix, format string, vars ...interface{}) {
	if level == 0 {
		Log(0, prefix, format, vars...)
	}
}

func Debug(prefix, format string, vars ...interface{}) {
	if level <= 1 {
		Log(1, prefix, format, vars...)
	}
}

func Info(prefix, format string, vars ...interface{}) {
	if level <= 2 {
		Log(2, prefix, format, vars...)
	}
}

func Warn(prefix, format string, vars ...interface{}) {
	if level <= 3 {
		Log(3, prefix, format, vars...)
	}
}

func Error(prefix, format string, vars ...interface{}) {
	if level <= 4 {
		Log(4, prefix, format, vars...)
	}
}

func Fatal(prefix, format string, vars ...interface{}) {
	if level <= 5 {
		Log(5, prefix, format, vars...)
	}
}
