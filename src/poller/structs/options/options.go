package options

import (
	"flag"
	"fmt"
	"os"
	"path"
)

type Options struct {
    Poller      string
    Daemon      bool
    Config      string
    Path        string
    Delay       int
    LogLevel    int
    Debug       bool
    Test        bool
    Collector   string
    Object      string
    Exporter    string
    Version     string
}

func GetOpts() (*Options, string, error)  {
	var args Options
    var err error
    args = Options{}

    flag.StringVar(&args.Poller, "poller", "",
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
    flag.StringVar(&args.Collector, "collector", "",
        "Only run this collector (overrides config)")
    flag.StringVar(&args.Object, "object", "",
        "Only run this object (overrides template)")
    flag.StringVar(&args.Exporter, "exporter", "",
            "Only run this exporter (overrides config)")

    flag.Parse()

    if args.Poller == "" {
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

    args.Version = "2.0.1"

    return &args, args.Poller, err
}
