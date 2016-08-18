#!/usr/bin/env python
# =============================================================================
# XChat Dbus
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
"""
A class that abstracts the dbus interface of XChat.
"""

import dbus

class XChatDbus:
    """
    A class that interfaces with XChat's dbus
    """
    def __init__(self):
        """
        Connect to xchat
        """
        bus = dbus.SessionBus()
        proxy = bus.get_object('org.xchat.service', '/org/xchat/Remote')
        remote = dbus.Interface(proxy, 'org.xchat.connection')
        path = remote.Connect ("ircbot.py",
	                   "YakBot",
	                   "A standalone Python IRC bot",
	                   __version__)
        proxy = bus.get_object('org.xchat.service', path)
        
        #: The dbus iface object to xchat
        self.xchat = dbus.Interface(proxy, 'org.xchat.plugin')
    
    def get_xchat_channels(self):
        """
        Retrieves a list of channels
        @rtype: list
        @return: A list of tuples, in the format tuple(network, channel)
        """
        channels = self.xchat.ListGet("channels")
        channels_list = []
        
        while self.xchat.ListNext(channels):
            if self.xchat.ListInt(channels, "type") != 2:
                continue
            
            chan = self.xchat.ListStr(channels, "channel")
            network = self.xchat.ListStr(channels, "server")
            channels_list.append((network, chan))
        
        return channels_list

