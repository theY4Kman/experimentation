#!/usr/bin/env python
# =============================================================================
# Yak Trie
#   A Trie implementation in Python to be a model for a C++ version.
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

class Trie:
    class TrieNode:
        def __init__(self, key="ROOT", value=None):
            self.children = {}
            self.key = key
            self.value = value
        
        def child(self, key, value):
            child = self.__class__(key, value)
            self.children[key[0]] = child
            return child
    
    
    def __init__(self):
        self.base = self.TrieNode()
    
    def insert(self, key, value, overwrite=True):
        """
        @type   key: string
        @param  key: Unique identifier
        @param  value: Value to store for this key
        @type   overwrite: boolean
        @param  overwrite: Overwrite the value of this key if existent?
        """
        
        length = len(key)
        this = self.base
        for i in range(length):
            if this.children.has_key(key[i]):
                next = this.children[key[i]]
                if next.key == key[i:] and overwrite:
                    # Key exists, overwrite?
                    next.value = value
                    break
                
                if i == length-1:
                    # Key already exists
                    if len(next.key) > 1:
                        # We need to create a new branch
                        next.child(next.key[1:], next.value)
                        next.key = next.key[0]
                        next.value = value
                    elif overwrite:
                        # No need for a new branch
                        next.value = value
                    break
                
                this = next
            else:
                # Create a new child
                this.child(key[i:], value)
                break
    
    def __str__(self):
        return self.tostr(self.base)
    
    limit = 10
    cur = 0
    def tostr(self, node, indent=0):
        nodestr = (" " * indent) + node.key + " => " + str(node.value) + "\n"
        
        for child in node.children.itervalues():
            if self.cur > self.limit:
                break
            
            self.cur += 1
            nodestr += self.tostr(child, indent+1)
        
        return nodestr

