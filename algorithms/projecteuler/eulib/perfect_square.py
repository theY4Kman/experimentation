#!/usr/bin/env python
# =============================================================================
# Perfect Square
#   Functions and generators associated with perfect squares
# Copyright (C) 2010 Zach "theY4Kman" Kanzler
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

import math

# Determines if a number is a perfect square or not.
def is_perfect_square(n):
  if n < 0:
    return false
  
  tst = long(math.sqrt(n) + 0.5)
  return tst * tst == n

def perfect_square_gen():
  n = 1
  while True:
    yield n*n
    n += 1
