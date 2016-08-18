#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 9
#   Find the only Pythagorean triplet, {a, b, c}, for which a + b + c = 1000.
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
    abcs = []
    for c in reversed(xrange(5, 995)):
        for b in reversed(xrange(4, 1000-c)):
            a = 1000 - b - c
            if a <= 0:
                continue
            if c**2 == a**2 + b**2:
                return a * b * c

if __name__ == "__main__":
    import cProfile
    cProfile.run("print execute()")

