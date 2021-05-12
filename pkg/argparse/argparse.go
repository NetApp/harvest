/*
Copyright NetApp Inc, 2021 All rights reserved

package argparse provides ArgParse for parsing positional
and named flags from command-line arguments
*/
package argparse

import (
	"fmt"
	"goharvest2/pkg/set"
	"os"
	"strconv"
	"strings"
)

type ArgParse struct {
	name        string
	bin         string
	description string
	longFlags   map[string]int
	shortFlags  map[string]int
	helpFlags   *set.Set
	flags       []*flag
	positionals []*flag
	errors      [][]string
	index       int
	helpText    string
	offset      int
}

type flag struct {
	class        string
	name         string
	short        string
	descr        string
	accept       []string
	targetString *string
	targetInt    *int
	targetBool   *bool
	targetSlice  *[]string
}

func New(name, bin, description string) *ArgParse {
	return &ArgParse{
		name:        name,
		bin:         bin,
		description: description,
		flags:       make([]*flag, 0),
		longFlags:   make(map[string]int),
		shortFlags:  make(map[string]int),
		helpFlags:   set.New(),
		positionals: make([]*flag, 0),
		errors:      make([][]string, 0),
		index:       0,
		offset:      1,
	}
}

// PosString adds the string target as positinal argument
// if values is not nil, only elements of values will be accepted
func (p *ArgParse) PosString(target *string, name, descr string, values []string) {
	p.addPositional(&flag{
		class:        "string",
		name:         name,
		descr:        descr,
		accept:       values,
		targetString: target,
	})
}

// PosSlice adds slice of strings as positinal argument
func (p *ArgParse) PosSlice(target *[]string, name, descr string) {
	p.addPositional(&flag{
		class:       "slice",
		name:        name,
		descr:       descr,
		targetSlice: target,
	})
}

// Bool adds the boolean target as optional argument
func (p *ArgParse) Bool(target *bool, long, short, descr string) {
	p.addFlag(&flag{
		class:      "bool",
		name:       long,
		short:      short,
		descr:      descr,
		targetBool: target}, long, short)
}

// String adds the string target as optional argument
func (p *ArgParse) String(target *string, long, short, descr string) {
	p.addFlag(&flag{
		class:        "string",
		name:         long,
		short:        short,
		descr:        descr,
		targetString: target}, long, short)
}

// Int adds the int target as optional argument
func (p *ArgParse) Int(target *int, long, short, descr string) {
	p.addFlag(&flag{
		class:     "int",
		name:      long,
		short:     short,
		descr:     descr,
		targetInt: target}, long, short)
}

// Slice adds the target slice as optional argument
func (p *ArgParse) Slice(target *[]string, long, short, descr string) {
	p.addFlag(&flag{
		class:       "slice",
		name:        long,
		short:       short,
		descr:       descr,
		targetSlice: target}, long, short)
}

// SetHelpFlag adds name as a help flag
// (help will be printed if first argument is this flag)
func (p *ArgParse) SetHelpFlag(name string) {
	p.helpFlags.Add(name)
}

// SetHelp sets text as the help text
// If this is not set, then ArgParse will generate help text
func (p *ArgParse) SetHelp(text string) {
	p.helpText = text
}

// SetOffset defines the index from which CLI args should be parsed
// This method can be useful if some args need to be fixed
// (e.g. already parsed by parent program)
func (p *ArgParse) SetOffset(offset int) {
	p.offset = offset
}

// ParseOrExit will parse CLI arguments and exit if help is asked,
// arguments are incorrect or insufficient
func (p *ArgParse) ParseOrExit() {
	switch p.parse() {
	case "help":
		p.PrintHelp()
		os.Exit(0)
	case "usage":
		p.PrintUsage()
		os.Exit(0)
	case "errors":
		p.PrintErrors()
		os.Exit(1)
	}
}

// Parse will parse CLI arguments and return false if help
// is asked, arguments are incorrect or insufficient. Otherwise
// it returns true.
func (p *ArgParse) Parse() bool {
	return p.parse() == ""
}

func (p *ArgParse) parse() string {

	posIndex := 0
	argIndex := p.offset

	if argIndex+1 > len(os.Args) {
		return "usage"
	}

	// check first argument once, to stop if help is asked
	arg := os.Args[argIndex]
	if arg == "-h" || arg == "--help" || arg == "-help" || p.helpFlags.Has(arg) {
		return "help"
	}

	for argIndex < len(os.Args) {

		arg := os.Args[argIndex]

		// long flag
		if len(arg) > 1 && arg[:2] == "--" {
			i := p.handleLong(argIndex, arg[2:])
			argIndex += i
			// short flag
		} else if string(arg[0]) == "-" {
			argIndex += p.handleShort(argIndex, arg[1:])
			// positional
		} else if len(p.positionals) != 0 {
			argIndex += p.handlePos(argIndex, posIndex)
			posIndex++
		} else {
			p.errors = append(p.errors, []string{arg, "unknown command"})
			argIndex++
		}
	}

	if len(p.errors) != 0 {
		return "errors"
	}

	return ""
}

func (p *ArgParse) addFlag(f *flag, long, short string) {

	if index, exists := p.longFlags[long]; !exists {
		p.flags = append(p.flags, f)
		p.longFlags[long] = p.index
		if short != "" {
			p.shortFlags[short] = p.index
		}
		p.index += 1
		// override if same flag is added again
	} else {
		p.flags[index] = f
	}
}

func (p *ArgParse) addPositional(f *flag) {
	p.positionals = append(p.positionals, f)
}

// handle positional argument(s)
// return number of args parsed
func (p *ArgParse) handlePos(argIndex, posIndex int) int {

	if len(p.positionals) <= posIndex {
		p.errors = append(p.errors, []string{os.Args[argIndex], "invalid positional at " + strconv.Itoa(argIndex)})
		return 1
	}

	f := p.positionals[posIndex]

	if f.class == "string" {

		arg := os.Args[argIndex]
		if len(f.accept) == 0 {
			*f.targetString = arg
			return 1
		}

		for _, x := range f.accept {
			if x == arg {
				*f.targetString = arg
				return 1
			}
		}
		p.errors = append(p.errors, []string{arg, "invalid value for " + f.name})
		return 1
	} else if f.class == "slice" {
		var i int
		for i = 0; i+argIndex < len(os.Args); i += 1 {
			arg := os.Args[i+argIndex]
			if string(arg[0]) == "-" {
				break
			}
			*f.targetSlice = append(*f.targetSlice, arg)
		}
		return i

	}
	panic("invalid positinal type: " + f.class)
}

// handle optional argument with long flag
// return number of args parsed
// if this is simply a flag (e.g. "-verbose"), return 1
// if it's a flag with values (e.g. "--collectors"), returns 1 + number or values
func (p *ArgParse) handleLong(i int, name string) int {

	var f *flag

	if index, exists := p.longFlags[name]; !exists {
		p.errors = append(p.errors, []string{name, "undefined"})
		return 1
	} else {
		f = p.flags[index]
	}

	if f.class == "bool" {
		//fmt.Println(" ~> bool flag: ", name)
		*f.targetBool = true
		return 1
	}

	// for other types we expect flag to be followed by value(s)

	if len(os.Args) < i+2 {
		p.errors = append(p.errors, []string{name, "value missing"})
		return 1
	}

	value := os.Args[i+1]

	if f.class == "int" {
		if x, err := strconv.Atoi(value); err != nil {
			p.errors = append(p.errors, []string{name, "invalid int " + value})
			return 1
		} else {
			*f.targetInt = x
			return 2
		}
	}

	if f.class == "string" {
		*f.targetString = value
		return 2
	}

	if f.class == "slice" {
		var k int
		for k = 0; i+k+1 < len(os.Args); k += 1 {
			val := os.Args[i+k+1]
			if string(val[0]) == "-" {
				break
			}
			*f.targetSlice = append(*f.targetSlice, val)
		}
		return k + 1
	}
	panic("invalid flag type: " + f.class)
}

func (p *ArgParse) handleShort(i int, name string) int {

	k := 1
	for j := 0; j < len(name); j += 1 {

		//fmt.Printf(" short= [%s]\n", string(name[j]))

		if index, exists := p.shortFlags[string(name[j])]; exists {
			//fmt.Printf(" => long=[%s]\n", o.options[index].name)
			//@TODO will fail if multiple value assignments
			x := p.handleLong(i, p.flags[index].name)
			//fmt.Printf(" short ++ %d (-1) ==> ", x)
			k += x - 1
			//fmt.Printf(" %d\n", k)
		} else {
			p.errors = append(p.errors, []string{string(name[j]), "undefined"})
		}
	}
	return k
}

func (p *ArgParse) PrintUsage() {
	fmt.Printf("Usage: %s", p.bin)
	for _, f := range p.positionals {
		fmt.Printf(" <%s>", strings.ToUpper(f.name))
	}
	for _, f := range p.flags {
		if f.short != "" {
			fmt.Printf(" [-%s", f.short)
		} else {
			fmt.Printf(" [--%s", f.name)
		}
		if f.class != "bool" {
			fmt.Printf(" %s", strings.ToUpper(f.class))
		}
		fmt.Printf("]")

	}
	fmt.Println()
}

func (p *ArgParse) PrintHelp() {

	if p.helpText != "" {
		fmt.Println(p.helpText)
		return
	}

	fmt.Printf("%s - %s\n\n", p.name, p.description)

	if len(p.positionals) != 0 {
		fmt.Println("Positional arguments:")

		for _, x := range p.positionals {
			fmt.Printf("    %-20s %s\n", x.name, x.descr)
			if len(x.accept) != 0 {
				fmt.Printf("%-25s one of: %s\n", "", strings.Join(x.accept, ", "))
			}
		}
	}

	fmt.Println()

	if len(p.flags) != 0 {
		fmt.Printf("Optional arguments:\n")

		for _, f := range p.flags {
			arg := f.name
			if f.short != "" {
				arg += ", -" + f.short
			}
			fmt.Printf("    %-20s %s\n", arg, f.descr)
		}
		fmt.Println()
	}
}

func (p *ArgParse) PrintErrors() {
	fmt.Printf("Error parsing arguments:\n\n")
	for _, err := range p.errors {
		fmt.Printf("    %-20s %s\n", err[0], err[1])
	}
}

func (p *ArgParse) PrintValues() {

	fmt.Printf("\npositional arguments:\n\n")
	for i, f := range p.positionals {
		fmt.Printf(" (%d)    %-20s %s\n", i, f.name, f.value2string())
	}

	fmt.Printf("\n\nnamed arguments:\n\n")
	for _, f := range p.flags {
		fmt.Printf("        %-20s %s", f.name, f.value2string())
		fmt.Println()
	}
}

func (f *flag) value2string() string {

	switch f.class {

	case "string":
		if f.targetString != nil {
			return *f.targetString
		}
		return "<none>"
	case "bool":
		if f.targetBool != nil {
			if *f.targetBool {
				return "true"
			}
			return "false"
		}
		return "<none>"
	case "int":
		if f.targetInt != nil {
			return strconv.Itoa(*f.targetInt)
		}
		return "<none>"
	case "slice":
		if f.targetSlice != nil {
			return strings.Join(*f.targetSlice, ", ")
		}
		return "<none>"
	}
	return "<panic>"
}
