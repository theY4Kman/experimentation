#!/usr/bin/env python
# =============================================================================
# Sieve of Eratosthenes
# Copyright (C) 2008 Zach "theY4Kman" Kanzler
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
    
    # Remove all multiples of 2
    i = 0
    if numbers[i] == 2:
        i = 1
    
    length = len(numbers)
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

if __name__ == "__main__":
    print "Sieve of Eratosthenes program, by theY4Kman"
    
    import sys
    n = 120
    if len(sys.argv) < 2:
        print "No upper limit supplied, using", n
    else:
        try:
            user_n = int(sys.argv[1])
            if user_n <= 2:
                print "Upper limit too small, using", n
            else:
                n = user_n
        except ValueError:
            print "Supplied upper limit is not an integer, using", n
    
    print erat(n)
