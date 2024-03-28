package generator

import (
	"fmt"
	"os"
)

func Generate(lexerFilename string, parserFilename string, outdir string) error {
	lexerFile, err := os.Open(lexerFilename)
	if err != nil {
		return fmt.Errorf("could not open lexer description file: %w", err)
	}
	defer lexerFile.Close()

	lexerDesc, err := parseLexerDescription(lexerFile)
	if err != nil {
		return fmt.Errorf("could not parse lexer description: %w", err)
	}

	if err := lexerDesc.compile(); err != nil {
		return fmt.Errorf("could not compile lexer: %w", err)
	}

	parserFile, err := os.Open(parserFilename)
	if err != nil {
		return fmt.Errorf("could not open parser file: %w", err)
	}
	defer parserFile.Close()

	parserDesc, err := parseGrammarDescription(parserFile)
	if err != nil {
		return fmt.Errorf("could not parse parser description: %w", err)
	}

	if err := parserDesc.compile(); err != nil {
		return fmt.Errorf("could not compile parser: %w", err)
	}

	f, err := emitFile(outdir)
	if err != nil {
		return fmt.Errorf("could not generate output file %s: %w", outdir, err)
	}
	defer f.Close()

	parserDesc.emitTokens(f)
	lexerDesc.emit(f)
	parserDesc.emit(f)

	return nil
}
