package options

import (
	"os"
	"fmt"
	"strconv"
)

type option struct {
	name string
	class string
	short string
	descr string
	deflt string
	values []string
}

type Options struct {
	name string
	descr string
	names map[string]int
	shorts map[string]int
	options []*option
	positionals []string
	errors [][]string
	index int
}

func New(program_name, short_descr string) *Options {
	o := Options{}
	o.name = program_name
	o.descr = short_descr
	o.options = make([]*option, 0)
	o.names = make(map[string]int)
	o.shorts = make(map[string]int)
	o.positionals = make([]string, 0)
	o.errors = make([][]string, 0)
	o.index = 0
	return &o
}

func (o *Options) add(opt *option, name, short, descr, deflt string) {

	opt.deflt = deflt
	opt.descr = descr
	opt.short = short

	if index, exists := o.names[name]; !exists {
		o.options = append(o.options, opt)
		o.names[name] = o.index
		if short != "" {
			o.shorts[short] = o.index
		}
		o.index += 1
	} else {
		o.options[index] = opt
	}
}

func (o *Options) AddBool(name, short, descr string, deflt bool) {
	opt := option{name: name, class: "bool"}
	opt.values = make([]string, 1)
	if deflt {
		o.add(&opt, name, short, descr, "true")
	} else {
		o.add(&opt, name, short, descr, "false")
	}
}

func (o *Options) AddString(name, short, descr, deflt string) {
	opt := option{name: name, class: "string"}
	opt.values = make([]string, 1)
	o.add(&opt, name, short, descr, deflt)
}

func (o *Options) AddInt(name, short, descr string, deflt int) {
	opt := option{name: name, class: "int"}
	opt.values = make([]string, 1)
	o.add(&opt, name, short, descr, strconv.Itoa(deflt))
}

func (o *Options) AddSlice(name, short, descr string) {
	opt := option{name: name, class: "slice"}
	opt.values = make([]string, 0)
	o.add(&opt, name, short, descr, "")
}

func (o *Options) GetBool(name string) (bool, bool) {
	var value, ok bool
	var raw string
	if index, exists := o.names[name]; exists {
		opt := o.options[index]
		if raw = opt.values[0]; raw != "" {
			ok = true
		} else {
			raw = opt.deflt
			ok = false
		}
		if raw == "true" {
			value = true
		} else {
			value = false
		}
		return value, ok
	}
	panic("invalid bool flag: " + name)
}

func (o *Options) GetInt(name string) (int, bool) {
	var value int
	var ok bool
	var raw string
	var err error
	if index, exists := o.names[name]; exists {
		opt := o.options[index]
		if raw = opt.values[0]; raw != "" {
			ok = true
		} else {
			raw = opt.deflt
			ok = false
		}
		if value, err = strconv.Atoi(raw); err != nil {
			ok = false
		}
		return value, ok
	}
	panic("invalid int flag: " + name)	
}

func (o *Options) GetString(name string) (string, bool) {
	var value string
	var ok bool
	if index, exists := o.names[name]; exists {
		opt := o.options[index]
		if value = opt.values[0]; value != "" {
			ok = true
		} else {
			value = opt.deflt
			ok = false
		}
		return value, ok
	}
	panic("invalid string flag: " + name)	
}


func (o *Options) GetSlice(name string) []string {
	if index, exists := o.names[name]; exists {
		return o.options[index].values
	}
	panic("invalid slice flag: " + name)	
}


func (o *Options) Parse() bool {

	for i:=1; i<len(os.Args); i+=1 {

		name := os.Args[i]

		if len(name) > 1 && name[:2] == "--" {
			i += o.handle_long(i, name[2:])
		} else if string(name[0]) == "-" {
			i += o.handle_short(i, string(name[1:]))
		} else {
			o.positionals = append(o.positionals, name)
		}
	}
	return len(o.errors) == 0
}

func (o *Options) ParseAndHandle() {
	if !o.Parse() {
		o.PrintErrors()
		os.Exit(1)
	} else if help, _ := o.GetBool("help"); help {
		o.PrintHelp()
		os.Exit(0)
	}
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
		opt.values[0] = "true"
		return 0
	}

	if len(os.Args) < i+2 {
		o.errors = append(o.errors, []string{name, "value missing"})
		return 0
	}

	value := os.Args[i+1]

	if opt.class == "int" {
		if _, err := strconv.Atoi(value); err != nil {
			o.errors = append(o.errors, []string{name, "invalid int " + value})
			return 0
		}
		opt.values[0] = value
		return 1
	}

	if opt.class == "string" {
		opt.values[0] = value
		return 1
	}

	if opt.class == "slice" {
		var k int
		for k=0; i+k+1 < len(os.Args); k+=1 {
			val := os.Args[i+k+1]
			if string(val[0]) == "-" {
				break
			}
			opt.values = append(opt.values, val)
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
	fmt.Printf("%s - %s\n\n", o.name, o.descr)
	fmt.Printf("Options are:\n\n")

	for _, opt := range o.options {
		flag := opt.name
		if opt.short != "" {
			flag += ", -" + opt.short
		}
		fmt.Printf("    %-20s %s", flag, opt.descr)
		if opt.deflt != "" {
			fmt.Printf(" (default: %s)", opt.deflt)
		}
		fmt.Println()
	}
}

func (o *Options) PrintErrors() {
	fmt.Printf("Error parsing arguments:\n\n")
	for _, err := range o.errors {
		fmt.Printf("    %-20s %s\n", err[0], err[1])
	}
}

func (o *Options) PrintValues() {
	fmt.Printf("Parsed arguments:\n\n")
	for _, opt := range o.options {
		//fmt.Printf("<%s>...\n", opt.name)
		x := true
		switch opt.class {
		case "string":
			v, ok := o.GetString(opt.name)
			fmt.Printf("    %-20s %s", opt.name, v)
			x = ok
		case "bool":
			v, ok := o.GetBool(opt.name)
			fmt.Printf("    %-20s %v", opt.name, v)
			x = ok
		case "int":
			v, ok := o.GetInt(opt.name)
			fmt.Printf("    %-20s %d", opt.name, v)
			x = ok
		case "slice":
			fmt.Printf("    %-20s %v", opt.name, o.GetSlice(opt.name))
		default:
			fmt.Printf("    flag [%s] with invalid type: %s\n", opt.name, opt.class)
			continue
		}
		if !x {
			fmt.Printf(" (default)")
		}
		fmt.Println()
	}
}