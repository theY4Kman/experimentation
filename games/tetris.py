#!/usr/bin/env python
# =============================================================================
# Tetris
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

import os
import sys
import random
from threading import Timer

import pygame
from pygame.locals import *

class Yaktris:
    """
    Yaktris: A Tetris clone by theY4Kman, using only CGI effects.
    """
    
    
    class Tetromino:
        """
        Represents a tetromino, storing informating about its position on the
        game board, relative positions of its individual blocks, the rendered
        image of itself, and various transformation methods.
        """
    
        class Block(pygame.sprite.Sprite):
            """
            A single, generic square (i.e. block) to compound with other blocks
            in order to create full Tetris blocks.
            """
            
            def __init__(self, color=Color(255)):
                #: The main color of the block: what appears inside the border
                self.fg_color = Color(*color)
                #: Border color. Appears around the block.
                self.border_color = self.fg_color - Color(40, 40, 40, 0)
                
                #: Contains the pygame.Surface for this block
                self.image = pygame.Surface((20, 20))
                self.image.fill(self.border_color)
                
                # Create a temp surface to blit onto our block
                center_surface = pygame.Surface((16, 16))
                center_surface.fill(self.fg_color)
                self.image.blit(center_surface, (2, 2))
                
                pygame.sprite.Sprite.__init__(self)
        
        TETROMINOES = {
            "I": (pygame.color.THECOLORS["cyan"],   # color
                  (4, 1),                           # size
                  [(0,0), (1,0), (2,0), (3,0)],     # offsets
                  1),                               # center
            "J": (pygame.color.THECOLORS["blue"],
                  (3, 2),
                  [(0,0), (1,0), (2,0), (2,1)],
                  2),
            "L": (pygame.color.THECOLORS["orange"],
                  (3, 2),
                  [(0,1), (0,0), (1,0), (2,0)],
                  1),
            "O": (pygame.color.THECOLORS["yellow"],
                  (2, 2),
                  [(0,0), (0,1), (1,0), (1,1)],
                  0),
            "S": (pygame.color.THECOLORS["green"],
                  (3, 2),
                  [(0,1), (1,1), (1,0), (2,0)],
                  1),
            "T": (pygame.color.THECOLORS["purple"],
                  (3, 2),
                  [(0,0), (1,0), (2,0), (0,1)],
                  1),
            "Z": (pygame.color.THECOLORS["red"],
                  (3, 2),
                  [(0,0), (1,0), (1,1), (2,1)],
                  2)
        }
        """
        Using the names of the tetrominoes as keys, the values of this
        dictionary contain tuples in this fashion: (color, size, offsets,
            center)
        
        `color` is obviously the color of the tetromino; it's represented as a 
        pygame.color.Color, and each can be found in the pygame.color.THECOLORS
        dictionary.
        
        `size` is a tuple containing the width and height (in blocks) of the
        tetromino, for easy creation of Surfaces. The format is (width, height).
        
        `offsets` contains a list of tuples that act as offsets from a central
        point in the tetromino. For instance, the "I" tetromino (four blocks
        stacked in one direction) would have these offsets::
                 __ __ __ __
                |__|__|__|__| [(0,0), (1,0), (2,0), (3,0)]
        
        `center` is the index into `offsets` which contains the center point of
        the tetromino. The point of this is to make the creation of the image
        much easier, because a center point will not have to be picked (very
        ambiguous) automatically.
        
        @see: http://en.wikipedia.org/wiki/Tetris#Colors_of_tetrominoes
        """
        
        def __init__(self, piece=None):
            """
            Initializes the components of the tetromino in `piece`
            
            @type   piece: char
            @param  piece: A single character which will be used as the key to
                L{self.BLOCKS} in order to get information about the block. If
                no piece is specified, the tetromino will choose a random piece
            """
            if piece is not None:
                if not self.TETROMINOES.has_key(piece):
                    raise ValueError("`piece` must be a valid piece")
                
                self.letter_piece = piece
            else:
                self.letter_piece = random.choice(self.TETROMINOES.keys())
            
            #: Holds information about this tetromino
            self.piece = self.TETROMINOES[self.letter_piece]
            
            #: The pygame.surface.Surface containing the rendered tetromino
            self.image = pygame.surface.Surface(
                (self.piece[1][0]*20, self.piece[1][1]*20))
            
            #: A single image of the individual blocks stored in this tetromino
            self.block = self.Block(self.piece[0])
            
            for offset in self.piece[2]:
                self.image.blit(self.block.image, (offset[0]*20, offset[1]*20))
    
    
    def __init__(self):
        """
        Initializes the pygame libraries, screens, various textures, and sounds
        """
        ### Initialize pygame's libraries
        pygame.init()
        
        #### Initialize our display
        self.window = pygame.display.set_mode((300, 310))
        pygame.display.set_caption("Yaktris")
        
        #: The display surface
        self.screen = pygame.display.get_surface()
        
        ### Draw our Tetris board
        self.draw_board(initial=True)
        pygame.display.flip()
        
        self.board = [[None] * 15] * 10
        """
        Creates a 10x15 two-dimensional list, portraying our tetris board.
        For example, it can be accessed by self.board[x][y]
        @type: two-dimensional list
        """
        
        ### Miscellaneous
        pygame.key.set_repeat(250, 250)
        
        #: FPS Limiter
        self.clock = pygame.time.Clock()
        
        ### Game setup
        self.active_tetromino = self.Tetromino()
        for i in range(7):
            self.screen.blit(self.Tetromino().image, (10,i*40))
    
    def tick(self):
        """
        This function runs every half-second to update various aspects of the
        board.
        @todo: FUCKING TAKE OUT THE FLIP(). USE DIRTY RECTS
        """
        pygame.display.flip()
    
    def draw_board(self, initial=False):
        """
        Draws the game board. If `initial` is True, this method draws the whole
        board. If it's False, it will only draw the part of the board that might
        be changed.
        """
        if initial:
            self.screen.fill(color.THECOLORS["black"])
            
            font = pygame.font.SysFont("Sans", size=20, bold=True)
            
            for i in range(4):
                brdr = "blue" + str(i+1)
                
                # Draw the playing field and border
                pygame.draw.lines(self.screen, color.THECOLORS[brdr], False,
                    ( # Series of points connected by lines.
                        (8-i, 0),
                        (8-i, 300+i),
                        (208+i, 300+i),
                        (208+i, 0),
                    ), 1)
                
                # Draw the "NEXT" box
                pygame.draw.rect(self.screen, color.THECOLORS[brdr],
                    Rect((220-i,10-i, 70+i*2,90+i*2)), 1)
                
                # Print "NEXT" under the "NEXT" box
                self.screen.blit(font.render("NEXT", True,
                    color.THECOLORS[brdr]),
                    (230-i, 105-i))
    
    def main(self):
        """
        Main loop. It loops. And handles events.
        """
        while True:
            self.clock.tick(60)
            
            for event in pygame.event.get():
                if event.type == QUIT:
                    sys.exit(0)
                print event

if __name__ == "__main__":
    this_game_is_my_game_this_game_is_your_game = Yaktris()
    this_game_is_my_game_this_game_is_your_game.main()

