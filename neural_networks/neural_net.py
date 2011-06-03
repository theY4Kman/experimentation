#!/usr/bin/env python
# =============================================================================
# A neural network
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

"""
This neural network is very simple. Each column has its own pool of Neurons. A
Neuron is created for every unique name in the database. (XXX: should this be
for each unique name in a column?) Each Neuron's name is used as a key in a
dictionary, whose value is a list of all the Neurons it excites, beginning with
itself. The other Neurons in a Neuron's pool are inhibitors.

Neuron's have a weight associated with them. If they are excited with a value
that matches or exceeds the Neuron's activation value, the Neuron's output
becomes its weight (with some decay).
"""

# Highest activation value of a neuron
max_activation = 1.0
# Lowest activation value of a neuron
min_activation = -0.2
# Lowest threshold value needed to produce an output
threshold = 0.0
# Amount to decay a Neuron's weight upon firing
decay = 0.1
# The weight of a Neuron at rest
rest = -0.2

# Other decays
alpha = 0.1
beta = 0.1
gamma = 0.4

neurons = []
pools = []
neurons_name = {}

class Neuron(object):
  def __init__(self, name, pool):
    self.name = name
    self.pool = pool
    
    self.reset()
  
  def commitnewact(self):
    """Sets the Neuron's activation to the value saved in self.newweight,
    computed in self.computenewact() in a previous iteration."""
    self.setactivation(self.newweight if hasattr(self, "newweight") else rest)
  
  def computenewact(self):
    """Computes the new weight of the neuron and saves it to self.newweight"""
    a = self.activation
    excite = sum([neuron.output for neuron in neurons_name[self.name][1:]])
    inhibit = sum([neuron.output for neuron in self.pool if neuron is not self]) - self.output
    
    netinput = alpha*excite - beta*inhibit + gamma*self.weight
    magnitude = (max_activation - a) if netinput > 0 else (a - min_activation)
    a = magnitude*netinput - decay*(a - rest) + a
    
    # newweight will have a value between max_activation and min_activation
    self.newweight = max(min(a, max_activation), min_activation)
  
  def reset(self):
    self.setweight(0.0)
    self.setactivation()
  
  def setactivation(self, activation=rest):
    self.activation = activation
    self.output = max(threshold, activation)
  
  def setweight(self, weight=1.0):
    self.weight = weight
    

def touch(names, weight=1.0):
  for name in names.split():
    neurons_name[name][0].setweight(weight)


def load(s):
  """Loads in a database string and parses it into a neural network. Each line
  should be a new row, and each column should be separated by whitespace."""
  
  global pools
  global neurons
  
  pools = []
  neurons = []
  
  for line in s.splitlines():
    last_neuron = len(neurons)
    for pool_idx,name in enumerate(line.split()):
      # Create a new neuron pool for any new column
      if len(pools) <= pool_idx:
        pools.append([])
      
      if name not in neurons_name:
        neuron = Neuron(name, pools[pool_idx])
        pools[pool_idx].append(neuron)
        neurons_name[name] = [neuron]
        neurons.append(neuron)
      else:
        neuron = neurons_name[name][0]
      
      if pool_idx > 0:
        if neuron not in neurons_name[neurons[last_neuron].name][1:]:
          neurons_name[neurons[last_neuron].name].append(neuron)
        if neurons[last_neuron] not in neurons_name[name][1:]:
          neurons_name[name].append(neurons[last_neuron])


def print_pools():
  global pools
  
  for pool in pools:
    for x,neuron in enumerate(pool):
      print "%8s%5.2f" % (neuron.name + ':', neuron.activation),
      if x % 4 == 3:
        print
    print


def run(cycles=100):
  global neurons
  global pools
  
  for cycle in xrange(cycles):
    for neuron in neurons:
      neuron.computenewact()
    for neuron in neurons:
      neuron.commitnewact()
  
  print_pools()

SampleFile = """
Art         Jets        40      jh      sing    pusher
Al          Jets        30      jh      mar     burglar
Sam         Jets        20      col     sing    bookie
Clyde       Jets        40      jh      sing    bookie
Mike        Jets        30      jh      sing    bookie
Jim         Jets        20      jh      div     burglar
Greg        Jets        20      hs      mar     pusher
John        Jets        20      jh      mar     burglar
Doug        Jets        30      hs      sing    bookie
Lance       Jets        20      jh      mar     burglar
George      Jets        20      jh      div     burglar
Pete        Jets        20      hs      sing    bookie
Fred        Jets        20      hs      sing    pusher
Gene        Jets        20      col     sing    pusher
Ralph       Jets        30      jh      sing    pusher

Phil        Sharks      30      col     mar     pusher
Ike         Sharks      30      jh      sing    bookie
Nick        Sharks      30      hs      sing    pusher
Don         Sharks      30      col     mar     burglar
Ned         Sharks      30      col     mar     bookie
Karl        Sharks      40      hs      mar     bookie
Ken         Sharks      20      hs      sing    burglar
Earl        Sharks      40      hs      mar     burglar
Rick        Sharks      30      hs      div     burglar
Ol          Sharks      30      col     mar     pusher
Neal        Sharks      30      hs      sing    bookie
Dave        Sharks      30      hs      div     pusher
"""

if __name__ == "__main__":
  load(SampleFile)
  touch("Sharks 40 mar")
  run()
