package share

import (
    "io/ioutil"
    "bytes"
    "fmt"
    "errors"
    "strconv"
)

func ImportTemplate(filepath string) (*Element, error) {
    var err error
    var root *Element
    var content []byte

    content, err = ioutil.ReadFile(filepath)
    if err == nil {
        root = NewElement([]byte("Root"))
        root.Parse(bytes.Split(content, []byte("\n")), 0, 0, &err)
    }
    return root, err
}


type Element struct {
    Parent *Element
    Name []byte
    Values [][]byte
    Children []*Element
}


func NewElement(name []byte) *Element {
    return &Element{ Name: name }
}


func (e *Element) Parse(lines [][]byte, index, depth int, err *error) {
    var indent int
    var child *Element
    var key, value []byte

    /* Reached end of file */
    if index == len(lines) {
        return
    }

    /* Check for consistent indentation */
    if depth % 2 != 0 {
        e := errors.New("Inconsistent indentation at line " + strconv.Itoa(index))
        err = &e
        return
    }

    indent, key, value = ParseLine(lines[index])

    /* Skip empty line */
    if len(key) == 0 && len(value) == 0 {
        e.Parse(lines, index+1, depth, err)
        return
    }

    /* Indentation is same, so parse for current element */
    if indent == depth {

        /* key was declared on previous line,
        means this is an element of a list
        */
        if len(key) == 0 {
            e.Values = append(e.Values, value)
            /* continue parsing next line */
            e.Parse(lines, index+1, depth, err)
            return
        } else {  /* create new element */
            child = NewElement(key)
            child.Parent = e
            e.Children = append(e.Children, child)

            if len(value) == 0 {  /* expect child values on next line(s) */
                child.Parse(lines, index+1, depth+2, err)
                return
            } else {  /* child is single-valued */
                child.Values = append(child.Values, value)
                e.Parse(lines, index+1, depth, err)
                return
            }
        }
    /* Jump back to previous element */
    } else if indent < depth {
        e.Parent.Parse(lines, index, depth-2, err)
    /* Current indent can't be larger than previous */
    } else {
        e := errors.New("Invalid indentation at line " + strconv.Itoa(index))
        err = &e
        return
    }
}

/* Function for debugging lib */
func (e *Element) Print() {
    var child_names, values string
    var value []byte
    var i int
    var child *Element

    for i, child = range e.Children {
        if i != 0 {
            child_names += ", "
        }
        child_names += string(child.Name)
    }
    for i, value = range e.Values {
        if i != 0 {
            values += ", "
        }
        values += string(value)
    }
    fmt.Printf("%-35s [%-35s] %s\n", e.Name, child_names, values)

    for _, child = range e.Children {
        child.Print()
    }
}


func ParseLine(line []byte) (int, []byte, []byte) {
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

    return start, key, value
}
