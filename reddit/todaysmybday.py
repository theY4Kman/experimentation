#!/usr/bin/env python
# =============================================================================
# Reddit TodayIsMyRedditBday Name Permutation Generator
#   Generates many different passable reddit usernames that convey the idea
#   "today is my reddit birthday."
# Copyright (C) 2011 Zach "theY4Kman" Kanzler <they4kman@gmail.com>
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

class Word:
  def __init__(self, *strings):
    self.strings = strings
  
  def __add__(self, op):
    if op.__class__ == Phrase:
      return Phrase(op.words + [self])
    elif op.__class__ == Word:
      return Phrase([self, op])
    else:
      raise TypeError("expected Phrase or Word, got %s" % str(type(op)))
  
  def render(self, underscores=False):
    for string in self.strings:
      yield string
      if underscores:
        yield string + '_'


class Phrase:
  def __init__(self, words):
    self.words = words
  
  def __add__(self, op):
    if op.__class__ == Phrase:
      return Phrase(op.words + self.words)
    elif op.__class__ == Word:
      return Phrase(self.words + [op])
    else:
      raise TypeError("expected Phrase or Word, got %s" % str(type(op)))
  
  def render(self, underscores=False):
    levels = len(self.words)
    max_level = levels - 1
    
    gens = [word.render(underscores=True) for word in self.words]
    cur_gen = [None] * levels
    
    cur_level = 0
    while cur_level >= 0:
      try:
        cur_gen[cur_level] = gens[cur_level].next()
      except StopIteration:
        gens[cur_level] = self.words[cur_level].render(underscores)
        cur_level -= 1
        continue
      
      if cur_level == max_level:
        if not cur_gen[cur_level].endswith('_'):
          yield ''.join(cur_gen)
      else:
        cur_level += 1


tomfoolery = False
if tomfoolery:
  Its = Word('Its', 'Itz', 'ItIs', 'ItIz', 'ItBe')
  Is = Word('Is', 'Iz')
  Today = Word('Today', '2day', 'Now')
  TodayIs = Word('Todays', '2days', 'Nows')
  Birthday = Word('Birthday', 'Bday')
  BirthdayIs = Word('Birthdays', 'Bdays')
  My = Word('My', 'Me', 'Mah', 'Mi')
  Me = Word('Me')
  Reddit = Word('Reddit')
  This = Word('This', 'Dis')
  Born = Word('Born')
  On = Word('On')
  For = Word('For', '4')
else:
  Its = Word('Its', 'ItIs')
  Is = Word('Is', 'Iz')
  Today = Word('Today', '2day', 'Now')
  TodayIs = Word('Todays', '2days', 'Nows')
  Birthday = Word('Birthday', 'Bday')
  BirthdayIs = Word('Birthdays', 'Bdays')
  My = Word('My')
  Me = Word('Me')
  Reddit = Word('Reddit')
  This = Word('This')
  Born = Word('Born')
  On = Word('On')
  For = Word('For', '4')

phrases = [
  Its + My + Reddit + Birthday + Today,
  Today + Is + My + Reddit + Birthday,
  Today + Its + My + Reddit + Birthday,
  Its + My + Reddit + Birthday,
  Born + Today + On + Reddit,
  This + Is + My + Reddit + Birthday,
  Birthday + For + Reddit + Today,
  TodayIs + My + Reddit + Birthday,
  My + Reddit + BirthdayIs + Today,
  This + My + Reddit + Birthday,
  Reddit + Birthday + For + Me,
  Reddit + Birthday + For + Me + Today,
]

if __name__ == '__main__':
  with open('todaysmybday.txt', 'w') as f:
    num_phrases = 0
    for phrase in phrases:
      for sent in phrase.render(underscores=True):
        if len(sent) <= 20:
          num_phrases += 1
          f.write('%s\n' % sent)
    f.write("%d phrases" % num_phrases)
    print num_phrases
