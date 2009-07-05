#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 6
#   What is the difference between the sum of the squares and the square of the
#   sums?
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
    sumof = sum(xrange(1, 101))**2
    squareof = sum([x**2 for x in xrange(1, 101)])
    return sumof - squareof

if __name__ == "__main__":
    import cProfile
    cProfile.run("print execute()")
