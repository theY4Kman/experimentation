#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 28
#   Sum of the diagonals in a 1001x1001 spiral.
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

from eulib import perfect_square_gen, prime_gen

def failed_attempt_one():
  """
  I noticed that the 5x5 spiral diagonals consisted of primes and squares, so I
  tried to generate all of the primes and squares up to 1002001.
  """
  primes = prime_gen()
  squares = perfect_square_gen()
  
  # Throw out 1 from our square gen first
  total = squares.next()
  cur_square = squares.next()
  cur_prime = 0
  
  # 1001*1001 = 1002001
  while cur_square <= 1002001:
    total += cur_square
    while cur_prime < cur_square:
      cur_prime = primes.next()
      total += cur_prime
    cur_square = squares.next()
  
  print total
  

if __name__ == '__main__':
  """
  This method is simple. Spirals have four sides, and the amount of numbers
  between the corners of each side is the same. The amount starts at 0, which
  is the center. Then it goes to 1, and from there it increases by two for
  every lap it makes around the spiral.
  
  Thus, we can easily enumerate the corners of the spiral by starting with n=1,
  then adding the amount of middle numbers plus one for each for sides, then
  increasing the amount of middle numbers by two until the desired corner
  (1002001) is reached.
  
  The number of times to go around the four sides of the spiral is the radius
  of the spiral without the center (1). So for the 1001x1001 spiral, we remove
  the center to get 1000x1000, and divide 1000/2 = 500.
  """
  
  n = 1
  middle = 1
  total = 1
  
  # floor(1001/2) = 500
  for i in xrange(500):
    for x in xrange(4):
      n += middle + 1
      total += n
    middle += 2
  
  print total
