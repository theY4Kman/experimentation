#!/usr/bin/env python
# =============================================================================
# Area of a Triangle
#   Given two sides and an angle or three sides, calculates the area of a
#   triangle.
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
import math

angles = {
    "A": raw_input("A = ") or None,
    "B": raw_input("B = ") or None,
    "C": raw_input("C = ") or None,
}

sides = {
    "a": raw_input("a = ") or None,
    "b": raw_input("b = ") or None,
    "c": raw_input("c = ") or None,
}

if sides["a"] and sides["b"] and angles["C"]:
    print "Area = %.2f" % ((1.0/2) * float(sides["a"]) * float(sides["b"]) *
        math.sin(math.radians(float(angles["C"]))))
elif sides["b"] and sides["c"] and angles["A"]:
    print "Area = %.2f" % ((1.0/2) * float(sides["b"]) * float(sides["c"]) *
        math.sin(math.radians(float(angles["A"]))))
elif sides["a"] and sides["c"] and angles["B"]:
    print "Area = %.2f" % ((1.0/2) * float(sides["a"]) * float(sides["c"]) *
        math.sin(math.radians(float(angles["B"]))))
elif sides["a"] and sides["b"] and sides["c"]:
    a = float(sides["a"])
    b = float(sides["b"])
    c = float(sides["c"])
    s = (1.0/2) * (a + b + c)
    print "Area = %.2f" % math.sqrt(s * (s - a) * (s - b) * (s - c))
else:
    sys.stderr.write("Unable to calculate area of triangle.\n")
    sys.exit(2)

