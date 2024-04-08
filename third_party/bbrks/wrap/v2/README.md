# wrap [![Go Reference](https://pkg.go.dev/badge/github.com/bbrks/wrap/v2.svg)](https://pkg.go.dev/github.com/bbrks/wrap/v2) [![GitHub tag](https://img.shields.io/github/tag/bbrks/wrap.svg)](https://github.com/bbrks/wrap/releases) [![license](https://img.shields.io/github/license/bbrks/wrap.svg)](https://github.com/bbrks/wrap/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/bbrks/wrap)](https://goreportcard.com/report/github.com/bbrks/wrap) [![codecov](https://codecov.io/gh/bbrks/wrap/branch/master/graph/badge.svg)](https://codecov.io/gh/bbrks/wrap)

An efficient and flexible word-wrapping package for Go (golang)

## Usage

```go
var loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed vulputate quam nibh, et faucibus enim gravida vel. Integer bibendum lectus et erat semper fermentum quis a risus. Fusce dignissim tempus metus non pretium. Nunc sagittis magna nec purus porttitor mollis. Pellentesque feugiat quam eget laoreet aliquet. Donec gravida congue massa, et sollicitudin turpis lacinia a. Fusce non tortor magna. Cras vel finibus tellus."

// Wrap when lines exceed 80 chars.
fmt.Println(wrap.Wrap(loremIpsum, 80))
// Output:
// Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed vulputate quam
// nibh, et faucibus enim gravida vel. Integer bibendum lectus et erat semper
// fermentum quis a risus. Fusce dignissim tempus metus non pretium. Nunc sagittis
// magna nec purus porttitor mollis. Pellentesque feugiat quam eget laoreet
// aliquet. Donec gravida congue massa, et sollicitudin turpis lacinia a. Fusce non
// tortor magna. Cras vel finibus tellus.
```

```go
var loremIpsum = "/* Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed vulputate quam nibh, et faucibus enim gravida vel. Integer bibendum lectus et erat semper fermentum quis a risus. Fusce dignissim tempus metus non pretium. Nunc sagittis magna nec purus porttitor mollis. Pellentesque feugiat quam eget laoreet aliquet. Donec gravida congue massa, et sollicitudin turpis lacinia a. Fusce non tortor magna. Cras vel finibus tellus. */"

w := wrap.NewWrapper()

// Trim the single-line block comment symbols from each input line.
w.TrimInputPrefix = "/* "
w.TrimInputSuffix = " */"

// Prefix each new line with a single-line comment symbol.
w.OutputLinePrefix = "// "

// Wrap when lines exceed 80 chars.
fmt.Println(w.Wrap(loremIpsum, 80))
// Output:
// // Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed vulputate quam
// // nibh, et faucibus enim gravida vel. Integer bibendum lectus et erat semper
// // fermentum quis a risus. Fusce dignissim tempus metus non pretium. Nunc
// // sagittis magna nec purus porttitor mollis. Pellentesque feugiat quam eget
// // laoreet aliquet. Donec gravida congue massa, et sollicitudin turpis lacinia
// // a. Fusce non tortor magna. Cras vel finibus tellus.
```

### Advanced Usage and more examples (custom breakpoints, prefixes, suffixes, etc.)

See [godoc.org/github.com/bbrks/wrap](https://godoc.org/github.com/bbrks/wrap) for more examples using the `Wrapper` type to provide custom breakpoints, prefixes, suffixes, etc.

## Contributing

Issues, feature requests or improvements welcome!

## License
This project is licensed under the [MIT License](LICENSE).
