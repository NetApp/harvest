package template

import (
    "io/ioutil"
    "bytes"
    "strconv"
    "errors"
    "fmt"
    "strings"
    "local.host/share"
)


func New(filepath string) (*Element, error) {
    var err error
    var root *Element
    var content []byte

    content, err = ioutil.ReadFile(filepath)
    if err == nil {
        root = &Element{ name : "Root" }
        err = root.parse(bytes.Split(content, []byte("\n")), 0, 0)
    }
    return root, err
}


type Element struct {
    parent *Element
    name string
    value string
    values []string
    children []*Element
}

/* Public methods */

func (e *Element) Name() string {
    return e.name
}

func (e *Element) Value() string {
    return e.value
}

func (e *Element) Values() []string {
    return e.values
}

func (e *Element) Children() []*Element {
    return e.children
}

func (e *Element) GetChild(name string) *Element {
    var child *Element
    for _, child = range e.children {
        if child.name == name {
            return child
        }
    }
    return nil
}

func (e *Element) HasChild(name string) bool {
    return e.GetChild(name) != nil
}


func (e *Element) GetChildValue(name string) string {

    var child *Element

    child = e.GetChild(name)
    if child == nil {
        return ""
    }
    return child.value
}


func (e *Element) MergeFrom(dest *Element) {
    var children []*Element
    var child *Element

    for _, child = range dest.children {
        if e.HasChild(child.name) {
            children = append(children, e.GetChild(child.name))
        } else {
            children = append(children, child)
        }
    }
    for _, child = range e.children {
        if !dest.HasChild(child.name) {
            children = append(children, child)
        }
    }
    e.children = children
}


/* Function for debugging  */
func (e *Element) PrintTree(depth int) {
    var value, values, name string
    var i int
    var child *Element

    for i, value = range e.values {
        if i != 0 {
            values += ", "
        }
        values += value
    }

    name = fmt.Sprintf("%s%s%s%s%s (%d)", strings.Repeat("  ", depth), share.Bold, share.Cyan, e.name, share.End, len(e.children))
    fmt.Printf("%-25s - %s%-15s%s - %s%s%s\n", name, share.Green, e.value, share.End, share.Pink, values, share.End)
    for _, child = range e.children {
        child.PrintTree(depth+1)
    }
}


/* Private methods */

func (e *Element) parse(lines [][]byte, index, depth int) error {
    var indent int
    var child *Element
    var key, value string

    /* Reached end of file */
    if index == len(lines) {
        return nil
    }

    /* Check for consistent indentation */
    if depth % 2 != 0 {
        return errors.New("Inconsistent indentation at line " + strconv.Itoa(index+1))
    }

    indent, key, value = parse_line(lines[index])

    /* Skip empty line */
    if len(key) == 0 && len(value) == 0 {
        return e.parse(lines, index+1, depth)
    }

    /* Indentation is same, so parse for current element */
    if indent == depth {

        /* key was declared on previous line,
        means this is an element of a list
        */
        if len(key) == 0 {
            e.values = append(e.values, value)
            /* continue parsing next line */
            return e.parse(lines, index+1, depth)
        } else {  /* create new element */
            child = &Element{ name : key }
            child.parent = e
            e.children = append(e.children, child)

            if len(value) == 0 {  /* expect child values on next line(s) */
                return child.parse(lines, index+1, depth+2)
            } else {  /* child is single-valued */
                child.value = value
                return e.parse(lines, index+1, depth)
            }
        }
    /* Jump back to previous element */
    } else if indent < depth {
        return e.parent.parse(lines, index, depth-2)
    /* Current indent can't be larger than previous */
    } else {
        return errors.New("Invalid indentation at " + strconv.Itoa(index+1))
    }
}

func parse_line(line []byte) (int, string, string) {
    /* variables holding indices of:
        start = position of first non-whitespace character, i.e. where indentation ends
        end = end of line or position where comment starts (#....)
        mid = position of key-value separator (:) */
    var start, end, mid int
    var key, value []byte

    /* find possible comment, to ignore */
    for end=0; end<len(line) && line[end] != '#'; end+=1 {
        ;
    }

    /* find first non-space character */
    for start=0; start<end && line[start] == ' '; start+=1 {
        ;
    }

    /* find key-value separator */
    for mid=start; mid<end && line[mid] != ':'; mid+=1 {
        ;
    }

    /* no separator, means we only have value */
    if mid == end {
        value = line[start:mid]
    } else if mid+1 == end {  /* seperator is last character, so we only have key */
        key = line[start:mid]
    } else {  /* both */
        key = line[start:mid]
        value = line[mid+2:end]
    }

    key = bytes.TrimPrefix(key, []byte("- "))
    value = bytes.TrimLeft(value, " ")
    value = bytes.TrimPrefix(value, []byte("- "))

    return start, string(key), string(value)
}

