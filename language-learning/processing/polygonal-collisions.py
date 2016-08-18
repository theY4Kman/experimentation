#!/usr/bin/jython
# =============================================================================
# Polygonal Collisions
#   A test of polygon-to-polygon collisions using the Processing library for
#   visual effects.
# Copyright (C) 2009 Zach "theY4Kman" Kanzler
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

class Polygon:
    """Represents a polygon"""
    def __init__(self, x, y, vertices):
        """
        @type   x: int
        @param  x: x coordinate of the origin of the polygon
        @type   y: int
        @param  y: y coordinate of the origin of the polygon
        @type   vertices: iterable
        @param  vertices: The vertices of the polygon as relative coordinates
            to the origin. Should be in the form [(x1,y1), (x2,y2), ...]
        """
        self.vertices = [x for x in vertices]
        self.vertices_cnt = len(self.vertices)
    
    def 

