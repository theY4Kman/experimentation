#!/usr/bin/env python
# =============================================================================
# Unit Circle
#   Calculates the basic unit circle.
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

import math
from math import pi

unitcircle = {}

for denom in [6, 4]:
    q = {}
    for quad in range(1, 5):
        q[quad] = (denom * 2) / 4 * quad
    
    for numer in range(1, denom*2):
        rad = numer * pi / denom
        unitcircle[rad] = (
            ((q[1] < numer < q[3]) and -1 or 1) * math.cos(rad),
            ((q[2] < numer < q[4]) and -1 or 1) * math.sin(rad),
        )

if __name__ == "__main__":
    print unitcircle

