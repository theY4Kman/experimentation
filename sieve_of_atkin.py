#!/usr/bin/env python
# =============================================================================
# Sieve of Atkin
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

import sys

def atkin(n):
    primes = [2, 3, 5]


if __name__ == "__main__":
    print "Sieve of Atkin program, by theY4Kman"
    
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
    
    print atkin(n)
