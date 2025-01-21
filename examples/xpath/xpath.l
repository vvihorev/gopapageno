%cut <
%preamble LexerPreallocMem

%%

IDENTIFIER [a-zA-Z0-9_\-:]+
INFO [^<]+
SPACE [ \t]
NEWLINE [\r\n]
VALUE ([^"\\]|\\.)*

%%

<{IDENTIFIER}({SPACE}+{IDENTIFIER}="{VALUE}")*>
{
	id, attributesList := getIdAndAttributesListFrom(text)
	semanticValue := xpath.NewOpenTagSemanticValue(id, attributesList, start, end)

	token.Type = OPENTAG
	token.Value = *semanticValue
}
<{IDENTIFIER}({SPACE}+{IDENTIFIER}="{VALUE}")*/>
{
	id, attributesList := getIdAndAttributesListFrom(text)
	semanticValue := xpath.NewOpenCloseTagSemanticValue(id, attributesList, start, end)

    token.Type = OPENCLOSETAG
    token.Value = *semanticValue
}
</{IDENTIFIER}>
{
	id, _ := getIdAndAttributesListFrom(text)
	semanticValue := xpath.NewCloseTagSemanticValue(id, start, end)

    token.Type = CLOSETAG
    token.Value = *semanticValue
}
({SPACE}|{NEWLINE})+
{
	return gopapageno.LexSkip
}
<?[^?]+?>
{
	return gopapageno.LexSkip
}
<![^>]+>
{
	return gopapageno.LexSkip
}
{INFO}
{
	semanticValue := xpath.NewTextSemanticValue(text, start, end)

	token.Type = TEXT
	token.Value = *semanticValue
}

%%

import (
	"github.com/giornetta/gopapageno/ext/xpath"
)

func getIdAndAttributesListFrom(text string) (id string, attributesList []*xpath.Attribute){
	var l, r int

	l = 1
	if text[1] == byte('/') {
		l = 2
	}

    for ; r < (len(text) - 1) && text[r] != byte(' '); r++ {}

	if text[r] == byte('>') {
		if text[r-1] == byte('/') {
			r--
			id = string(text[l:r])
		} else {
			id = string(text[l:r])
		}
		return
	} else {
		id = string(text[l:r])
	}

	for r < len(text) {
		for ; text[r] == byte(' ') && r < len(text); r++ {}
		l = r

		for ; text[r] != byte('=') && r < len(text); r++ {}
		key := string(text[l:r])
		r = r + 2 // skip ="
		l = r

		for ; text[r] != byte('"') && r < len(text); r++ {}
		value := string(text[l:r])
		r = r + 1 // skip "

		attribute := xpath.NewAttribute(key, value)
		attributesList = append(attributesList, attribute)

		if text[r] == byte('/') || text[r] == byte('>') {
			break
		}
	}
	return
}

func LexerPreallocMem(inputSize int, numThreads int){
}
