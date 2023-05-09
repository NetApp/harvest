package collectors

import (
	"bytes"
	"compress/gzip"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
	output := gjson.GetManyBytes(bb, "records", "num_records", "_links.next.href")

	data := output[0]
	numRecords := output[1]
	isNonIterRestCall := !data.Exists()

	if isNonIterRestCall {
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
