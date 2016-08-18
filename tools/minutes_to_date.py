#!/usr/bin/env python
# =============================================================================
# Minutes to Date
#    Accepts an amount of minutes and prints out the number of months, weeks,
#    days, hours, and minutes are in that length of time.
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

#: The amount of minutes in each measure of time
conversions = [
    ("weeks", 60 * 24 * 7),
    ("days", 60 * 24),
    ("hours", 60),
    ("minutes", 1),
]

def minutes_to_duration(minutes):
    """
    Converts an amount of minutes and prints out the number of months, weeks,
    days, hours, and minutes are in that length of time.
    
    @type   minutes: int
    @param  minutes: An amount of minutes
    @rtype: string
    @return: String representation of those minutes reduced to largest lengths
        of time, such as weeks, days, and hours.
    """
    if type(minutes) is not int:
        raise TypeError("minutes is not an int")
    
    strtime = ""
    
    for measurement,amount in conversions:
        msrmnt_cnt = minutes / amount
        
        if msrmnt_cnt == 0:
            continue
        
        minutes -= msrmnt_cnt * amount
        strtime = "%s%d %s, " % (strtime, msrmnt_cnt, msrmnt_cnt == 1 and 
            measurement[:-1] or measurement)
    
    return strtime

if __name__ == "__main__":
    if len(sys.argv) > 1:
        try:
            minutes = int(sys.argv[1])
        except ValueError:
            sys.stderr.write("The supplied amount of minutes "
                "was not a number\n")
            sys.exit(2)
        
        print minutes_to_duration(minutes)[:-2]
    else:
        sys.stderr.write("Usage: %s <minutes>\n" % sys.argv[0])
        sys.exit(2)

