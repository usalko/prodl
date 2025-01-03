package ast

// String methods for different types

func (o CopyOption) String() string {
	return o.Type.ToString() + " " + o.Value
}

// String returns the unescaped column name. It must
// not be used for SQL generation. Use sql_parser.String
// instead. The Stringer conformance is for usage
// in templates.
func (node ColIdent) String() string {
	atStr := ""
	for i := NoAt; i < node.At; i++ {
		atStr += "@"
	}
	return atStr + node.Val
}

// Bool value string representation
func (val BoolVal) String() string {
	if val {
		return "true"
	}
	return "false"
}
