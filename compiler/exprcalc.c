/**
 * =============================================================================
 * Expression Compiler
 *   Reads in arithmetic expressions and compiles them into NASM assembly code.
 * Copyright (C) 2008 Zach "theY4Kman" Kanzler
 * =============================================================================
 *
 * This program is free software; you can redistribute it and/or modify it under
 * the terms of the GNU General Public License, version 3.0, as published by the
 * Free Software Foundation.
 * 
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 * FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
 * details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program.  If not, see <http://www.gnu.org/licenses/>.
 */


#include <stdio.h>
#include <string.h>

#define IS_NUMERIC(c) ((c>='0')&&(c<='9'))
#define STR_IS_EMPTY(str) (str[0] == '\0')

int usage()
{
    printf("exprcalc postfix-expression\n");
    
    return 2;
}

int main(int argc, char *argv[])
{
    if(argc <= 1)
    {
        return usage();
    }
    
    /* printf setup and main code setup */
    printf("extern printf\n"\
           "segment .data\n"\
           "\toutputfmt:\tdb \"= %%d\", 10, 0\n\n"\
           "segment .text\n"\
           "\tglobal main\n"
           "main:\n");
    
    int len = strlen(argv[1]);
    int i;
    for(i=0; i<len; i++)
    {
        char c = argv[1][i];
        if(IS_NUMERIC(c))
        {
            printf("\tpush\tdword %c\n", c);
        }
        else
        {
            char op[4];
            switch(c)
            {
            case '+':
                strcpy(op, "add");
                break;
            case '-':
                strcpy(op, "sub");
                break;
            case '*':
                strcpy(op, "mul");
                break;
            case '/':
                strcpy(op, "div");
                break;
            default:
                op[0] = '\0';
                break;
            }
            
            if(!STR_IS_EMPTY(op))
            {
                /* `pop` expression operand 1 to eax and operand 2 to ebx. */
                printf("\tpop\tebx\n"\
                       "\tpop\teax\n"\
                       "\t%s\teax, ebx\n"\
                       "\tpush\teax\n", op);
            }
        }
    }
    
    printf("\tpush\tdword outputfmt\n"\
           "\tcall\tprintf\n"\
           "\tmov\teax, 1\n"\
	       "\tmov\tebx, 0\n"\
           "\tint\t80h\n");
    
    return 0;
}
