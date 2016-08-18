#!/usr/bin/env python
# Project Euler - Problem 2
#   Find the sum of all the even-valued terms in the Fibonacci sequence which
#   do not exceed four million.
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

def execute():
    total = 2
    
    # Modified Fibonacci series discovery
    result = 3
    last1 = 2
    last2 = 1
    while last2 <= 4000000:
        result = last1 + last2
        if result & 1 != 1:
            total += result
        last2 = last1
        last1 = result
    
    print "Answer:", total
    return total

if __name__ == "__main__":
    import cProfile
    cProfile.run("execute()")

