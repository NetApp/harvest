package argparse

import (
	"os"
	"fmt"
	"strconv"
	"strings"
)

type option struct {
	name string
	class string
	short string
	descr string
    accept []string
    target_string *string
    target_int *int
    target_bool *bool
    target_slice *[]string
}

func (opt *option) value2string() string {

	switch opt.class {

	case "string":
		if opt.target_string != nil {
			return *opt.target_string
		}
		return "<none>"
	case "bool":
		if opt.target_bool != nil {
			if *opt.target_bool {
				return "true"
			}
			return "false"
		}
		return "<none>"
	case "int":
		if opt.target_int != nil {
			return strconv.Itoa(*opt.target_int)
		}
		return "<none>"
	case "slice":
		if opt.target_slice != nil {
			return strings.Join(*opt.target_slice, ", ")
		}
		return "<none>"
	}

	return "<panic>"
}

type Parser struct {
	name string
	bin string
	descr string
	names map[string]int
	shorts map[string]int
	options []*option
	positionals []*option
	errors [][]string
	index int
	help string
}

func New(program_name, program_bin, short_descr string) *Parser {
	p := Parser{}
	p.name = program_name
	p.bin = program_bin
	p.descr = short_descr
	p.options = make([]*option, 0)
	p.names = make(map[string]int)
	p.shorts = make(map[string]int)
	p.positionals = make([]*option, 0)
	p.errors = make([][]string, 0)
	p.index = 0
	return &p
}

func (p *Parser) add(opt *option, name, short string) {

	if index, exists := p.names[name]; !exists {
		p.options = append(p.options, opt)
		p.names[name] = p.index
		if short != "" {
			p.shorts[short] = p.index
		}
		p.index += 1
	// jic same flag is added again
	} else {
		p.options[index] = opt
	}
}

func (p *Parser) PosString(target *string, name, descr string, values []string) {
    opt := option{name: name, class: "string", descr: descr, accept: values, target_string: target}
    p.positionals = append(p.positionals, &opt)
}

func (p *Parser) PosSlice(target *[]string, name, descr string) {
	opt := option{name: name, class: "slice", descr: descr, target_slice: target}
	p.positionals = append(p.positionals, &opt)
}

func (p *Parser) Bool(target *bool, name, short, descr string) {
	opt := option{name: name, class: "bool", short: short, descr: descr, target_bool: target}
	p.add(&opt, name, short)
}

func (p *Parser) String(target *string, name, short, descr string) {
	opt := option{name: name, class: "string", short: short, descr: descr, target_string: target}
	p.add(&opt, name, short)
}

func (p *Parser) Int(target *int, name, short, descr string) {
	opt := option{name: name, class: "int", short: short, descr: descr, target_int: target}
	p.add(&opt, name, short)
}

func (p *Parser) Slice(target *[]string, name, short, descr string) {
	opt := option{name: name, class: "slice", short: short, descr: descr, target_slice: target}
	p.add(&opt, name, short)
}

func (p *Parser) SetHelp(help string) {
	p.help = help
}

func (p *Parser) Parse() bool {

	pos_index := 0
	arg_index := 1

	for arg_index < len(os.Args) {

		flag := os.Args[arg_index]

        //fmt.Printf("%d - \"%s\"\n", arg_index, flag)

		// help stops here
		if flag == "-h" || flag == "--help" || flag == "-help" || flag == p.help {
			p.PrintHelp()
			return false
		// long flag
		} else if len(flag) > 1 && flag[:2] == "--" {
            i := p.handle_long(arg_index, flag[2:])
            //fmt.Printf("=> %d\n", i)
			arg_index += i
		// short flag
		} else if string(flag[0]) == "-" {
			arg_index += p.handle_short(arg_index, string(flag[1:]))
		// positional
		} else if len(p.positionals) != 0 {
			arg_index += p.handle_pos(arg_index, pos_index)
			pos_index += 1
		} else {
			p.errors = append(p.errors, []string{flag, "unknown command"})
		}

        //fmt.Printf("++ %d\n", arg_index)
	}

	if len(p.errors) == 0 {
		return true
	}

	p.PrintErrors()
	return false
}

// handle positional argument(s)
// return number of args parsed
func (p *Parser) handle_pos(arg_index, pos_index int) int {

	if len(p.positionals) <= pos_index {
		p.errors = append(p.errors, []string{os.Args[arg_index], "invalid positional at " + strconv.Itoa(arg_index)})
		return 1
	}

	opt := p.positionals[pos_index]

	if opt.class == "string" {

		flag := os.Args[arg_index]
		if len(opt.accept) == 0 {
			*opt.target_string = flag
			return 1
		}

		for _, x := range opt.accept {
			if x == flag {
				*opt.target_string = flag
				return 1
			}
		}
		p.errors = append(p.errors, []string{flag, "invalid value for " + opt.name})
		return 1
	} else if opt.class == "slice" {
		var i int
		for i=0; i+arg_index<len(os.Args); i+=1 {
			flag := os.Args[i+arg_index]
			if string(flag[0]) == "-" {
				break
			}
			*opt.target_slice = append(*opt.target_slice, flag)
		}
		//fmt.Printf(" ~> positional slice, count=%d\n", i)
		return i

	}
	panic("invalid option class: " + opt.class)
}

// handle optional argument with long flag
// return number of args parsed
// if this is simply a flag (e.g. "-verbose"), return 1
// if it's a flag with values (e.g. "--collectors"), returns 1 + number or values
func (p *Parser) handle_long(i int, name string) int {

	//fmt.Printf("~> (%d) parsing long: [%s]\n", i, name)

	var opt *option

	if index, exists := p.names[name]; !exists {
		p.errors = append(p.errors, []string{name, "undefined"})
		return 1
	} else {
		opt = p.options[index]
	}

	if opt.class == "bool" {
		//fmt.Println(" ~> bool flag: ", name)
		*opt.target_bool = true
		return 1
	}

	if len(os.Args) < i+2 {
		p.errors = append(p.errors, []string{name, "value missing"})
		return 1
	}

	value := os.Args[i+1]

	if opt.class == "int" {
		if x, err := strconv.Atoi(value); err != nil {
			p.errors = append(p.errors, []string{name, "invalid int " + value})
			return 1
		} else {
			*opt.target_int = x
			return 2
		}
	}

	if opt.class == "string" {
		*opt.target_string = value
		return 2
	}

	if opt.class == "slice" {
		var k int
		for k=0; i+k+1 < len(os.Args); k+=1 {
			val := os.Args[i+k+1]
			if string(val[0]) == "-" {
				break
			}
			*opt.target_slice = append(*opt.target_slice, val)
		}
		return k+1
	}
	panic("invalid option type: " + opt.class)
}


func (p *Parser) handle_short(i int, name string) int {

	//fmt.Printf("parsing short(s) [%s]\n", name)

    k := 1
	for j:=0; j<len(name); j+=1 {

		//fmt.Printf(" short= [%s]\n", string(name[j]))

		if index, exists := p.shorts[string(name[j])]; exists {
			//fmt.Printf(" => long=[%s]\n", o.options[index].name)
			//@TODO will fail if multiple value assignments
			x := p.handle_long(i, p.options[index].name)
			//fmt.Printf(" short ++ %d (-1) ==> ", x)
			k += (x-1)
			//fmt.Printf(" %d\n", k)
		} else {
			p.errors = append(p.errors, []string{string(name[j]), "undefined"})
		}
	}
	return k
}

func (p *Parser) PrintHelp() {

	if p.help != "" {
		fmt.Println(p.help)
		return
	}

	fmt.Printf("%s - %s\n\n", p.name, p.descr)

    if len(p.positionals) != 0 {
        fmt.Println("Positional arguments:\n")

        for _, x := range p.positionals {
		    fmt.Printf("    %-20s %s\n", x.name, x.descr)
            if len(x.accept) != 0 {
                fmt.Printf("%-25s one of: %s\n", "", strings.Join(x.accept, ", "))
            }
        }
    }

    fmt.Println()

    if len(p.options) != 0 {
	    fmt.Printf("Optional arguments:\n")

	    for _, opt := range p.options {
		    flag := opt.name
		    if opt.short != "" {
			    flag += ", -" + opt.short
		    }
		    fmt.Printf("    %-20s %s\n", flag, opt.descr)
	    }
	    fmt.Println()
    }
	fmt.Println()
}

func (p *Parser) PrintErrors() {
	fmt.Printf("Error parsing arguments:\n\n")
	for _, err := range p.errors {
		fmt.Printf("    %-20s %s\n", err[0], err[1])
	}
}

func (p *Parser) PrintValues() {

	fmt.Printf("\npositional arguments:\n\n")
	for i, opt := range p.positionals {
		fmt.Printf(" (%d)    %-20s %s\n", i, opt.name, opt.value2string())
	}

	fmt.Printf("\n\nnamed arguments:\n\n")
	for _, opt := range p.options {
		fmt.Printf("        %-20s %s", opt.name, opt.value2string())
		fmt.Println()
	}
}
