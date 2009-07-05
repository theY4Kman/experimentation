#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 4
#   Find the largest palindrome made from the product of two 3-digit numbers.
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
    top = 0
    for i in reversed(xrange(100, 1000)):
        for j in reversed(xrange(100, 991)):
            if j % 11 != 0: continue
            x = i * j
            strx = str(x)
            if x > top and strx == strx[::-1]:
                top = x
    
    return top

if __name__ == "__main__":
    import cProfile
    cProfile.run("print execute()")

