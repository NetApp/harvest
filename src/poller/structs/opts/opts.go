package opts

import (
	"flag"
	"fmt"
	"os"
	"path"
)

type Opts struct {
    Daemon      bool
    Config      string
    Path        string
    Delay       int
    LogLevel    int
    Debug       bool
    Test        bool
}

func GetOpts() (*Opts, string, error)  {
	var args Opts
	var name string
    var err error
    args = Opts{}

    flag.StringVar(&name, "poller", "",
        "Poller name as defined in config")
    flag.BoolVar(&args.Daemon, "daemon", false,
        "Start as daemon")
    flag.StringVar(&args.Config, "config", "config.yaml",
        "Configuration file")
    flag.StringVar(&args.Path, "path", "",
        "Harvest installation directory")
    flag.IntVar(&args.Delay, "delay", 0,
        "Delay startup in seconds")
    flag.IntVar(&args.LogLevel, "loglevel", 0,
        "logging level, one of: debug, info, warning, error, critical")
    flag.BoolVar(&args.Debug, "debug", false,
        "Debug mode, no data will be exported")
    flag.BoolVar(&args.Test, "test", false,
        "Startup collectors and exporters, and exit")

    flag.Parse()

    if name == "" {
        fmt.Println("Missing required argument: poller")
        flag.PrintDefaults()
        os.Exit(1)
    }
    if args.Path == "" {
        var cwd string
		cwd, _ = os.Getwd()
        if base := path.Base(cwd); base == "poller" {
            fmt.Println("base=", base)
            cwd, _ = path.Split(cwd)
            fmt.Println("=> ", cwd)
        }
		if base := path.Base(cwd); base == "src" {
            fmt.Println("base=", base)
			cwd, _ = path.Split(cwd)
            fmt.Println("=> ", cwd)
		}
		args.Path = cwd
    }

    return &args, name, err
}
