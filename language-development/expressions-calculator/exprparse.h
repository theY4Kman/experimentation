#define IS_NUMERIC(c) ((c>='0')&&(c<='9'))
#define CHR_TO_INT(c) (c-'0')

/* struct factor appeasement */
struct expr;

/**
 * Semantic rules:
 *   factor := 0..9
 *   factor := ( expr )
 */
struct factor
{
    factor():exp(NULL){};
    
    int value;
    expr *exp;
};

/* e.g., '+' or '/' */
typedef char oper;

/** Semantic rules: 
 *    term := term * factor
 *    term := term / factor
 *    term := factor
 */
struct term
{
    term *trm;
    factor *fct;
    oper op;
};

/**
 * Semantic rules:
 *   expr := expr | term.t | oper
 *   expr := term.t
 */
struct expr
{
    expr():parent(NULL),exp(NULL),trm(0),op(-1){};
    
    expr *exp;
    expr *parent;
    term *trm;
    oper op;
};

/**
 * By storing the opcodes like this (from '*', ASCII code 42),
 * we can provide extremely fast lookup of standard operators.
 * This, of course, will change in the future. Hooray!
 * PREMATURE OPTIMIZATION OH YEAH!
 */
static char oper_opcodes[][4] = {
    "mul", // * (42)
    "add", // + (43)
    "\0",  // , (44)
    "sub", // - (45)
    "\0",  // . (46)
    "div", // / (47)
};

static expr *root = new expr();

int usage();
void printerr(char *msg, unsigned int line=0, unsigned int chr=0);
int main(int argc, char *argv[]);
void traverse_tree(expr *node);
