package validator

type Document interface {
	Tags() []string
	Description() string
	Input() string
	Output() string
}

type _MarkdownDocument struct {
	input  string
	output string
}

func (m *_MarkdownDocument) Input() string {
	return m.input
}

func (m *_MarkdownDocument) Output() string {
	return m.output
}

type _undefined struct{}

var Undefined = &_undefined{}

func NewMarkdownDocument(input interface{}, output interface{}) Document {
	return nil
}
