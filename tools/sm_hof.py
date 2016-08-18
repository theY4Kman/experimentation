#!/usr/bin/env python
# =============================================================================
# SourceMod Donor Hall of Fame
# Copyright (C) 2010 Zach "theY4Kman" Kanzler <they4kman@gmail.com>
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

import re
from urllib2 import urlopen

class HallOfFameError(Exception):
  pass


RGX_DONOR = re.compile('<td style="width: 45px;">\s*<i>(?P<date>[^<]+)</i>\s*</td>\s*<td>'
  '\s*\\$(?P<amount>\d+) - (<a href="http://forums\\.alliedmods\\.net/member\\.php'
  '\\?u=(?P<forum_uid>\d+)">)?(?P<name>[^<]+)(</a>)?\s*'
  '(\\(<a href="(?P<homepage>[^"]*)">homepage</a>\\))?\s*</td>', re.M)

RGX_TOPDONOR = re.compile('<td style="width: 45px;">\s*\\$(?P<amount>\d+)'
  '\s*</td>\s*<td>\s*(<a href="http://forums\\.alliedmods\\.net/member\\.php'
  '\\?u=(?P<forum_uid>\d+)">)?(?P<name>[^<]+)(</a>)?\s*'
  '(\\(<a href="(?P<homepage>[^"]*)">homepage</a>\\))?\s*</td>', re.M)

def top_donors():
  page = urlopen('http://www.sourcemod.net/halloffame.php')
  if page is None:
    raise HallOfFameError('could not access sourcemod.net/halloffame.php')
  
  contents = page.read()
  
  mtx = RGX_TOPDONOR.findall(contents)
  if len(mtx) == 0:
    raise HallOfFameError('no top donors. Get on the tippy top of the hall of fame!'
      ' http://www.sourcemod.net/donate.php')
  
  donors = []
  for amount,crumtussels,uid,name,sapcrangle,frizcramble,homepage in mtx:
    donors.append({
      'amount': amount,
      'uid': uid,
      'homepage': homepage,
      'name': name.strip(),
    })
  
  return donors[:10]

def latest_donors():
  page = urlopen('http://www.sourcemod.net/halloffame.php')
  if page is None:
    raise HallOfFameError('could not access sourcemod.net/halloffame.php')
  
  contents = page.read()
  
  latest_idx = contents.find('Donors this month:')
  if latest_idx == -1:
    raise HallOfFameError('could not find latest donors on '
      'sourcemod.net/halloffame.php')
  
  goodbits = contents[latest_idx:]
  mtx = RGX_DONOR.findall(goodbits)
  if len(mtx) == 0:
    raise HallOfFameError('no recent donors. Get on the hall of fame!'
      ' http://www.sourcemod.net/donate.php')
  
  donors = []
  for date,amount,crumtussels,uid,name,sapcrangle,frizcramble,homepage in mtx:
    donors.append({
      'date': date,
      'amount': amount,
      'uid': uid,
      'homepage': homepage,
      'name': name.strip(),
    })
  
  return donors
