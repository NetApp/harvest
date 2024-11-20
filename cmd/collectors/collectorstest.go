package collectors

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/tree"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"github.com/netapp/harvest/v2/third_party/tidwall/sjson"
	"io"
	"os"
	"path/filepath"
)

// Common unit testing helpers

var (
	gsonCache = make(map[string][]gjson.Result)
)

func JSONToGson(path string, flatten bool) []gjson.Result {
	var (
		result []gjson.Result
		err    error
	)
	results, ok := gsonCache[path]
	if ok {
		return results
	}

	var reader io.Reader
	if filepath.Ext(path) == ".gz" {
		open, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer open.Close()
		reader, err = gzip.NewReader(open)
		if err != nil {
			panic(err)
		}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		reader = bytes.NewReader(data)
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, reader) //nolint:gosec
	if err != nil {
		return nil
	}
	bb := b.Bytes()
	output := gjson.ParseBytes(bb)
	data := output.Get("records")
	numRecords := output.Get("num_records")

	if !data.Exists() {
		contentJSON := `{"records":[]}`
		response, err := sjson.SetRawBytes([]byte(contentJSON), "records.-1", bb)
		if err != nil {
			panic(err)
		}
		value := gjson.GetBytes(response, "records")
		result = append(result, value)
	} else if numRecords.Int() > 0 {
		if flatten {
			result = append(result, data.Array()...)
		} else {
			result = append(result, data)
		}
	}

	gsonCache[path] = result
	return result
}

func Params(object string, path string) *node.Node {
	yml := `
schedule:
  - data: 9999h
objects:
  %s: %s
`
	yml = fmt.Sprintf(yml, object, path)
	root, err := tree.LoadYaml([]byte(yml))
	if err != nil {
		panic(err)
	}
	return root
}
