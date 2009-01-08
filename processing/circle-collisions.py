#!/usr/bin/jython
# =============================================================================
# Circle to Circle Collisions
#   A graphical representation of circle collisions
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

from processing.core import PApplet
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

class CircleCollisions(PApplet):
    def setup(self):
        self.size(350, 250)
        self.setup_circles()
    
    def setup_circles(self):
        self.c1 = (self.random(300)+25, self.random(300)+25, self.random(150), self.random(255))
        self.c2 = (self.random(300)+25, self.random(300)+25, self.random(150), self.random(255))
        self.colliding = are_colliding(self.c1, self.c2)
        
        if self.colliding:
            print "Colliding"
        else:
            print "Not colliding"
    
    def circle(self, x, y, radius):
        self.ellipse(x, y, radius * 2, radius * 2)
    
    def draw(self):
        self.background(255)
        self.fill(self.c1[3])
        self.circle(*self.c1[:3])
        self.fill(self.c2[3])
        self.circle(*self.c2[:3])
    
    def mousePressed(self, arg):
        self.setup_circles()

if __name__ == "__main__":
    import pawt
    pawt.test(CircleCollisions())

