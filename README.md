Go PAPAGENO
========

Go PAPAGENO (PArallel PArser GENeratOr) is a parallel parser generator based on Floyd's Operator Precedence Grammars.

It generates parallel Go parsers starting from a lexer and a grammar specification.
These specification files resemble Flex and Bison ones, although with some differences.

GoPAPAGENO is able to either generate type stubs to be integrated in a Go project, or completely self-contained programs that can be used without further effort.

This work is based on [Papageno](https://github.com/PAPAGENO-devels/papageno), a C parallel parser generator.

### Authors and Contributors

 * Michele Giornetta <michelegiornetta@gmail.com> (Refactor, AOPP and C-OPP Extensions)
 * Luca Marzi <luca.marzi@mail.polimi.it> (XPath Extension)
 * Simone Guidi <simone.guidi@mail.polimi.it> (Original Version)
