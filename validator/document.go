package validator

type Document interface {
	Tags() []string
	Description() string
	Input() string
	Output() string
}

type _SimpleDocument struct {
	description string
	input       string
	output      string
}

func (m *_SimpleDocument) Input() string {
	return m.input
}

func (m *_SimpleDocument) Output() string {
	return m.output
}

type _undefined struct{}

var Undefined = &_undefined{}

func NewDocument(input interface{}, output interface{}) Document {
	return nil
}

func NewOpenAPI() Document {
	// todo
	return nil
}
