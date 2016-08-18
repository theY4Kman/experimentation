#!/usr/bin/env python
# =============================================================================
# Master Server Querier
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

import socket
from struct import pack, unpack

class MasterServerQuery:
    """
    Queries the Valve Master Server for game servers.
    """
    
    #: Contains some of the master servers that can be queried
    MASTER_SERVERS = {
        "source": ("hl2master.steampowered.com", 27011),
        "goldsrc": ("hl1master.steampowered.com", 27010)
    }
    
    REPLY_HEADER = "\xff\xff\xff\xff\x66\x0A"
    SERVERLIST_REQUEST = '1'
    
    def __init__(self, master=MASTER_SERVERS["source"]):
        """
        Creates the socket and connects to the supplied master server
        
        @type   master: 2-tuple
        @param  master: (host, port) master server to connect to
        """
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.sock.connect(master)
    
    def close(self):
        self.sock.close()
    
    def _buildQuery(self, region, startip, filters):
        packet = self.SERVERLIST_REQUEST
        packet += region
        packet += "%s\x00" % (startip)
        packet += "%s\x00" % (filters)
        return packet
    
    def get_servers(self, limit=256, region=0xFF, dedicated=False, secure=False,
        gamedir=None, map=None, linux=False, empty=False, full=False,
        proxy=False, napp=None):
        """
        A blocking method to retrieve a list of servers.
        
        @type   limit: int
        @param  limit: If `limit` amount of servers are retrieved, stops
            retrieving more and returns the list of servers.
        @see L{retrieve_servers}
        """
        pass
    
    def retrieve_servers(self, callback, region='\xFF', dedicated=False,
        secure=False, gamedir=None, map=None, linux=False, empty=False,
        full=False, proxy=False, napp=None):
        """
        Retrieves game server IPs from the master server.
        
        @type   callback: callable
        @param  callback: A function that will be called every time a server is
            retrieved. The calling format is callback((ip, port)) where `ip` is
            a string and `port` is an integer. Return False in the callback to
            stop retrieval of servers.
        @type   region: char
        @param  region: The region of the world to find servers in.
            See http://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol#Region_codes
        @type   dedicated: bool
        @param  dedicated: Retrieve only dedicated servers
        @type   secure: bool
        @param  secure: Retrieve only servers using anti-cheat technology
            (VAC, but potentially others as well)
        @type   gamedir: str
        @param  gamedir: Retrieve only servers running the specified mod
            (e.g. "cstrike")
        @type   map: str
        @param  map: Retrieve only servers running the specified map
            (e.g. "cs_italy")
        @type   linux: bool
        @param  linux: Retrieve only servers running on a Linux platform
        @type   empty: bool
        @param  empty: Retrieve only servers that are _not_ empty
        @type   full: bool
        @param  full: Retrieve only servers that are _not_ full
        @type   proxy: bool
        @param  proxy: Retrieve only servers that are spectator proxies
            (e.g. SourceTV)
        @type   napp: int
        @param  napp: Retrieve only servers that are NOT running game [napp]
            (This was introduced to block Left 4 Dead games from the Steam
            Server Browser) 
        """
        
        filter = ""
        if dedicated: filter += r"\type\d"
        if secure: filter += r"\secure\1"
        if linux: filter += r"\linux\1"
        if empty: filter += r"\empty\1"
        if full: filter += r"\full\1"
        if proxy: filter += r"\proxy\1"
        if gamedir is not None:
            filter += r"\gamedir\%s" % str(gamedir)
        if map is not None:
            filter += r"\map\%s" % str(map)
        if napp is not None:
            filter += r"\napp\%d" % napp
        
        packet = self._buildQuery(region, "0.0.0.0:0", filter)
        self.sock.send(packet)
        
        buf = self.sock.recv(6)
        if buf != self.REPLY_HEADER:
            return
        
        off = 0
        while True:
            buf = self.sock.recv(6)
            
            reply = unpack("BBBB!H", buf, off)
            ip = ".".join(str(x) for x in reply[:3])
            port = reply[4]
            print ip, port
            
            if callback((ip, port)) is False:
                return


if __name__ == "__main__":
    def strtohex(string):
        pck = unpack(str(len(string))+"B", string)
        return " ".join(("%.2x" % x) for x in pck)
    
    def servers(reply):
        print reply[0] + ":" + str(reply[1])
    
    msq = MasterServerQuery()
    msq.retrieve_servers(servers)

