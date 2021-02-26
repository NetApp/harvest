package options

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

type Options struct {
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

func New(program_name, program_bin, short_descr string) *Options {
	o := Options{}
	o.name = program_name
	o.bin = program_bin
	o.descr = short_descr
	o.options = make([]*option, 0)
	o.names = make(map[string]int)
	o.shorts = make(map[string]int)
	o.positionals = make([]*option, 0)
	o.errors = make([][]string, 0)
	o.index = 0
	return &o
}

func (o *Options) add(opt *option, name, short string) {

	if index, exists := o.names[name]; !exists {
		o.options = append(o.options, opt)
		o.names[name] = o.index
		if short != "" {
			o.shorts[short] = o.index
		}
		o.index += 1
	// jic same flag is added again
	} else {
		o.options[index] = opt
	}
}

func (o *Options) PosString(target *string, name, descr string, values []string) {
    opt := option{name: name, class: "string", descr: descr, accept: values, target_string: target}
    o.positionals = append(o.positionals, &opt)
}

func (o *Options) Bool(target *bool, name, short, descr string) {
	opt := option{name: name, class: "bool", short: short, descr: descr, target_bool: target}
	o.add(&opt, name, short)
}

func (o *Options) String(target *string, name, short, descr string) {
	opt := option{name: name, class: "string", short: short, descr: descr, target_string: target}
	o.add(&opt, name, short)
}

func (o *Options) Int(target *int, name, short, descr string) {
	opt := option{name: name, class: "int", short: short, descr: descr, target_int: target}
	o.add(&opt, name, short)
}

func (o *Options) Slice(target *[]string, name, short, descr string) {
	opt := option{name: name, class: "slice", short: short, descr: descr, target_slice: target}
	o.add(&opt, name, short)
}

func (o *Options) SetHelp(help string) {
	o.help = help
}

func (o *Options) Parse() bool {

	pos_index := 0

	for i:=1; i<len(os.Args); i+=1 {

		flag := os.Args[i]

		// help stops here
		if flag == "-h" || flag == "--help" || flag == "-help" {
			o.PrintHelp()
			return false
		// long flag
		} else if len(flag) > 1 && flag[:2] == "--" {
			i += o.handle_long(i, flag[2:])
		// short flag
		} else if string(flag[0]) == "-" {
			i += o.handle_short(i, string(flag[1:]))
		// positional
		} else if len(o.positionals) != 0 {
			o.handle_pos(pos_index, flag)
			pos_index += 1
		} else {
			o.errors = append(o.errors, []string{flag, "unknown command"})
		}
	}

	if len(o.errors) == 0 {
		return true
	}

	o.PrintErrors()
	return false
}


func (o *Options) handle_pos(i int, flag string) {

	if len(o.positionals) <= i {
		return
	}

	opt := o.positionals[i]

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

	o.errors = append(o.errors, []string{flag, "invalid value for " + opt.name})

}


func (o *Options) handle_long(i int, name string) int {

	//fmt.Printf("parsing long [%s]\n", name)

	var opt *option

	if index, exists := o.names[name]; !exists {
		o.errors = append(o.errors, []string{name, "undefined"})
		return 0
	} else {
		opt = o.options[index]
	}

	if opt.class == "bool" {
		*opt.target_bool = true
		return 0
	}

	if len(os.Args) < i+2 {
		o.errors = append(o.errors, []string{name, "value missing"})
		return 0
	}

	value := os.Args[i+1]

	if opt.class == "int" {
		if x, err := strconv.Atoi(value); err != nil {
			o.errors = append(o.errors, []string{name, "invalid int " + value})
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


func (o *Options) handle_short(i int, name string) int {

	//fmt.Printf("parsing shorts [%s]\n", name)

	var k int
	k = 0
	for j:=0; j<len(name); j+=1 {

		//fmt.Printf(" short= [%s]\n", string(name[j]))

		if index, exists := o.shorts[string(name[j])]; exists {
			//fmt.Printf(" => long=[%s]\n", o.options[index].name)
			k += o.handle_long(i+k, o.options[index].name)
		} else {
			o.errors = append(o.errors, []string{string(name[j]), "undefined"})
		}
	}
	return k
}

func (o *Options) PrintHelp() {

	if o.help != "" {
		fmt.Println(o.help)
		return
	}

	fmt.Printf("%s - %s\n\n", o.name, o.descr)
	fmt.Printf("Options are:\n\n")

	for _, opt := range o.options {
		flag := opt.name
		if opt.short != "" {
			flag += ", -" + opt.short
		}
		fmt.Printf("    %-20s %s\n", flag, opt.descr)
	}
	fmt.Println()
	fmt.Println()
}

func (o *Options) PrintErrors() {
	fmt.Printf("Error parsing arguments:\n\n")
	for _, err := range o.errors {
		fmt.Printf("    %-20s %s\n", err[0], err[1])
	}
}

func (o *Options) PrintValues() {

	fmt.Printf("\npositional arguments:\n\n")
	for i, opt := range o.positionals {
		fmt.Printf(" (%d)    %-20s %s", i, opt.name, opt.value2string())
	}

	fmt.Printf("\n\nnamed arguments:\n\n")
	for _, opt := range o.options {
		fmt.Printf("        %-20s %s", opt.name, opt.value2string())
		fmt.Println()
	}
}
