package wrap

import (
	"strings"
	"unicode/utf8"
)

const (
	defaultBreakpoints = " -"
	defaultNewline     = "\n"
)

// Wrapper contains settings for customisable word-wrapping.
type Wrapper struct {
	// Breakpoints defines which characters should be able to break a line.
	// By default, this follows the usual English rules of spaces, and hyphens.
	// Default: " -"
	Breakpoints string

	// Newline defines which characters should be used to split and create new lines.
	// Default: "\n"
	Newline string

	// OutputLinePrefix is prepended to any output lines. This can be useful
	// for wrapping code-comments and prefixing new lines with "// ".
	// Default: ""
	OutputLinePrefix string

	// OutputLineSuffix is appended to any output lines.
	// Default: ""
	OutputLineSuffix string

	// LimitIncludesPrefixSuffix can be set to false if you don't want prefixes
	// and suffixes to be included in the length limits.
	// Default: true
	LimitIncludesPrefixSuffix bool

	// TrimPrefix can be set to remove a prefix on each input line.
	// This can be paired up with OutputPrefix to create a block of C-style
	// comments (/* * */ ) from a long single-line comment.
	// Default: ""
	TrimInputPrefix string

	// TrimSuffix can be set to remove a suffix on each input line.
	// Default: ""
	TrimInputSuffix string

	// StripTrailingNewline can be set to true if you want the trailing
	// newline to be removed from the return value.
	// Default: false
	StripTrailingNewline bool

	// CutLongWords will cause a hard-wrap in the middle of a word if the word's length exceeds the given limit.
	CutLongWords bool
}

// NewWrapper returns a new instance of a Wrapper initialised with defaults.
func NewWrapper() Wrapper {
	return Wrapper{
		Breakpoints:               defaultBreakpoints,
		Newline:                   defaultNewline,
		LimitIncludesPrefixSuffix: true,
	}
}

// Wrap is shorthand for declaring a new default Wrapper calling its Wrap method
func Wrap(s string, limit int) string {
	return NewWrapper().Wrap(s, limit)
}

// Wrap will wrap one or more lines of text at the given length.
// If limit is less than 1, the string remains unwrapped.
func (w Wrapper) Wrap(s string, limit int) string {

	// Subtract the length of the prefix and suffix from the limit
	// so we don't break length limits when using them.
	if w.LimitIncludesPrefixSuffix {
		limit -= utf8.RuneCountInString(w.OutputLinePrefix) + utf8.RuneCountInString(w.OutputLineSuffix)
	}

	var ret string
	for _, str := range strings.Split(s, w.Newline) {
		str = strings.TrimPrefix(str, w.TrimInputPrefix)
		str = strings.TrimSuffix(str, w.TrimInputSuffix)
		ret += w.line(str, limit) + w.Newline
	}

	if w.StripTrailingNewline {
		return strings.TrimSuffix(ret, w.Newline)
	}
	return ret
}

// line will wrap a single line of text at the given length.
// If limit is less than 1, the string remains unwrapped.
func (w Wrapper) line(s string, limit int) string {
	if limit < 1 || utf8.RuneCountInString(s) < limit+1 {
		return w.OutputLinePrefix + s + w.OutputLineSuffix
	}

	// Find the index of the last breakpoint within the limit.
	i := strings.LastIndexAny(s[:limit+1], w.Breakpoints)

	breakpointWidth := 1

	// Can't wrap within the limit
	if i < 0 {
		if w.CutLongWords {
			// wrap at the limit
			i = limit
			breakpointWidth = 0
		} else {
			// wrap at the next breakpoint instead
			i = strings.IndexAny(s, w.Breakpoints)
			// Nothing left to do!
			if i < 0 {
				return w.OutputLinePrefix + s + w.OutputLineSuffix
			}
		}
	}

	// Recurse until we have nothing left to do.
	return w.OutputLinePrefix + s[:i] + w.OutputLineSuffix + w.Newline + w.line(s[i+breakpointWidth:], limit)
}
