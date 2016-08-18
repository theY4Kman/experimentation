#!/usr/bin/env python
# =============================================================================
# Project Euler - Problem 42
#   How many common English words are triangle words?
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

import csv

def triangle_gen():
    """Generates triangle numbers"""
    triangle = 0
    next = 1
    while True:
        triangle += next
        yield triangle
        next += 1

if __name__ == '__main__':
  gen = triangle_gen()
  n = gen.next()
  tri_list = [n]
  
  # In words.txt, the highest word value is 179, so we precalculate a list of
  # triangle numbers to 179
  while n < 179:
    n = gen.next()
    tri_list.append(n)
  
  fp = open('42-words.txt', 'r')
  reader = csv.reader(fp, delimiter=',', quotechar='"')
  
  # ord(c) - ord('A') + 1 == char index
  A = ord('A')-1
  
  triangles = 0
  top = 0
  for word in reader.next():
    # Calculate the word value
    n = sum([ord(c)-A for c in word])
    
    if top < n:
      top = n
    
    if n in tri_list:
      triangles += 1
  
  print triangles, 'triangle words.'