#!/usr/bin/env python
# =============================================================================
# Polar Coordinates
#   Translates polar coordinates to rectangular coordinates.
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

import re
from sys import stdout, exit
from math import pi, cos, sin

RGX_ANGLE = re.compile(r"(?P<neg>-)?(?P<numer>\d{1,2})?pi/?(?P<denom>\d)?")
def str_to_angle(string):
    try:
        angle = int(string)
        return pi / 180 * angle
    except ValueError:
        pass
    
    try:
        angle = float(string)
        return angle
    except ValueError:
        pass
    
    mtch = RGX_ANGLE.match(string)
    if mtch is None:
        raise ValueError("string is not an acceptable angle")
    
    denom = mtch.group("denom") and float(mtch.group("denom")) or 1.0
    numer = mtch.group("numer") and float(mtch.group("numer")) or 1.0
    ret = numer * pi / denom
    
    if mtch.group("neg"):
        return 2 * pi - ret
    return ret

try:
    while True:
        r = float(raw_input("r = "))
        theta = str_to_angle(raw_input("A = "))
        
        print " (%.2f, %.2f)\n" % (r * cos(theta), r * sin(theta))
except KeyboardInterrupt:
    print "\nExiting..."
    exit(2)

