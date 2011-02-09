#!/usr/bin/env python
# =============================================================================
# Logic Design/Simulator
# Copyright (C) 2011 Zach "theY4Kman" Kanzler
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
import re
import ConfigParser

import pygame
pygame.init()

RGX_TUPLE = re.compile(r'\((\d+),(\d+)\)')

class InputsError(Exception):
  pass


def check_inputs(f):
  '''Decorator function that automates the checking of the inputs'''
  def check(self, inputs):
    if not isinstance(inputs, tuple):
      raise InputsError("Inputs is wrong type (expected tuple, found %d)"
          % str(type(inputs)))
      return None
    
    len_inputs = len(inputs)
    if len_inputs != self.num_inputs:
      raise InputsError("Wrong number of inputs (expected %d, found %d)"
          % (self.num_inputs, len_inputs))
      return None
    
    for i in inputs:
      if not isinstance(i, int):
        raise InputsError("Input %d is wrong type"
            % (self.num_inputs, len_inputs))
        return None
    
    return f(self, inputs)
  
  return check


class Drawable:
  '''A class that can be drawn on a surface'''
  
  def draw(self, surface, x=0, y=0):
    pass


class Connectable(Drawable):
  prev = None # Previous gate or wire
  next = None # Next gate or wire


class Wire(Connectable):
  def __init__(self, prev, next):
    self.prev = prev
    self.next = next


class Gate(Connectable):
  name = ''
  num_inputs = 0
  input_locs = None
  output_loc = None
  surface = None
  
  def _parse_tuple(self, s):
    match = RGX_TUPLE.match(s.replace(' ', ''))
    if match is None:
      return None
    
    return (int(match.group(1)), int(match.group(2)))
  
  def __str__(self):
    return '<%s Gate>' % self.name
  
  def __init__(self, window):
    self.window = window
    
    # Parse the locations of the inputs and output from the config file
    input_locs = [self._parse_tuple(self.window.images.get(
        self.__class__.__name__, 'input'+str(cfg))) for cfg in
        xrange(0,self.num_inputs)]
    output_loc = self._parse_tuple(self.window.images.get(
        self.__class__.__name__, 'output'))
  
  def get_surface(self):
    if self.surface is None:
      if self.window.gate_surfaces.has_key(self.__class__.__name__):
        self.surface = self.window.gate_surfaces[self.__class__.__name__]
      else:
        img_loc = self.window.images.get(self.__class__.__name__, 'file')
        self.surface = pygame.image.load(img_loc)
        self.window.gate_surfaces[self.__class__.__name__] = self.surface
    
    return self.surface
  
  @check_inputs
  def get_output(self, inputs):
    '''
    Does some processing and returns a boolean value indicating the output of
    the gate.
    
    @type   inputs: tuple
    @param  inputs: A list of 0's and 1's of length self.num_inputs
    '''
    pass
  
  def draw(self, surface, x=0, y=0):
    surface.blit(self.get_surface(), (x,y))
    
    size = self.get_surface().get_size()
    return pygame.Rect(x,y, x+size[0], y+size[1])


class NotGate(Gate):
  name = 'NOT'
  num_inputs = 1
  
  @check_inputs
  def get_output(self, inputs):
    return 0 if inputs[0] else 1


class OrGate(Gate):
  name = 'OR'
  num_inputs = 2
  
  @check_inputs
  def get_output(self, inputs):
    return 1 if inputs[0] or inputs[1] else 0


class NorGate(Gate):
  name = 'NOR'
  num_inputs = 2
  
  @check_inputs
  def get_output(self, inputs):
    return 1 if not(inputs[0] or inputs[1]) else 0


class XorGate(Gate):
  name = 'XOR'
  num_inputs = 2
  
  @check_inputs
  def get_output(self, inputs):
    return (inputs[0] + inputs[1]) % 2


class XnorGate(Gate):
  name = 'XNOR'
  num_inputs = 2
  
  @check_inputs
  def get_output(self, inputs):
    return (inputs[0] + inputs[1] + 1) % 2


class AndGate(Gate):
  name = 'AND'
  num_inputs = 2
  
  @check_inputs
  def get_output(self, inputs):
    return 1 if inputs[0] and inputs[1] else 0


class NandGate(Gate):
  name = 'NAND'
  num_inputs = 2
  
  @check_inputs
  def get_output(self, inputs):
    return 1 if not(inputs[0] and inputs[1]) else 0


# TODO: Images for Gnd and Vdd
class GndGen(Gate):
  name = '0'
  num_inputs = 0
  
  def get_output(self, inputs):
    return 0


class VddGen(Gate):
  name = '1'
  num_inputs = 0
  
  def get_output(self, inputs):
    return 1


class Menu(Drawable):
  '''Handles the top menu/toolbox'''
  
  height = 48
  border = 2
  pad_gates = 20
  pad_top = 2
  gate_width = 32 # TODO: Move this somewhere else?
  
  bg_color = (225, 225, 215)
  selected_color = (210, 210, 200)
  
  def __init__(self, window):
    self.window = window
    
    self.surface = pygame.Surface((self.window.width, self.height))
    
    self.draw_base()
    
    self.font = pygame.font.SysFont('', 12)
    self.gates = []
    self.gate_selected = []
    self.gate = None
    self.gate_redraw = False
    
    for x,gate in enumerate(self.window.gates.values()):
      self.gates.append(gate)
      self.draw_gate(x)
    
    self.window.reg_draw(self)
  
  def get_selected_gate_class(self):
    # Returns the class of the gate currently selected
    if self.gate is None:
      return None
    
    return self.gates[self.gate].__class__
  
  def draw_base(self):
    # BG color
    self.surface.fill(self.bg_color)
    # Bottom border
    pygame.draw.line(self.surface, (200, 200, 180), (0,self.height-self.border),
        (self.window.width,self.height-self.border), self.border)
  
  def draw_gate(self, index, selected=False):
      gate = self.gates[index]
      name = gate.name
      bg = self.bg_color if not selected else self.selected_color
      
      # Draw the gate image
      draw_x = self.pad_gates+index*(self.gate_width+self.pad_gates)
      gate.draw(self.surface, draw_x, self.pad_top)
      
      # Draw the name of the gate
      width = self.font.size(name)
      text = self.font.render(name, True, (0,0,0), bg)
      self.surface.blit(text, (draw_x+(self.gate_width-width[0])/2,
          self.pad_top+self.gate_width+self.pad_top))
  
  def onrightclick(self, x, y):
    pass
  
  def onclick(self, x, y):
    # Find the index from the x position
    idx = (x-self.pad_gates/2)/(self.gate_width+self.pad_gates)
    if idx < len(self.gates):
      self.select_gate(idx)
  
  def select_gate(self, index):
    gate = self.gates[index]
    if self.gate is not None:
      # Redraw the previously selected gate with a normal background
      bg_x = self.pad_gates/2 + self.gate*(self.gate_width+self.pad_gates)
      self.gate_selected = [pygame.Rect(bg_x, 0,
          bg_x+self.gate_width+self.pad_gates/2, self.height-self.border)]
      fill = pygame.Rect(bg_x, 0, self.gate_width+self.pad_gates,
          self.height-self.border)
      self.surface.fill(self.bg_color, fill)
      self.draw_gate(self.gate)
      
      self.gate_redraw = True
      
      if self.gates[self.gate] == gate:
        self.gate = None
        return
    
    self.gate_redraw = True
    self.gate = index
    
    # Draw selected background
    bg_x = self.pad_gates/2 + index*(self.gate_width+self.pad_gates)
    self.gate_selected.append(pygame.Rect(bg_x, 0,
        bg_x+self.gate_width+self.pad_gates/2, self.height-self.border))
    fill = pygame.Rect(bg_x, 0, self.gate_width+self.pad_gates,
        self.height-self.border)
    self.surface.fill(self.selected_color, fill)
    
    # Redraw the gate
    self.draw_gate(index, True)
    
    self.window.reg_draw(self)
  
  def draw(self, surface, x=0, y=0):
    if self.gate_redraw:
      self.gate_redraw = False
      
      # Draw the gate surfaces
      for gate in self.gate_selected:
        dest = gate[0], gate[1]
        surface.blit(self.surface, dest, gate)
      
      rect = self.gate_selected
      # Reset the invalidated rect list
      self.gate_selected = []
      
      return rect
    else:
      surface.blit(self.surface, (x,y))
      return self.surface.get_rect()


class Board(Drawable):
  '''Maintains and handles the current logic board state'''
  
  def __init__(self, window):
    self.window = window
    
    # The gate that's been selected for connection with a right-click
    self.selected_gate = None
    
    self.surface = pygame.Surface((self.window.width,
        self.window.height-self.window.menu.height))
    self.surface.fill((255,255,255))
    
    # The ground and high voltage generators. The board always starts here.
    self.gnd = GndGen(self.window)
    self.vdd = VddGen(self.window)
    
    # Draw the gnd and vdd
    self._draw_gate(self.gnd, 0, self.surface.get_height()
        - self.window.menu.gate_width)
    self._draw_gate(self.vdd, 0, 0)
    
    # A list of all the gates registered on this board. The list is in the
    # format [(Rect,Gate)], where the origin is the top left. The origin for the
    # gate surfaces is also the top left (i.e., if a gate is at x,y, its right
    # and bottom borders can be found at x+gate_width,y+gate_width)
    self.gates = [(pygame.Rect((0,0), (self.window.menu.gate_width,
        self.window.menu.gate_width)), self.gnd), (pygame.Rect((0,
        self.surface.get_height() - self.window.menu.gate_width),
        (self.window.menu.gate_width, self.window.menu.gate_width)), self.vdd)]
    
    # A list that reflects self.gates holding the Rects for redrawing.
    # Filled by default with the Gnd and Vdd rects
    self.gate_rects = [pygame.Rect((0, self.window.menu.height),
        (self.window.menu.gate_width, self.window.menu.gate_width)),
        pygame.Rect((0, self.window.height - self.window.menu.gate_width),
        (self.window.menu.gate_width, self.window.menu.gate_width))]
    
    self.window.reg_draw(self)
  
  def translate_coords(self, x, y):
    '''Translates from window coordinates to board coordinates.'''
    return (x, y-self.window.menu.height)
  
  def draw_gate(self, index):
    # Draws a gate using its index into self.gates
    x,y,gate = self.gates[index]
    self._draw_gate(gate, x, y)
  
  def _draw_gate(self, gate, x, y):
    # Draws the specified gate at the specified coordinates
    gate.draw(self.surface, x, y)
  
  def onrightclick(self, x, y):
    screen_x,screen_y = x,y
    x,y = self.translate_coords(x,y)
    
    gate_width = self.window.menu.gate_width
    new_rect = pygame.Rect(x,y, gate_width,gate_width)
    
    found_gate = None
    for rect,gate in self.gates:
      gx,gy = rect.left,rect.top
      print gx,gy,gate##############
      if gx <= x <= gx+gate_width and gy <= y <= gy+gate_width:
        # Clicked on an existing gate
        found_gate = gate
    print found_gate#############
    
    if found_gate is None:
      return
    
    if self.selected_gate is None:
      # First selection
      self.selected_gate = found_gate
    else:
      # Already saved a previously selected gate, connect the two
      self.selected_gate.next = found_gate
      found_gate.prev = self.selected_gate
  
  def onclick(self, x, y):
    screen_x,screen_y = x,y
    x,y = self.translate_coords(x,y)
    
    gate_width = self.window.menu.gate_width
    new_rect = pygame.Rect(x,y, gate_width,gate_width)
    
    overlap = False
    
    for rect,gate in self.gates:
      gx,gy = rect.left,rect.top
      if gx <= x <= gx+gate_width and gy <= y <= gy+gate_width:
        # Clicked on an existing gate
        return
      elif rect.colliderect(new_rect):
        # If a new gate were to be placed, it would overlap another
        overlap = True
    
    # Clicked on an empty space
    new_gate_cls = self.window.menu.get_selected_gate_class()
    if new_gate_cls is None or overlap is True:
      return
    
    # Create a new gate, draw it, and add it to our list of gates
    new_gate = new_gate_cls(self.window)
    
    self.gates.append((new_rect,new_gate))
    self._draw_gate(new_gate, x, y)
    
    self.gate_rects.append(pygame.Rect(screen_x,screen_y,
        screen_x+gate_width,screen_y+gate_width))
    self.window.reg_draw(self)
  
  def draw(self, surface, x=0, y=None):
    if y is None:
      y = self.window.menu.height
    
    surface.blit(self.surface, (x, y))
    
    rects = self.gate_rects
    self.gate_rects = []
    
    return rects


class Window:
  '''Handles the GUI window'''

  def __init__(self, width=640, height=480, tick=30):
    self.size = self.width,self.height = width,height
    self.screen = pygame.display.set_mode(self.size)
    pygame.display.set_caption("Yak's Logic Design")
    
    self.tick = tick
    self.clock = pygame.time.Clock()
    
    self.redraw = []
    
    self.images = ConfigParser.ConfigParser()
    self.images.read('images.ini')
    
    self.init_gates()
    
    self.menu = Menu(self)
    self.board = Board(self)
  
  def init_gates(self):
    self.gate_surfaces = {}
    
    self.gates = {}
    for cls in [NotGate, OrGate, NorGate, XorGate, XnorGate, AndGate, NandGate]:
      gate = cls(self)
      self.gates[gate.name] = gate
  
  def reg_draw(self, drawable):
    '''Add the drawable to the needs-to-be-redrawn list'''
    self.redraw.append(drawable)
  
  def run(self):
    self.screen.fill((255, 255, 255))
    pygame.display.flip()
    
    while True:
      self.clock.tick(self.tick)
      
      for event in pygame.event.get():
        if event.type == pygame.QUIT:
          sys.exit()
        
        if event.type == pygame.MOUSEBUTTONDOWN:
          # Left-click
          if event.button == 1:
            if event.pos[1] < (self.menu.height - self.menu.border):
              # Clicked the menu
              self.menu.onclick(*event.pos)
            elif event.pos[1] > self.menu.height:
              # Clicked under the menu
              self.board.onclick(*event.pos)
          
          # Right-click
          elif event.button == 3:
            if event.pos[1] < (self.menu.height - self.menu.border):
              # Clicked the menu
              self.menu.onrightclick(*event.pos)
            elif event.pos[1] > self.menu.height:
              # Clicked under the menu
              self.board.onrightclick(*event.pos)
      
      while self.redraw != []:
        # Change to one update?
        pygame.display.update(self.redraw.pop().draw(self.screen))


def main():
  window = Window()
  window.run()

if __name__ == '__main__':
  main()