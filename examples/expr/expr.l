%cut \n

LPAR    \(
RPAR    \)
PLUS    \+
TIMES   \*
DIGIT   [0-9]
SPACE   [ \t]
NEWLINE [\r\n]

%%

{LPAR}
{
    token.Type = LPAR
}
{RPAR}
{
	token.Type = RPAR
}
{TIMES}
{
	token.Type = TIMES
}
{PLUS}
{
    token.Type = PLUS
}
{DIGIT}+
{
	num, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return err
	}

	*token.Type = NUMBER
	*token.Value = num
}
{SPACE}
{
	return ErrSkip
}
{NEWLINE}
{
	return ErrSkip
}

%%
import (
	"strconv"
)