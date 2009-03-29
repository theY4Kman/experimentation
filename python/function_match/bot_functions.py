#!/usr/bin/env python
# =============================================================================
# Bot Function Matcher
# Copyright (C) 2009 Zach "theY4Kman" Kanzler
# =============================================================================
#
# This program is free software; you can redistribute it and/or modify it under
# the terms of the GNU General Public License, version 3.0, as published by the
# Free Software Foundation.
# 
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
# FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
# details.
#
# You should have received a copy of the GNU General Public License along with
# this program.  If not, see <http://www.gnu.org/licenses/>.
#
# In addition, CShadowRun has permission to utilize, distribute, and do with
# this code and the above license whatever he sees fit.

import sys
import re
from pyparsing import Word, alphas, alphanums, delimitedList, Optional, \
                      Literal, Suppress, QuotedString, nums, ParseException

# Creates a grammar for function calls
LPAREN = Suppress("(")
RPAREN = Suppress(")")
fn_nameident = Word(alphas + "_", alphanums + "_")
fn_name = delimitedList(fn_nameident, ".", combine=True)
fn_arg  = QuotedString(quoteChar='"') ^ QuotedString(quoteChar="'") ^ Word(nums)
fn_args = LPAREN + Optional(delimitedList(fn_arg), default=[])("args") + RPAREN
fn_func = fn_name("name") + fn_args

#: Regular expression to match a function call
rgx_function_call = re.compile("(?P<name>\S+)\s*\((?P<args>.*)\)")
#: Contains functions available to the bot in the format "func": (func, argc)
available_funcs = {}

def botfunc(fn):
    """
    A decorator that automatically adds a function to the list of functions
    available to the bot.
    """
    available_funcs[fn.func_name] = (fn, fn.func_code.co_argcount)

@botfunc
def example(arg1, arg2):
    """
    This function will be made available to processed strings. All arguments
    passed are strings. The return value will be made into a string.
    """
    return arg1 + arg2

def rprocess_function(sfunc, args):
    """
    Processes a recursive function call.
    @type   sfunc: str
    @param  sfunc: The name of the function to call.
    @type   args: list
    @param  args: A list of strings containing the arguments to the function
        call. They will also be processed for function calls.
    @rtype: str
    @return: The return value of the function, or None on recoverable error.
    @raise  Exception: Raises the appropriate error when an error occurs.
    """
    if not available_funcs.has_key(sfunc):
        raise NameError("\"%s\" is not a valid function." % sfunc)
    
    func = available_funcs[sfunc]
    passed_argc = len(args)
    if func[1] != passed_argc:
        raise TypeError("%s() takes exactly %d argument%s (%d given)" %
            (sfunc, func[1], "" if func[1] == 1 else "s", passed_argc))
    
    # This will store the parsed arguments
    pargs = []
    for arg in args:
        match = rgx_function_call.match(arg)
        if match is None:
            pargs.append(arg)
        else:
            # TODO: parse function calls
            arg_func = parse_function(arg)
            pargs.append(str(rprocess_function(*arg_func)))
    
    return str(func[0](*pargs))

def parse_function(call):
    """
    Parses a function call.
    @type   call: str
    @param  call: String containing the function call.
    @rtype: tuple
    @return: A tuple containing the function name and list of arguments
    @raise  SyntaxError: Improperly structured call
    """
    try:
        fn = fn_func.parseString(call)
    except ParseException, err:
        synerr = SyntaxError("invalid syntax", ("<string>", err.lineno,
            err.column, err.line))
        synglobals = {
            '__name__': '<string>',
            '__file__': '<string>',
            '__exception__': synerr
        }
        exec('\n' * (err.lineno - 1) + 'raise __exception__, None', synglobals,
            synglobals)
    
    return (fn.name, fn.args)

def process_function(call):
    """
    Processes a recursive function call.
    @type   sfunc: str
    @param  sfunc: The name of the function to call.
    @type   args: list
    @param  args: A list of strings containing the arguments to the function
        call. They will also be processed for function calls.
    @rtype: str
    @return: The return value of the function, or None on recoverable error.
    @raise  Exception: Raises the appropriate error when an error occurs.
    """
    return rprocess_function(*parse_function(call))

def process_string(text):
    """
    Parses text for function calls, replacing them with their respective return
    values.
    
    >>> process_string("Hi, my name is mynick()!")
    'Hi, my name is yakbot!'
    >>> process_string("Hi, my name is reverse(mynick())")
    'Hi, my name is tobkay!'
    
    @type   text: str
    @param  text: The string to process
    @rtype: str
    @return: The processed text.
    """
    while True:
        match = rgx_function_call.search(text)
        if match is None:
            break
        
        repl = process_function(match.group(0))
        text = text[:match.start()] + repl + text[match.end():]
    
    return text

if __name__ == "__main__":
    if len(sys.argv) <= 1:
        print "USAGE:", sys.argv[0], "<text>"
        print "Processes text with inline function calls"
    else:
        print process_string(' '.join(sys.argv[1:]))

