#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 5
# What is the smallest number divisible by each of the numbers 1 to 20?
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
    # We can reduce the number of divisors by finding the largest numbers from
    # 1-20 which can divide smaller numbers evenly
    nums = range(11, 20 + 1)

    i = 20
    while True:
        # The number _has_ to be even and divisible by 20 (the largest divisor)
        i += 20
        works = True
        for num in nums:
            if i % num != 0:
                works = False
                break
        if works:
            break
    
    return i

if __name__ == "__main__":
    import cProfile
    cProfile.run("print execute()")

