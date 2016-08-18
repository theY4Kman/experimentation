#!/usr/bin/env python
# =============================================================================
# Sorting Base Class
#   Displays a window to visualize a sorting procedure
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

import pygtk
pygtk.require("2.0")
import gtk

class BaseSort:
    a = list()
    
    def __init__(self):
        self.window = gtk.DrawingArea()
        self.window.set_size_request(50, 50)
        print dir(self.window.get_size())
        #gtk.main()
        
        #self.sort()
    
    def configure_event(self, widget, event):
        #self.pixmap = gtk.gdk.Pixmap(widget.window, wid)
        pass
    
    def sort(self):
        while 1: continue

