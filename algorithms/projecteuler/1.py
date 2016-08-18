#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 1
#   Add all the natural numbers below one thousand that are multiples of 3 or 5.
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
    total = 0
    for i in xrange(1, 1000):
        if i % 3 == 0 or i % 5 == 0:
            total += i
    
    print total
    return total

if __name__ == "__main__":
    import cProfile
    cProfile.run("execute()")

