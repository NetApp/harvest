package logger

import (
    "log"
    "fmt"
    "os"
    "path/filepath"
)

const flags int = log.Ldate | log.Ltime | log.Lshortfile | log.Lmsgprefix
var levels = [6]string{ " [TRACE  ] ", " [DEBUG  ] ", " [INFO   ] ", " [WARNING] ", " [ERROR  ] ", " [FATAL  ] " }

var file *os.File
const fileflags int = os.O_APPEND | os.O_CREATE | os.O_WRONLY
const fileperm os.FileMode = 0644
const dirperm os.FileMode = 0755

func OpenFileOutput(rootdir, filename string) error {
    var info os.FileInfo
    var err error

    fmt.Printf("Checking if dir [%s] exists...\n", filepath.Join(rootdir, "log"))

    info, err = os.Stat(filepath.Join(rootdir, "log"))
    if err != nil || info.IsDir() == true {
        fmt.Printf("Creating dir...\n")
        err = os.Mkdir(filepath.Join(rootdir, "log"), dirperm)
    }
    if err == nil || os.IsExist(err) {
        fmt.Printf("Opening file [%s]...\n", filepath.Join(rootdir, "log", filename))

        file, err = os.OpenFile(filepath.Join(rootdir, "log", filename), fileflags, fileperm)
        if err == nil {
            fmt.Printf("Setting as handler output (%T) %v\n", file, file)
            log.SetOutput(file)
        } else {
            fmt.Printf("Failed: %v\n", err)
        }
    }
    return err
}

func CloseFileOutput() error {
    return file.Close()
}

type Logger struct {
    level int
    handler *log.Logger
}

func New(level int, prefix string) *Logger {
    var L Logger
    if prefix != "" {
        prefix = fmt.Sprintf("[%12s]", prefix)
    }
    L = Logger{ level : level, handler : log.New(log.Writer(), prefix, flags) }
    return &L
}

func (L *Logger) Trace(format string, vars... interface{}) {
    if L.level == 0 {
        L.handler.Printf(levels[0] + format, vars...)
    }
}

func (L *Logger) Debug(format string, vars... interface{}) {
    if L.level < 2 {
        L.handler.Printf(levels[1] + format, vars...)
    }
}

func (L *Logger) Info(format string, vars... interface{}) {
    if L.level < 3 {
        L.handler.Printf(levels[2] + format, vars...)
    }
}

func (L *Logger) Warn(format string, vars... interface{}) {
    if L.level < 4 {
        L.handler.Printf(levels[3] + format, vars...)
    }
}

func (L *Logger) Error(format string, vars... interface{}) {
    if L.level < 5 {
        L.handler.Printf(levels[4] + format, vars...)
    }
}

func (L *Logger) Fatal(format string, vars... interface{}) {
    if L.level < 6 {
        L.handler.Printf(levels[5] + format, vars...)
    }
}

