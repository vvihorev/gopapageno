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
	name          string
	attributes    []*Attribute
	posInDocument *position
}

func newElement(name string, attributes []*Attribute, posInDocument *position) *Element {
	return &Element{name, attributes, posInDocument}
}

func (e *Element) position() *position {
	return e.posInDocument
}

func (e *Element) String() string {
	return fmt.Sprintf("<%v %v></%v>", e.name, e.attributes, e.name)
}

func (e *Element) SetFromExtremeTags(openTag OpenTagSemanticValue, closeTag CloseTagSemanticValue) {
	if openTag.Id != closeTag.Id {
		panic("Invalid Element construction")
	}
	e.name = openTag.Id
	e.attributes = openTag.Attributes
	e.posInDocument = newPosition(openTag.StartPos, closeTag.EndPos)
}

func (e *Element) SetFromSingleTag(openCloseTag OpenCloseTagSemanticValue) {
	e.name = openCloseTag.Id
	e.attributes = openCloseTag.Attributes
	e.posInDocument = newPosition(openCloseTag.StartPos, openCloseTag.EndPos)
}

// Text node
type Text struct {
	data          string
	posInDocument *position
}

func newText(data string, posInDocument *position) *Text {
	return &Text{data, posInDocument}
}

func (t *Text) String() string {
	return fmt.Sprintf("Text(%q)", t.data)
}

func (t *Text) SetFromText(tsv TextSemanticValue) {
	t.data = tsv.data
	t.posInDocument = newPosition(tsv.startPos, tsv.endPos)
}

func (t *Text) position() *position {
	return t.posInDocument
}
