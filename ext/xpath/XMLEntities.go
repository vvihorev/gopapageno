package xpath

import "fmt"

type node interface {
	position() *position
}

type Attribute struct {
	key   string
	value string
}

func (a *Attribute) String() string {
	return fmt.Sprintf("%v=%v", a.key, a.value)
}

func NewAttribute(key, value string) *Attribute {
	return &Attribute{key, value}
}

type Element struct {
	attributes    []*Attribute
	startPos      int
	endPos        int
	name          int8
}

// used for tests only
func newElement(name string, attributes []*Attribute, posInDocument *position) *Element {
	encodedName, exists := QueryIds[name]
	if !exists {
		panic(fmt.Sprintf("expected to find encoded value for name: %s", name))
	}
	if posInDocument == nil {
		posInDocument = &position{0, 0}
	}
	return &Element{attributes, posInDocument.start, posInDocument.end, encodedName}
}

func (e *Element) position(p *position)  {
	p.start = e.startPos
	p.end = e.endPos
}

func (e *Element) String() string {
	return fmt.Sprintf("<%v %v></%v>", e.name, e.attributes, e.name)
}

func (e *Element) SetFromExtremeTags(openTag OpenTagSemanticValue, closeTag CloseTagSemanticValue) {
	if openTag.id != closeTag.id {
		panic("Invalid Element construction")
	}
	e.name = openTag.id
	e.attributes = openTag.attributes
	e.startPos = openTag.startPos
	e.endPos = closeTag.endPos
}

func (e *Element) SetFromSingleTag(openCloseTag OpenCloseTagSemanticValue) {
	e.name = openCloseTag.id
	e.attributes = openCloseTag.attributes
	e.startPos = openCloseTag.startPos
	e.endPos = openCloseTag.endPos
}

// Text node
type Text struct {
	TextSemanticValue
}

// used only for tests
func newText(data string, posInDocument *position) *Text {
	t := Text{}
	t.data = data
	if posInDocument == nil {
		posInDocument = &position{0, 0}
	}
	t.startPos = posInDocument.start
	t.endPos = posInDocument.end
	return &t
}

func (t *Text) String() string {
	return fmt.Sprintf("Text(%q)", t.data)
}

func (t *Text) position(p *position) {
	p.start = t.startPos
	p.end = t.endPos
}
