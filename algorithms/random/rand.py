#!/usr/bin/env python
# =============================================================================
# Random Number Generators
#   An implementation of the Linear Congruential Method for generating pseudo-
#   random numbers.
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

import time

class LCGenerator:
    """
    Generates random numbers based on the Linear Congruential Method. We're
    using a few rules of advice by Knuth and other arbitrary sources.
    @see: http://en.wikipedia.org/wiki/Linear_congruential_method
    """
    
    m = 65532
    """
    A constant. Should be a very large number, usually a power of 10 or 2.
    """
    
    b = 6421
    """
    Another constant. Should have one less digit than L{m}, and in the format
    `x21`, where `x` is an even number. This requirement prevents some problems,
    such as repeating patterns of numbers.
    """
    
    def __init__(self, seed=None):
        """
        @type   seed: long
        @param  seed: The seed for the generator
        @ivar   a: Holds the last generated number or seed
        """
        
        if seed is not None:
            self.a = seed
        else:
            self.a = long(time.time() * 1000)
    
    def next(self):
        self.a = (self.a * self.b + 1) % self.m
        return self.a

if __name__ == "__main__":
    # Some testing
    gen = LCGenerator()
    for i in range(25):
        print gen.next()

