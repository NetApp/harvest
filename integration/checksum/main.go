package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/Netapp/harvest-automation/test/cmds"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	checksumsFilename = "sha256sums.txt"
)

var (
	dir              = ""
	checksumLocation = ""
)

func main() {
	cmds.SetupLogging()
	parseCLI()
	begin()
}

func parseCLI() {
	flag.StringVar(&dir, "dir", "", "Directory of files to checksum. Required")
	flag.StringVar(&checksumLocation, "out", ".", "Directory to write "+checksumsFilename)

	flag.Parse()
	if dir == "" {
		printRequired("dir")
	}
}

func begin() {
	slog.Info("Check checksums for files", slog.String("dir", dir))
	checksums, err := calculateSHA256s(dir)

	if err != nil {
		fatal(fmt.Errorf("failed to calculate checksums: %w", err))
	}

	file, err := os.Create(filepath.Join(checksumLocation, checksumsFilename))
	if err != nil {
		fatal(fmt.Errorf("failed to create checksums file: %w", err))
	}

	defer file.Close()

	for _, c := range checksums {
		if _, err := fmt.Fprintf(file, "%x  %s\n", c.checksum, c.filename); err != nil {
			fatal(fmt.Errorf("failed to write to checksums file: %w", err))
		}
	}

	slog.Info("Checksums written to file",
		slog.String("file", filepath.Join(checksumLocation, checksumsFilename)),
		slog.Int("count", len(checksums)),
	)
}

func fatal(err error) {
	slog.Error(err.Error())
	os.Exit(1)
}

func printRequired(name string) {
	fmt.Printf("%s is required\n", name)
	fmt.Printf("usage: \n")
	flag.PrintDefaults()
	os.Exit(1)
}

type checksumSHA256 struct {
	filename string
	checksum []byte
}

func calculateSHA256s(path string) ([]checksumSHA256, error) {
	var checksums []checksumSHA256
	path = fmt.Sprintf("%s%c", filepath.Clean(path), filepath.Separator)

	calculateSHA256 := func(filepath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		file, err := os.Open(filepath)
		if err != nil {
			return err
		}

		defer file.Close()

		hash := sha256.New()
		if _, err = io.Copy(hash, file); err != nil {
			return err
		}

		checksums = append(checksums, checksumSHA256{
			filename: strings.TrimPrefix(filepath, path),
			checksum: hash.Sum(nil),
		})

		return nil
	}

	if err := filepath.WalkDir(path, calculateSHA256); err != nil {
		return nil, err
	}
	return checksums, nil
}
