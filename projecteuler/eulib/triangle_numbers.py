#!/usr/bin/env python
# =============================================================================
# Triangle Numbers
#   Functions and generators associated with triangle numbers
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

def triangle_gen():
    """Generates triangle numbers"""
    triangle = 0
    next = 1
    while True:
        triangle += next
        yield triangle
        next += 1
