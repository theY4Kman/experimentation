#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include "exprparse.h"

int usage()
{
    printf("Usage: exprparse expression\n");
    return 2;
}

void printerr(char *msg, unsigned int line, unsigned int chr)
{
    printf("error:%d:%d: %s\n", line ? line : 1, chr ? chr : 0, msg);
}

int main(int argc, char *argv[])
{
    if(argc <= 1)
    {
        return usage();
    }
    
    expr *node = root;
    
    int i, j;
    for(j=1; j<argc; j++)
    {
        int len = strlen(argv[j]);
        for(i=0; i<len; i++)
        {
            char c = argv[j][i];
            if(c == ' ') continue;
            
            if(IS_NUMERIC(c))
            {
                int value = 0;
                
                /* New terms _always_ mean new expressions (read: semantic rules for `term`) */
                expr *new_node = new expr();
                term *new_term = new term();
                factor *new_fct = new factor();
                
                do
                {
                    /* Multi-digit terms (basically atoi, but increment the loop counter) */
                    new_fct->value = (new_fct->value * 10) + CHR_TO_INT(argv[j][i]);
                    i++;
                } while(IS_NUMERIC(argv[j][i]));
                i--; // we overshoot by one in the above loop
                
                new_term->fct = new_fct;
                new_node->trm = new_term;
                new_node->parent = node;
                node->exp = new_node;
                node = new_node;
            }
            else
            {
                switch(c)
                {
                default:
                    break;
                case '+':
                case '-':
                case '*':
                case '/':
                    if(node->parent == NULL)
                    {
                        printerr((char*)"unexpected operator", 0, i);
                        continue;
                    }
                    
                    node->op = c;
                }
            }
        }
    }   
    
    /* no expression means no terms means no evaluation */
    if(root->exp == NULL)
    {
        printf("error: no expression to be evaluated.\n");
        return 1;
    }
    
    /* printf setup and main code setup */
    printf("extern printf\n"\
           "segment .data\n"\
           "\toutputfmt:\tdb \"= %%d\", 10, 0\n\n"\
           "segment .text\n"\
           "\tglobal main\n"
           "main:\n");
    
    /** !!! GIANT TODO: Make tree traversal recursive !!! **/
    
    /* For the first term, we must mov it into eax. */
    printf("\tmov\teax, %d\n", root->exp->trm->fct->value);
    
    traverse_tree(root->exp);
    
    printf("\tpush\teax\n"\
           "\tpush\tdword outputfmt\n"\
           "\tcall\tprintf\n");
#ifdef __linux__
    printf("\tmov\teax, 1\n"\
	       "\tmov\tebx, 0\n"\
           "\tint\t80h\n");
#elif defined(_WIN32)
    printf("\tmov\teax, 0xf\n"\
           "\tmov\tedx, 0\n"\
           "\tint\t21h\n");
#endif
    
    return 0;
}

void traverse_tree(expr *node)
{
    while(node->exp != NULL)
    {
        /* Reuse numbers from previous expressions */
        if(node->parent != root && 
           node->parent->trm->fct->value == node->trm->fct->value)
        {/* Do nothing */}
        else if(node->parent != root &&
                node->parent->trm != NULL &&
                node->trm != NULL &&
                (node->parent->trm->fct->value == node->trm->fct->value-1 ||
                 node->parent->trm->fct->value == node->trm->fct->value+1))
        {
            printf("\t%s\tebx\n", node->exp->trm->fct->value == node->trm->fct->value-1 ? "dec" : "inc");
        }
        else
        {
            printf("\tmov\tebx, %d\n", node->exp->trm->fct->value);
        }
        
        if(node->op == -1)
        {
            printerr((char*)"no operator found for expression!");
            //return 1;
        }
        
        if(node->op == '*' || node->op == '/')
        {
            printf("\t%s\tebx\n", oper_opcodes[node->op-42]);
        }
        else
        {
            printf("\t%s\teax, ebx\n", oper_opcodes[node->op-42]);
        }
        
        traverse_tree(node->exp);
    }
}
