package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

func main() {
	setupLogging()
	c := newCli()
	c.initPrTypes()
	cobra.CheckErr(c.Root().Execute())
}

func newCli() *cli {
	return &cli{
		prsByKind: make(map[string][]pr),
		prTypes:   make(map[string]prType),
	}
}

type cli struct {
	title             string
	highlights        string
	releaseHighlights string
	prsByKind         map[string][]pr
	prTypes           map[string]prType
	prOrder           []string
	openIssues        []string
}

func (c *cli) makeChangelog() {
	highlights, err := os.ReadFile(c.highlights)
	if err != nil {
		log.Fatal().Err(err).Str("file", c.highlights).Msg("Failed to read file")
	}
	releaseNotes, err := os.ReadFile(c.releaseHighlights)
	if err != nil {
		log.Fatal().Err(err).Str("file", c.highlights).Msg("Failed to read file")
	}
	highlights = formatContributors(highlights)
	c.readPrs(releaseNotes)
	c.printChangelog(highlights)
}

func formatContributors(notes []byte) []byte {
	header := "## Thanks to all the awesome contributors"
	index := bytes.LastIndex(notes, []byte(header))
	if index == -1 {
		log.Fatal().
			Str("header", header).
			Msg("Release highlights does not contain the header. Please add so contributors can be sorted and included")
	}
	thanks := notes[index:]
	loc := regexp.MustCompile("(?m)^-").FindIndex(thanks)
	if loc == nil {
		log.Fatal().
			Msg("Release highlights does not contain a markdown list of contributors. Please add")
	}
	contributors := make([]string, 0)
	scanner := bufio.NewScanner(bytes.NewReader(thanks[loc[0]:]))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "-") {
			contributors = append(contributors, strings.TrimSpace(line[1:]))
		}
	}
	sort.Strings(contributors)
	names := strings.Join(contributors, ", ")

	buffer := bytes.NewBuffer(notes[:index+loc[0]])
	buffer.WriteString(names)
	return buffer.Bytes()
}

type pr struct {
	kind  string
	title string
	url   string
	id    int
}

func (c pr) linkToIssue() string {
	if c.id == -1 {
		return c.url
	}
	return fmt.Sprintf("([#%d](%s))", c.id, c.url)
}

// * docs: improve security panel info for ONTAP 9.10+ by @foo in https://github.com/NetApp/harvest/pull/1238
var prRegex = regexp.MustCompile(`\* (.*?): (.*?) by @(\w+) in (https://.*)$`)

func (c *cli) readPrs(notes []byte) {
	scanner := bufio.NewScanner(bytes.NewReader(notes))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "## New Contributors") {
			break
		}
		if !strings.HasPrefix(line, "*") || strings.HasPrefix(line, "**") {
			continue
		}
		matches := prRegex.FindStringSubmatch(line)
		if len(matches) != 5 {
			log.Warn().Str("line", line).Msg("expected 5 matches")
			continue
		}
		com := newPr(matches)
		list, ok := c.prsByKind[com.kind]
		if !ok {
			list = make([]pr, 0)
		}
		list = append(list, com)
		c.prsByKind[com.kind] = list
	}

	// sort prs by id
	for _, prs := range c.prsByKind {
		sort.SliceStable(prs, func(i, j int) bool {
			a := prs[i]
			b := prs[j]
			return a.id < b.id
		})
	}
}

func newPr(matches []string) pr {
	com := pr{
		kind:  matches[1],
		title: matches[2],
		url:   matches[4],
	}
	before, _, found := strings.Cut(com.kind, "(")
	if found {
		com.kind = before
	}
	splits := strings.Split(com.url, "/")
	if len(splits) == 1 {
		com.id = -1
	}
	id, err := strconv.Atoi(splits[len(splits)-1])
	if err != nil {
		log.Error().Err(err).Str("s", splits[len(splits)-1]).Msg("failed to convert s to int")
		com.id = -1
	}
	com.id = id
	if com.kind == "docs" {
		com.kind = "doc"
	}
	return com
}

func (c *cli) printChangelog(highlightBytes []byte) {
	fmt.Printf("## %s / %s\n", c.title, time.Now().Format("2006-01-02"))
	fmt.Printf(":pushpin: Highlights of this major release include:\n")
	highlights := string(highlightBytes)
	highlights = strings.TrimSpace(highlights)
	fmt.Println(highlights)
	c.printPrSummary()
	caser := cases.Title(language.Und)

	for _, kind := range c.prOrder {
		ct, ok := c.prTypes[kind]
		if !ok {
			log.Fatal().Str("kind", kind).Msg("missing kind")
		}
		prs := c.prsByKind[kind]
		if len(prs) == 0 {
			continue
		}
		fmt.Printf("\n### %s\n", ct.header)
		for _, pr := range prs {
			title := caser.String(pr.title)
			title = strings.TrimSpace(title)
			fmt.Printf("- %s %s\n", title, pr.linkToIssue())
			c.openIssue(pr)
		}
	}
	fmt.Printf("\n---\n")
}

func (c *cli) printPrSummary() {
	b := strings.Builder{}
	for i, k := range c.prOrder {
		prs, ok := c.prsByKind[k]
		if !ok {
			continue
		}
		pt, ok := c.prTypes[k]
		if !ok {
			log.Fatal().Str("kind", k).Msg("missing kind")
		}
		if i == len(c.prOrder)-1 {
			b.WriteString(fmt.Sprintf("and %d %s pull requests.", len(prs), pt.summary))
		} else {
			b.WriteString(fmt.Sprintf("%d %s, ", len(prs), pt.summary))
		}
	}
	fmt.Printf("\n:seedling: This release includes %s\n", b.String())
}

func (c *cli) openIssue(pr pr) {
	shouldOpen := false
	for _, kind := range c.openIssues {
		if pr.kind == kind {
			shouldOpen = true
			break
		}
	}
	if !shouldOpen {
		return
	}
	err := exec.Command("open", pr.url).Run() //nolint:gosec
	if err != nil {
		log.Error().Err(err).Str("url", pr.url).Msg("failed top open url")
		return
	}
	secs, _ := time.ParseDuration(strconv.Itoa(rand.Intn(4)) + "s") //nolint:gosec
	time.Sleep(secs)
}

type prType struct {
	id      string
	summary string
	header  string
}

func (c *cli) initPrTypes() {
	c.prOrder = []string{"feat", "fix", "doc", "test", "style", "refactor", "chore", "ci"}

	c.addPrType(prType{id: "feat", summary: "features", header: ":rocket: Features"})
	c.addPrType(prType{id: "fix", summary: "bug fixes", header: ":bug: Bug Fixes"})
	c.addPrType(prType{id: "doc", summary: "documentation", header: ":closed_book: Documentation"})
	c.addPrType(prType{id: "test", summary: "testing", header: ":wrench: Testing"})
	c.addPrType(prType{id: "style", summary: "styling", header: "Styling"})
	c.addPrType(prType{id: "refactor", summary: "refactoring", header: "Refactoring"})
	c.addPrType(prType{id: "chore", summary: "miscellaneous", header: "Miscellaneous"})
	c.addPrType(prType{id: "ci", summary: "ci", header: ":hammer: CI"})
}

func setupLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).With().Caller().Timestamp().Logger()
}

func (c *cli) Root() *cobra.Command {
	r := &cobra.Command{
		Use:   "changelog",
		Short: "create changelog",
		Run: func(cmd *cobra.Command, args []string) {
			c.makeChangelog()
		},
	}
	r.Flags().StringVarP(&c.title, "title", "t", "", "Title of release")
	r.Flags().StringVar(&c.highlights, "highlights", "", "Path to markdown file of release highlights")
	r.Flags().StringVarP(&c.releaseHighlights, "releaseHighlights", "r", "",
		"Path to GitHub generated release notes")
	r.Flags().StringSliceVar(&c.openIssues, "open", nil,
		"Prs of this type will be opened in a browser tab: "+strings.Join(c.prOrder, ", ")+", all")

	_ = r.MarkFlagRequired("title")
	_ = r.MarkFlagRequired("highlights")
	_ = r.MarkFlagRequired("releaseHighlights")
	return r
}

func (c *cli) addPrType(ct prType) {
	c.prTypes[ct.id] = ct
}
