#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 7
#   Find the 10001st prime.
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
    x = 1
    found = 0
    pairs = []
    while found < 10000:
        cur = None
        while cur is None:
            x += 2
            cur = x
            for prime,square in pairs:
                if cur < square:
                    break
                if cur % prime == 0:
                    cur = None
                    break
        pairs.append((cur, cur**2))
        found += 1
    return pairs[-1][0]

if __name__ == "__main__":
    import cProfile
    cProfile.run("print execute()")

