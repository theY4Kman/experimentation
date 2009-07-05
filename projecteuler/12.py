#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 12
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

def triangle_gen():
    """Generates triangle numbers"""
    triangle = 0
    next = 1
    while True:
        triangle += next
        yield triangle
        next += 1

prime_pairs = []
def prime_gen():
    yield 2
    x = 1
    for x,y in prime_pairs:
        yield x
    
    while True:
        cur = None
        while cur is None:
            x += 2
            cur = x
            for prime,square in prime_pairs:
                if cur < square:
                    break
                if cur % prime == 0:
                    cur = None
                    break
        prime_pairs.append((cur, cur**2))
        yield cur

def execute():
    triangles = triangle_gen()
    sq = 0
    num = 0
    divisors = 0
    
    while divisors <= 500:
        divisors = 0
        num = triangles.next()
        print num
        sq = sqrt(num)
        primes = prime_gen()
        next_prime = 0
        next_divisor = 0
        
        while next_prime < sq:
            next_prime = primes.next()
            if num % next_prime != 0:
                continue
            
            divisors += 1
            next_divisor = next_prime * 2
            limit = num / next_prime
            while next_divisor < limit:
                if num % next_divisor != 0:
                    next_divisor += next_prime
                    continue
                divisors += 1
                next_divisor += next_prime
    
    return num

if __name__ == "__main__":
    from cProfile import run
    run("print execute()")

