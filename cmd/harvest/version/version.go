/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package version

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	githubReleases = "https://github.com/NetApp/harvest/releases/latest"
	youHaveLatest  = "you have the latest ✓"
)

var (
	VERSION   = "2.0.2"
	Release   = "rc2"
	Commit    = "HEAD"
	BuildDate = "undefined"
)

func String() string {
	return fmt.Sprintf("harvest version %s-%s (commit %s) (build date %s) %s/%s\n",
		VERSION,
		Release,
		Commit,
		BuildDate,
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show application version and check for latest",
		Long:  "Show application version and check for latest",
		Run:   doVersion,
	}
}

func doVersion(cmd *cobra.Command, _ []string) {
	fmt.Printf(cmd.Root().Version)
	checkLatest()
}

func checkLatest() {
	//goland:noinspection GoBoolExpressions
	if Commit == "HEAD" {
		return
	}
	fmt.Printf("checking GitHub for latest...")
	latest, err := latestRelease()
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	available, err := isNewerAvailable(Release, latest)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	if available {
		fmt.Printf(" newer version %s available at %s\n", latest, githubReleases)
	} else {
		fmt.Printf(" %s\n", youHaveLatest)
	}
}

func isNewerAvailable(current string, remote string) (bool, error) {
	//fmt.Printf("isNewerAvail cur=%s remote=%s ", current, remote)
	if remote == current {
		return false, nil
	} else {
		remoteVersion, err := version.NewVersion(remote)
		if err != nil {
			return false, err
		}
		currentVersion, err := version.NewVersion(current)
		if err != nil {
			return false, err
		}
		if currentVersion.GreaterThanOrEqual(remoteVersion) {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func latestRelease() (string, error) {
	client := &http.Client{
		Transport: &http.Transport{},
		Timeout:   5 * time.Second,
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return fmt.Errorf("redirect")
	}
	resp, err := client.Get(githubReleases)
	// check if we got a redirect to the latest release
	if err != nil {
		var location *url.URL
		if resp == nil {
			return "", fmt.Errorf(" error checking GitHub %s", err)
		}
		if resp.StatusCode == http.StatusFound {
			location, err = resp.Location()
			if err != nil {
				return "", err
			}
			// returns like https://github.com/NetApp/harvest/releases/tag/v21.05.3
			lastSlash := strings.LastIndex(location.String(), "/")
			return location.String()[lastSlash+1:], nil
		}
	}
	return "", fmt.Errorf(" error checking GitHub")
}
