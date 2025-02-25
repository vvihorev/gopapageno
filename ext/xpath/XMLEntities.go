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
	posInDocument *position
	name          int8
}

func newElement(name string, attributes []*Attribute, posInDocument *position) *Element {
	encodedName, exists := queryIds[name]
	if !exists {
		panic(fmt.Sprintf("expected to find encoded value for name: %s", name))
	}
	return &Element{attributes, posInDocument, encodedName}
}

func (e *Element) position() *position {
	return e.posInDocument
}

func (e *Element) String() string {
	return fmt.Sprintf("<%v %v></%v>", e.name, e.attributes, e.name)
}

func (e *Element) SetFromExtremeTags(openTag OpenTagSemanticValue, closeTag CloseTagSemanticValue) {
	if openTag.id != closeTag.id {
		panic("Invalid Element construction")
	}
	e.name = queryIds[openTag.id]
	e.attributes = openTag.attributes
	e.posInDocument = newPosition(openTag.startPos, closeTag.endPos)
}

func (e *Element) SetFromSingleTag(openCloseTag OpenCloseTagSemanticValue) {
	e.name = queryIds[openCloseTag.id]
	e.attributes = openCloseTag.attributes
	e.posInDocument = newPosition(openCloseTag.startPos, openCloseTag.endPos)
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
