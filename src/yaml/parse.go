package yaml

import (
	"strconv"
	"bytes"
	"errors"
)

/* Minimalistic YAML parser
	- Each element either have a single value, or a list of values
	- Values are always strings or other elements (child nodes)
*/

func (node *Node) parse(lines [][]byte, index, depth int) error {
    var indent int
    var child *Node
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
        return node.parse(lines, index+1, depth)
    }

    /* Indentation is same, so parse for current node */
    if indent == depth {

        /* key was declared on previous line,
        means this is an element of a list
        */
        if len(key) == 0 {
            node.AddValue(value)
            /* continue parsing next line */
            return node.parse(lines, index+1, depth)
        } else {  /* create new element */
            child = &Node{ Name : key }
            child.parent = node
            node.AddChild(child)

            if len(value) == 0 {  /* expect child values on next line(s) */
                return child.parse(lines, index+1, depth+2)
            } else {  /* child is single-valued */
                child.SetValue(value)
                return node.parse(lines, index+1, depth)
            }
        }
    /* Jump back to parent node */
    } else if indent < depth {
        return node.parent.parse(lines, index, depth-2)
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

	/* seperator is last character, so we only have key */
    } else if mid+1 == end {  
		key = line[start:mid]
		
	/* both */
    } else {  
        key = line[start:mid]
        value = line[mid+2:end]
    }

    key = bytes.TrimPrefix(key, []byte("- "))
    value = bytes.TrimLeft(value, " ")
    value = bytes.TrimPrefix(value, []byte("- "))

    return start, string(key), string(value)
}

