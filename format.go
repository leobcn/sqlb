//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//
package sqlb

type FormatOptions struct {
	SeparateClauseWith string
	PrefixWith         string
}

var defaultFormatOptions = &FormatOptions{
	SeparateClauseWith: " ",
	PrefixWith:         "",
}

var defaultScanner = &sqlScanner{
	dialect: DIALECT_MYSQL,
	format:  defaultFormatOptions,
}

type ElementSizes struct {
	// The number of interface{} arguments that the element will add to the
	// slice of interface{} arguments that will eventually be constructed for
	// the Query
	ArgCount int
	// The number of bytes in the output buffer to represent this element
	BufferSize int
}

// The struct that holds information about the formatting and dialect of the
// output SQL that sqlb writes to the output buffer
type sqlScanner struct {
	dialect Dialect
	format  *FormatOptions
}

func (s *sqlScanner) scan(b []byte, args []interface{}, scannables ...Scannable) {
	curArg := 0
	bw := 0
	bw += copy(b[bw:], s.format.PrefixWith)
	for _, scannable := range scannables {
		bw += scannable.scan(s, b[bw:], args, &curArg)
	}
}

func (s *sqlScanner) size(elements ...element) *ElementSizes {
	buflen := 0
	argc := 0

	for _, el := range elements {
		argc += el.argCount()
		buflen += el.size(s)
	}
	buflen += interpolationLength(s.dialect, argc)
	buflen += len(s.format.PrefixWith)

	return &ElementSizes{
		ArgCount:   argc,
		BufferSize: buflen,
	}
}
