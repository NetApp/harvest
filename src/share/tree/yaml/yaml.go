package yaml

import (
    "bytes"
    "errors"
    "strconv"
    "goharvest2/share/tree/node"
)

func Load(data []byte) (*node.Node, error) {
    root := node.New([]byte("Root"))
    // be tolerant against tabs, replace with 2 spaces
    return root, parse(root, bytes.Split(bytes.ReplaceAll(data, []byte("\t"), []byte("  ")), []byte("\n")), 0, 0)
}

func parse(node *node.Node, lines [][]byte, index, depth int) error {
    //fmt.Printf(" -> %d\n", index)
    // Reached end of file
    if index == len(lines) {
        return nil
    }

    // Indentation should be 2 spaces
    // Stop if invalid indentation
    if depth % 2 != 0 {
        return errors.New("inconsistent indentation, line: " + strconv.Itoa(index+1))
    }

    // parse current line
    depthNew, key, value := parseLine(lines[index])
    //fmt.Printf("  depth=%d, depthNew=%d, key=%s, value=%s\n", depth, depthNew, string(key), string(value))

    // If line is empty, jump to next line
    if len(key) == 0 && len(value) == 0 {
        return parse(node, lines, index+1, depth)
    }

    // Indentation is same, so parse for current node
    if depthNew == depth {

        child := node.NewChild(key, value)
        // no key, means key was defined on previous line
        // means value is element of list
        if len(key) == 0 {
            return parse(node, lines, index+1, depth)
        } else {
            // key with no value, means we expect a list of values on next line(s)
            if len(value) == 0 {
                return parse(child, lines, index+1, depth+2)
            } else { // single-value child
                return parse(node, lines, index+1, depth)
            }
        }
    // Jump back to parent node
    } else if depthNew < depth {
        return parse(node.GetParent(), lines, index, depth-2)
    } else { // Current depth can't be larger than previous
        return errors.New("invalid indentation, line: " + strconv.Itoa(index+1))
    }
}

func parseLine(line []byte) (int, []byte, []byte) {
    /* variables hold indices of:
    start = position of first non-whitespace character, i.e. where indentation ends
    mid = position of key-value separator (first occurance of ":")
    end = end of line or position where comment starts (#....) */
    var start, mid, end int
    var key, value []byte

    start = -1

    for end=0; end<len(line); end+=1 {
        // stop if comment starts
        if line[end] == '#' {
            break
        }
        // first non-space character
        if start == -1 && line[end] != ' ' {
            start = end
        }
        // first seperator character
        if mid == 0 && line[end] == ':' {
            mid = end
        }
    }

    if start == -1 {
        start = 0
    }

    //fmt.Printf(" <%d> start=%d, mid=%d, end=%d\n", len(line), start, mid, end)

    // no seperator, means line only has value
    if mid == 0 {
        value = line[start:end]
    // seperator is last character, means only key
    } else if mid+1 == end {
        key = line[start:mid]
    // both present
    } else {
        key = line[start:mid]
        value = line[mid+2:end]
    }
    key = bytes.TrimPrefix(key, []byte("- "))
    value = bytes.TrimPrefix(value, []byte("- "))
    return start, bytes.Trim(key, " "), bytes.Trim(value, " ")
}

func Dump(root *node.Node) ([]byte, error) {
    data := make([][]byte, 0)
    for _, n := range root.GetChildren() {
        dumpRecursive(n, &data, 0)
    }
    return bytes.Join(data, []byte("\n")), nil
}

func dumpRecursive(node *node.Node, data *[][]byte, depth int) {
    indentation := bytes.Repeat([]byte("  "), depth)
    if len(node.GetName()) != 0 && len(node.GetContent()) != 0 {
        *data = append(*data, joinAll(indentation, node.GetName(), []byte(": "), node.GetContent()))
    } else if len(node.GetName()) != 0 {
        *data = append(*data, joinAll(indentation, node.GetName(), []byte(":")))
    } else {
        *data = append(*data, joinAll(indentation, []byte("- "), node.GetContent()))
    }
    for _, child := range node.GetChildren() {
        dumpRecursive(child, data, depth+1)
    }
}

func joinAll(slices ...[]byte) []byte {
    return bytes.Join(slices, []byte(""))
}
