#!/usr/bin/env python
# =============================================================================
# Prime Numbers
#   Functions and generators for dealing with prime numbers
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

def prime_gen():
  prime_pairs = []
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
