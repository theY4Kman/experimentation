#!/usr/bin/env python
# =============================================================================
# ID3 Class
#   A small test implementation of an ID3 tag reader.
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
import struct

import getopt

class ID3:
    def __init__(self, path=None):
        if path != None:
            self.link(path)
    
    def link(self, path):
        try:
            fp = open(path, "rb")
        except IOError, error:
            raise IOError(error)
        
        header = None
        flags = 0
        version = ""
        
        try:
            # First we search for a prepended header
            header = fp.read(10)
        except IOError, error:
            raise IOError("Could not read from file: %s" % error)
        
        if header[:3] == "ID3":
            # Only open ID3v2.3 or 2.4 files
            if header[3] == '\x04' or header[3] == '\x03':
                version = "%d.%d" % (ord(header[3]), ord(header[4]))
        
        fp.close()
        self.path = path
        
        self.header = header
        self.version = version
    
    def getVersion(self):
        return self.version

if __name__ == "__main__":
    def usage():
        print """\
Usage: %s [-h,--help] [-v, --verbose level] [-f,--file] path to file"""% (sys.argv[0].rsplit('/')[-1])
        
    if len(sys.argv) < 2:
        usage()
        sys.exit(2)
    
    arguments = "hvf"
    long_arguments = ["help", "verbose=", "file"]
    
    try:
        opts, args = getopt.getopt(sys.argv[1:], arguments, long_arguments)
    except getopt.GetoptError, error:
        print str(error)
        usage()
        sys.exit(2)
    
    options = {
        "file": ' '.join(args),
        "verbose": 0,
        }
    
    for opt, arg in opts:
        if opt in ("-h", "--help"):
            usage()
            sys.exit()
        elif opt in ("-v", "--verbose"):
            options["verbose"] = int(arg) or 0
        elif opt in ("-f", "--file"):
            options["file"] = arg
        else:
            assert(False, "unhandled option.")
    
    def verbose(msg, level=0):
        if level <= options["verbose"]:
            print msg
    
    verbose("Linking file...", 1)
    mp3 = ID3(options["file"])
    
    verbose("ID3 version: ID3v%s" % mp3.getVersion(), 0)

