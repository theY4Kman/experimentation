#!/usr/bin/env python
# =============================================================================
# Bag of Crap Watcher
#   Monitors Woot's RSS feed for the bag of crap, then sets off an alarm
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

import feedparser
import time
from os import system

while True:
    feed = feedparser.parse("http://woot.com/Blog/Feed.ashx")
    title = feed["items"][0]["title"].lower()
    if title.find("crap") != -1 or title.find("bag") != -1 or \
        title.find("agbay") != -1 or title.find("apcray") != -1:
        break
    time.sleep(5)
    
system("mplayer /home/they4kman/Music/Nine\\ Inch\\ Nails\\ -\\ Closer.mp3")

