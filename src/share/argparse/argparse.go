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

type parser struct {
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

func New(program_name, program_bin, short_descr string) *parser {
	p := parser{}
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

func (p *parser) add(opt *option, name, short string) {

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

func (p *parser) PosString(target *string, name, descr string, values []string) {
    opt := option{name: name, class: "string", descr: descr, accept: values, target_string: target}
    p.positionals = append(p.positionals, &opt)
}

func (p *parser) Bool(target *bool, name, short, descr string) {
	opt := option{name: name, class: "bool", short: short, descr: descr, target_bool: target}
	p.add(&opt, name, short)
}

func (p *parser) String(target *string, name, short, descr string) {
	opt := option{name: name, class: "string", short: short, descr: descr, target_string: target}
	p.add(&opt, name, short)
}

func (p *parser) Int(target *int, name, short, descr string) {
	opt := option{name: name, class: "int", short: short, descr: descr, target_int: target}
	p.add(&opt, name, short)
}

func (p *parser) Slice(target *[]string, name, short, descr string) {
	opt := option{name: name, class: "slice", short: short, descr: descr, target_slice: target}
	p.add(&opt, name, short)
}

func (p *parser) SetHelp(help string) {
	p.help = help
}

func (p *parser) Parse() bool {

	pos_index := 0

	for i:=1; i<len(os.Args); i+=1 {

		flag := os.Args[i]

		// help stops here
		if flag == "-h" || flag == "--help" || flag == "-help" {
			p.PrintHelp()
			return false
		// long flag
		} else if len(flag) > 1 && flag[:2] == "--" {
			i += p.handle_long(i, flag[2:])
		// short flag
		} else if string(flag[0]) == "-" {
			i += p.handle_short(i, string(flag[1:]))
		// positional
		} else if len(p.positionals) != 0 {
			p.handle_pos(pos_index, flag)
			pos_index += 1
		} else {
			p.errors = append(p.errors, []string{flag, "unknown command"})
		}
	}

	if len(p.errors) == 0 {
		return true
	}

	p.PrintErrors()
	return false
}


func (p *parser) handle_pos(i int, flag string) {

	if len(p.positionals) <= i {
		return
	}

	opt := p.positionals[i]

	if len(opt.accept) == 0 {
		*opt.target_string = flag
		return
	}

	for _, x := range opt.accept {
		if x == flag {
			*opt.target_string = flag
			return
		}
	}

	p.errors = append(p.errors, []string{flag, "invalid value for " + opt.name})

}


func (p *parser) handle_long(i int, name string) int {

	//fmt.Printf("parsing long [%s]\n", name)

	var opt *option

	if index, exists := p.names[name]; !exists {
		p.errors = append(p.errors, []string{name, "undefined"})
		return 0
	} else {
		opt = p.options[index]
	}

	if opt.class == "bool" {
		*opt.target_bool = true
		return 0
	}

	if len(os.Args) < i+2 {
		p.errors = append(p.errors, []string{name, "value missing"})
		return 0
	}

	value := os.Args[i+1]

	if opt.class == "int" {
		if x, err := strconv.Atoi(value); err != nil {
			p.errors = append(p.errors, []string{name, "invalid int " + value})
			return 0
		} else {
			*opt.target_int = x
			return 1
		}
	}

	if opt.class == "string" {
		*opt.target_string = value
		return 1
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
		return k
	}
	panic("invalid option type: " + opt.class)
}


func (p *parser) handle_short(i int, name string) int {

	//fmt.Printf("parsing shorts [%s]\n", name)

	var k int
	k = 0
	for j:=0; j<len(name); j+=1 {

		//fmt.Printf(" short= [%s]\n", string(name[j]))

		if index, exists := p.shorts[string(name[j])]; exists {
			//fmt.Printf(" => long=[%s]\n", o.options[index].name)
			k += p.handle_long(i+k, p.options[index].name)
		} else {
			p.errors = append(p.errors, []string{string(name[j]), "undefined"})
		}
	}
	return k
}

func (p *parser) PrintHelp() {

	if p.help != "" {
		fmt.Println(p.help)
		return
	}

	fmt.Printf("%s - %s\n\n", p.name, p.descr)
	fmt.Printf("Options are:\n\n")

	for _, opt := range p.options {
		flag := opt.name
		if opt.short != "" {
			flag += ", -" + opt.short
		}
		fmt.Printf("    %-20s %s\n", flag, opt.descr)
	}
	fmt.Println()
	fmt.Println()
}

func (p *parser) PrintErrors() {
	fmt.Printf("Error parsing arguments:\n\n")
	for _, err := range p.errors {
		fmt.Printf("    %-20s %s\n", err[0], err[1])
	}
}

func (p *parser) PrintValues() {

	fmt.Printf("\npositional arguments:\n\n")
	for i, opt := range p.positionals {
		fmt.Printf(" (%d)    %-20s %s", i, opt.name, opt.value2string())
	}

	fmt.Printf("\n\nnamed arguments:\n\n")
	for _, opt := range p.options {
		fmt.Printf("        %-20s %s", opt.name, opt.value2string())
		fmt.Println()
	}
}
