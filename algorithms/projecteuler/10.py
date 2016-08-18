#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 10
#   Calculate the sum of all the primes below two million.
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

from math import sqrt

def erat(n):
    # Create a list filled with ints from 2 to n, inclusive.
    numbers = range(2, n+1)
    
    length = n - 2
    limit = int(sqrt(n))
    
    j = 0
    while numbers[j] <= limit:
        i = 2
        prime_check = numbers[j]
        while i < length:
            if numbers[i] % prime_check == 0:
                numbers.pop(i)
                length -= 1
            else:
                i += 1
        j += 1
    
    return numbers

def execute():
    #return sum(erat(2000000))
    total = 2
    x = 1
    pairs = []
    while x < 2000000:
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
        total += cur
        print "%15d%15d" % (cur, total)
    return total

if __name__ == "__main__":
    import cProfile
    cProfile.run("print execute()")

