%axiom Query

%preamble ParserPreallocMem

%%

Query : Path
{};

Path : Path CHILD Step
{
} | Path PARENT Step
{
} | Path ANCESTOR Step
{
} | Path DESCENDANT Step
{
} | CHILD Step
{
} | PARENT Step
{
} | ANCESTOR Step
{
} | DESCENDANT Step
{
};

Step : Test
{
} | Test LBR OrExpr RBR
{
};

Test : IDENT 
{
} | AT IDENT
{
} | AT IDENT EQ STRING
{
} | TEXT
{
} | TEXT EQ STRING
{
};

OrExpr : AndExpr
{} | AndExpr OR AndExpr
{};

AndExpr : Factor
{} | Factor AND Factor
{};

Factor : Path
{
} | NOT Path
{
} | LPAR Predicate RPAR
{
} | NOT LPAR Predicate RPAR
{
};

%%

// var parserElementsPools []*gopapageno.Pool[xpath.Element]

// ParserPreallocMem initializes all the memory pools required by the semantic function of the parser.
func ParserPreallocMem(inputSize int, numThreads int) {
    // poolSizePerThread := 10000

    // parserElementsPools = make([]*gopapageno.Pool[xpath.Element], numThreads)
    // for i := 0; i < numThreads; i++ {
    //     parserElementsPools[i] = gopapageno.NewPool[xpath.Element](poolSizePerThread)
    // }
}
