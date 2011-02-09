#!/usr/bin/env python
# =============================================================================
# Circle-to-circle Collision Detection
#   A test!
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

def are_colliding(c1, c2):
    u"""
    Tests if two circles are colliding.
    @type   c1: 3-tuple
    @param  c1: (x, y, r)
    @type   c2: 3-tuple
    @param  c2: (x, y, r)
    """
    dx = abs(c1[0] - c2[0])
    dy = abs(c1[1] - c2[1])
    h  = sqrt(dx**2 + dy**2)
    
    if h <= (c1[2] + c2[2]):
        return True
    else:
        return False



