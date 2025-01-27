package xpath

import "fmt"

type node interface {
	position() *position
}

type Attribute struct {
	Key   string
	Value string
	Next *Attribute
}

func (a *Attribute) String() string {
	return fmt.Sprintf("%v=%v", a.Key, a.Value)
}

type Element struct {
	name          string
	posInDocument *position
	attribute     *Attribute
}

func newElement(name string, attribute *Attribute, posInDocument *position) *Element {
	return &Element{name, posInDocument, attribute}
}

func (e *Element) position() *position {
	return e.posInDocument
}

func (e *Element) String() string {
	return fmt.Sprintf("<%v %v></%v>", e.name, e.attribute, e.name)
}

func (e *Element) SetFromExtremeTags(openTag OpenTagSemanticValue, closeTag CloseTagSemanticValue) {
	if openTag.Id != closeTag.Id {
		panic("Invalid Element construction")
	}
	e.name = openTag.Id
	e.attribute = openTag.Attribute
	e.posInDocument = newPosition(openTag.StartPos, closeTag.EndPos)
}

func (e *Element) SetFromSingleTag(openCloseTag OpenTagSemanticValue) {
	e.name = openCloseTag.Id
	e.attribute = openCloseTag.Attribute
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
