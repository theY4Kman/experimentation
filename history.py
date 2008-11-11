#!/usr/bin/env python
# =============================================================================
# History
#   A test implementation of delta objects for undo/redo
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
The data for a DELTA_ADD is a single integer, which holds a position into the
user's entered string where the deltaText should be inserted.
"""
DELTA_ADD       = 0

"""
DELTA_DELETE holds two integers in a tuple, marking the substring to be removed.
The format is (start, end)
"""
DELTA_DELETE    = 1

class TextDelta:
    def __init__(self, deltaType, deltaText, deltaData, deltaPosition):
        self.type = deltaType
        self.text = deltaText
        self.data = deltaData
        
        # The position of the cursor when the delta was created
        self.cpos = deltaPosition
    
    def __str__(self):
        mystr = self.type == DELTA_ADD and "DELTA_ADD" or "DELTA_DELETE"
        mystr += " (%s): '%s'" % (str(self.data), self.text)
        return mystr

if __name__ == "__main__":
    current_delta = 0
    current_position = 0
    current_string = ""
    deltas = [TextDelta(DELTA_ADD, "", 0, current_position)]
    
    log_file = open("log", "w+")
    def log(msg):
        global log_file
        log_file.write(str(msg)+"\n")
    
    def decrement_current_position():
        global current_position
        current_position -= current_position and 1 or 0
    def increment_current_position():
        global current_position
        current_position += current_position < (len(current_string)+1) and 1 or 0
    
    def decrement_current_delta():
        global current_delta
        current_delta -= current_delta and 1 or 0
    def increment_current_delta():
        global current_delta
        current_delta += current_delta < (len(deltas)) and 1 or 0
    
    def process_delta(undo=True):
        global deltas, current_string, current_delta
        
        if not undo and current_delta >= len(deltas)-1:
            return False
        
        # If undoing, current_delta is already set at the correct delta
        # However, if redoing, we must add one.
        idelta = deltas[undo and current_delta or current_delta+1]
        
        if undo:
            # When undoing, we treat DELTA_ADD like DELTA_DELETE
            if idelta.type == DELTA_ADD:
                # Undoing addition: splice out a substring at idelta.text position
                #   and of the same length as idelta.text
                current_string = current_string[:idelta.data] + current_string[len(idelta.text)+idelta.data:]
            else:
                current_string = current_string[:idelta.data[0]] + idelta.text + current_string[idelta.data[0]:]
            decrement_current_delta()
        else:
            if idelta.type == DELTA_ADD:
                # Redoing addition: just insert idelta.text
                current_string = current_string[:idelta.data] + idelta.text  + current_string[idelta.data:]
            else:
                # Redoing removal
                current_string = current_string[:idelta.data[0]] + current_string[idelta.data[1]:]
            increment_current_delta()
        
        current_position = idelta.cpos
        
        return True
    
    import termios, fcntl, sys, os
    
    fd = sys.stdin.fileno()

    oldterm = termios.tcgetattr(fd)
    newattr = termios.tcgetattr(fd)
    newattr[3] = newattr[3] & ~termios.ICANON & ~termios.ECHO
    termios.tcsetattr(fd, termios.TCSANOW, newattr)

    oldflags = fcntl.fcntl(fd, fcntl.F_GETFL)
    fcntl.fcntl(fd, fcntl.F_SETFL, oldflags | os.O_NONBLOCK)

    try:
        sys.stdout.write("\033[s")
        while 1:
            try:
                c = sys.stdin.read(1)
                
                if c == '\033':
                    # Arrow keys
                    mv = sys.stdin.read(2)
                    
                    # Left
                    if mv[1] == 'D':
                        decrement_current_position()
                    # Right
                    elif mv[1] == 'C':
                        increment_current_position()
                elif c == '\x7f': # Backspace
                    if current_position == 0:
                        continue
                    
                    bkspc_index = current_position and current_position-1 or 0
                    
                    deltas.append(TextDelta(DELTA_DELETE, current_string[bkspc_index:bkspc_index+1],
                                  (bkspc_index, bkspc_index+1), current_position))
                    current_string = current_string[:bkspc_index] + current_string[bkspc_index+1:]
                    
                    increment_current_delta()
                    decrement_current_position()
                elif c == '\x16': # Ctrl-V, undo
                    process_delta(undo=True)
                elif c == '\x02': # Ctrl-B, redo
                    process_delta(undo=False)
                else:
                    current_string = current_string[:current_position] + c + current_string[current_position:]
                    deltas.append(TextDelta(DELTA_ADD, c, current_position, current_position))
                    
                    increment_current_delta()
                    increment_current_position()
                
                sys.stdout.write("\033[u\033[K")
                sys.stdout.write(current_string + " ")
                sys.stdout.write("\033[%dD" % (abs(len(current_string) - current_position + 1)))
            except IOError:
                pass
    finally:
        termios.tcsetattr(fd, termios.TCSAFLUSH, oldterm)
        fcntl.fcntl(fd, fcntl.F_SETFL, oldflags)
        
        # Add a newline to normalize the comamnd prompt
        print

