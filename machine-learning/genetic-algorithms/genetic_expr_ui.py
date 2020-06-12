import random
import signal
from bisect import bisect
from collections import deque
from dataclasses import dataclass, field
from enum import Enum, auto
from threading import Thread

from tkinter import *
from tkinter import font
from tkinter.simpledialog import askinteger
from typing import List, Optional

from genetic_expr2 import Simulation, Chromosome


MAX_DIGITS = 3
ALLOW_STANDALONE_ZEROES = True


class Type(Enum):
    UNKNOWN = -1
    NUMBER = auto()
    OPERATOR = auto()

    @classmethod
    def of_char(cls, c: Optional[str]) -> Optional['Type']:
        if c is None:
            return Type.UNKNOWN
        elif c.isnumeric():
            return Type.NUMBER
        elif c in '+-*/':
            return Type.OPERATOR
        else:
            return Type.UNKNOWN


class Char(str):
    index: int
    type: Type = None

    def __new__(cls, s, index: int) -> 'Char':
        instance = super().__new__(cls, s)
        instance.index = index
        return instance


@dataclass
class Token:
    type: Type
    chars: List[Char] = field(default_factory=list)

    def __add__(self, c: Char):
        c.type = self.type
        self.chars.append(c)
        return self


class UIChromosome(Chromosome):
    COLOR_VALID = 'green'
    COLOR_INVALID = 'yellow'
    COLOR_UNKNOWN = 'red'

    GENE_SIZE = 5  # Number of binary digits in a gene

    GENES = {
        '00000': None,
        '01111': None,

        '00001': '+',
        '01000': '-',

        '00101': '*',
        '01010': '/',

        '00010': '1',
        '00100': '9',

        '00011': '3',
        '00110': '5',
        '01100': '7',

        '01001': '0',

        '01011': '2',
        '01101': '4',

        '00111': '6',
        '01110': '8',
    }

    GENE_VALUE_BITS = {
        bits: value
        for bits, value in sorted(GENES.items())
        if value is not None
    }

    def __init__(self, gene_string, solution, *,
                 decode_max_digits: int = MAX_DIGITS,
                 decode_allow_standalone_zeroes: bool = ALLOW_STANDALONE_ZEROES):
        self.decode_max_digits = decode_max_digits
        self.decode_allow_standalone_zeroes = decode_allow_standalone_zeroes

        self.gene_colors = []
        super(UIChromosome, self).__init__(gene_string, solution)

        self.evaluated_str = '?'
        if self.evaluated:
            self.evaluated_str = f'{self.evaluated: .2f}'

        self.decoded_str = '?'
        if self.decoded:
            self.decoded_str = self.decoded.replace('*', '×').replace('/', '÷')

    def decode_gene(self, gene_value, *, pretty: bool = False):
        decoded = super().decode_gene(gene_value)

        if pretty:
            if decoded == '*':
                return '×'
            elif decoded == '/':
                return '÷'

        return decoded

    def decode(self):
        unknown = []
        tokens = []
        gene_exprs = [self.decode_gene(gene) for gene in self.genes]
        for i, expr in enumerate(gene_exprs):
            c = Char(expr, i)
            toktype = Type.of_char(c)
            if toktype == Type.UNKNOWN:
                unknown.append(c)
                continue

            if not tokens or toktype == Type.OPERATOR or toktype != tokens[-1].type:
                tokens.append(Token(type=toktype))
            tokens[-1] += c

        invalid = []
        valid = []
        while tokens:
            tok = tokens.pop(0)
            if not tok.chars:
                continue

            peek = tokens and tokens[0]
            past = valid and valid[-1]

            if tok.type == Type.NUMBER:
                # Remove leading zeroes
                while len(tok.chars) > 1 and tok.chars[0] == '0':
                    invalid.append(tok.chars.pop(0))

                # Truncate to max digits
                while len(tok.chars) > self.decode_max_digits:
                    invalid.append(tok.chars.pop())

                valid += tok.chars

            elif tok.type == Type.OPERATOR:
                op = tok.chars[0]

                if (
                    # Allow unary + or - (allow them whenever followed by a number)
                    (op in '+-' and (peek and peek.type == Type.NUMBER))

                    # Allow other ops if preceded and followed by a number
                    or (past and past.type == Type.NUMBER and peek and peek.type == Type.NUMBER)
                ):
                    valid.append(op)
                else:
                    invalid.append(op)

        color_indices = [
            (c.index, color)
            for color, chars in [(self.COLOR_UNKNOWN, unknown),
                                 (self.COLOR_INVALID, invalid),
                                 (self.COLOR_VALID, valid)]
            for c in chars
        ]
        self.gene_colors = [color for index, color in sorted(color_indices)]

        return ''.join(valid)

    def get_value_color(self) -> str:
        red         = (1.0, 0.0, 0.0)
        orange      = (1.0, 0.5, 0.0)
        yellow      = (1.0, 1.0, 0.0)
        dark_green  = (0.0, 0.7, 0.0)
        green       = (0.0, 1.0, 0.0)

        steps = [
            (0.00, red),
            (0.10, orange),
            (0.50, yellow),
            (0.96, dark_green),
            (0.97, green),
        ]
        thresholds, colors = zip(*steps)

        to_index = bisect(thresholds, self.fitness)
        if to_index >= len(colors):
            amplitudes = colors[-1]
        elif to_index == 0:
            amplitudes = colors[0]
        else:
            from_index = to_index - 1
            from_threshold, from_color = steps[from_index]
            to_threshold, to_color = steps[to_index]

            distance = (self.fitness - from_threshold) / (to_threshold - from_threshold)

            distance /= 0.5
            amplitudes = tuple(
                min(1, max(0, from_comp + distance * (to_comp - from_comp)))
                for from_comp, to_comp in zip(from_color, to_color)
            )

        components = tuple(
            round(amplitude * 255)
            for amplitude in amplitudes
        )
        return '#%02x%02x%02x' % components


class UISimulation(Simulation):
    chromosome_class = UIChromosome


class GeneticExprUI:
    FRAMES_PER_SECOND = 60
    MILLISECONDS_PER_FRAME = int(1000 / FRAMES_PER_SECOND)

    def __init__(self, simulation: UISimulation, *, population_size: int = 30):
        self.sim = simulation
        self.population_size = population_size

        self.tk = Tk()
        self.tk.minsize(1175, 250)
        self.tk.title('Genetic Expressions')
        self.tk.configure(background='black')

        self.font_chromosome = font.Font(family='monospace', size=9)
        self.font_chromosome_evaluated = font.Font(family='monospace', size=10)


        self.button_frame = Frame(self.tk, relief=FLAT, bg='black')

        self.pause_button = Button(self.button_frame, text='Pause')
        self.pause_button.bind('<Button-1>', self.on_pause_button_pressed)
        self.pause_button.pack(side=LEFT)
        self.tk.bind('<space>', self.on_pause_button_pressed)

        self.step_button = Button(self.button_frame, text='⏽⏵︎')
        self.step_button.bind('<Button-1>', self.on_step_button_pressed)
        self.step_button.pack(side=LEFT)
        self.tk.bind('<Right>', self.on_step_button_pressed)

        self.toggle_decoded_button = Button(self.button_frame, text='Show decoded︎')
        self.toggle_decoded_button.bind('<Button-1>', self.on_toggle_decoded_button_pressed)
        self.toggle_decoded_button.pack(side=LEFT)
        self.tk.bind('d', self.on_toggle_decoded_button_pressed)

        self.new_button = Button(self.button_frame, text='New Simulation')
        self.new_button.bind('<Button-1>', self.on_new_simulation_button_pressed)
        self.new_button.pack(side=LEFT)
        self.tk.bind('<Return>', self.on_new_simulation_button_pressed)

        self.target_button = Button(self.button_frame, text='Enter Target Manually')
        self.target_button.bind('<Button-1>', self.on_enter_new_target_button_pressed)
        self.target_button.pack(side=LEFT)

        self.pop_size_button = Button(self.button_frame, text='Change Population Size')
        self.pop_size_button.bind('<Button-1>', self.on_change_pop_size_button_pressed)
        self.pop_size_button.pack(side=LEFT)

        self.restart_button = Button(self.button_frame, text='Restart')
        self.restart_button.bind('<Button-1>', self.on_restart_button_pressed)
        self.restart_button.pack(side=LEFT)

        self.button_frame.pack()

        self.canvas = Canvas(self.tk, height=700)
        self.canvas.configure(background='black')
        self.canvas.pack(fill=BOTH)

        self.iteration = 0
        self.solution_chromosome = None
        self.drawn_solution = False
        self._paused = False
        self.paused = False
        self._show_decoded = False

        self._value_pads = deque((10,), maxlen=25)

        self.tk.update()
        self._resize()

    def run(self):
        self.tk.after(self.MILLISECONDS_PER_FRAME, self.draw)
        self.iteration = 0
        self.tk.mainloop()

    def draw(self):
        try:
            self._draw()
        except KeyboardInterrupt:
            self.tk.quit()

    def _draw(self):
        self._iterate()
        self.tk.after(self.MILLISECONDS_PER_FRAME, self.draw)

    def _iterate(self, *, step: bool = False):
        if not self.solution_chromosome or not self.drawn_solution:
            self._redraw()
            if self.solution_chromosome:
                self.paused = True
                self.drawn_solution = True

        if (not self.paused or step) and not self.solution_chromosome:
            self.iteration += 1
            self.sim.iteration = self.iteration
            self.solution_chromosome = self.sim.step()

            max_eval_len = max(len(chromosome.evaluated_str)
                               for chromosome in self.sim.population)
            max_eval_len = max(max_eval_len, 10)
            self._value_pads.append(max_eval_len)

        self.tk.update()

    def _redraw(self):
        self.canvas.delete(ALL)
        self.chromosome_ids = self._draw_lines(20, 50)
        self._draw_top_status(10, 5)

    def stop(self):
        self.tk.quit()
        self.tk.update()

    def _resize(self):
        pop_minheight = 15 * self.sim.population_size
        status_height = 50
        canvas_height = pop_minheight + status_height

        button_frame_height = self.button_frame.winfo_height()

        bottom_margin = 20
        height = canvas_height + button_frame_height + bottom_margin

        gene_width = self.font_chromosome.measure('0') * self.sim.chromosome_class.GENE_SIZE
        gene_space_width = self.font_chromosome.measure(' ')
        num_genes = self.sim.chromosome_size
        chromosome_width = gene_width * num_genes + gene_space_width * (num_genes - 1)

        evaluated_width = 300  # heuristic
        side_margin = 20

        width = chromosome_width + evaluated_width + 2 * side_margin

        self.canvas.config(height=canvas_height)
        self.tk.geometry(f'{width}x{height}')

    @property
    def paused(self):
        return self._paused

    @paused.setter
    def paused(self, is_paused: bool):
        self._paused = is_paused
        self.pause_button.config(text='⏵' if is_paused else '⏸')
        self.step_button.config(state=NORMAL if is_paused else DISABLED)

    @property
    def show_decoded(self):
        return self._show_decoded

    @show_decoded.setter
    def show_decoded(self, show_decoded: bool):
        self._show_decoded = show_decoded
        self.toggle_decoded_button.config(text='Show genes' if show_decoded else 'Show decoded')

    def _draw_lines(self, x=0, y=0):
        value_pad = max(self._value_pads)

        ids = []
        for i, chromosome in enumerate(self.sim.population):
            ids += self._draw_line(x, y + i*15, chromosome, value_pad=value_pad)
        return ids

    def _draw_line(self, x, y, chromosome: UIChromosome, *, value_pad: int = 10):
        ids = []
        orig_bbox = bbox = (x, y, x-5, y)
        for gene_string, gene_color in zip(chromosome.genes, chromosome.gene_colors):
            if self._show_decoded:
                text = chromosome.decode_gene(gene_string, pretty=True) or '?'
                text = (text * chromosome.GENE_SIZE)[:chromosome.GENE_SIZE]
            else:
                text = gene_string

            text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                              text=text, anchor=NW,
                                              fill=gene_color, font=self.font_chromosome)
            bbox = self.canvas.bbox(text_id)
            ids.append(text_id)

        if chromosome.is_solution:
            self.canvas.create_rectangle(
                (orig_bbox[0] - 5, y, bbox[2] + 4, bbox[3] - 2),
                fill='', outline='white')

        # Draw expression
        value_str = f'{chromosome.evaluated_str:>{value_pad}}'

        value_color = chromosome.get_value_color()
        static_color = 'white'

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text='=', anchor=NW,
                                          fill=static_color, font=self.font_chromosome_evaluated)
        bbox = self.canvas.bbox(text_id)

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text=value_str, anchor=NW,
                                          fill=value_color, font=self.font_chromosome_evaluated)
        bbox = self.canvas.bbox(text_id)

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text='=', anchor=NW,
                                          fill=static_color, font=self.font_chromosome_evaluated)
        bbox = self.canvas.bbox(text_id)

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text=chromosome.decoded_str, anchor=NW,
                                          fill=value_color, font=self.font_chromosome_evaluated)
        bbox = self.canvas.bbox(text_id)

        return ids

    def _draw_top_status(self, x=0, y=0):
        sol_bbox = self._draw_solution(x, y)
        equals_bbox = self._draw_equals_sign(sol_bbox[2], y)

        right_x = self.tk.winfo_width()
        self._draw_iteration(right_x - 20, y)

        middle_y = float(equals_bbox[1] + equals_bbox[3]) / 2

        if self.solution_chromosome:
            top_chromosome = self.solution_chromosome
        else:
            top_chromosome = sorted(self.sim.population, key=lambda c: c.fitness, reverse=True)[0]

        self._draw_top_chromosome(top_chromosome, equals_bbox[2], middle_y)

    def _draw_solution(self, x=0, y=0):
        text_id = self.canvas.create_text((x, y), text=str(self.sim.solution),
                                          anchor=NW, font='monospace 30',
                                          fill='white')
        return self.canvas.bbox(text_id)

    def _draw_equals_sign(self, x=0, y=0):
        symbol = '=' if self.solution_chromosome else '≠'
        text_id = self.canvas.create_text((x, y), text=symbol,
                                          font='monospace 30', fill='white',
                                          anchor=NW)
        return self.canvas.bbox(text_id)

    def _draw_top_chromosome(self, chromosome: UIChromosome, x=0, y=0):
        if not self.solution_chromosome:
            value = chromosome.evaluated
            if value:
                value_str = f'{value:.2f}'
            else:
                value_str = '?'

            value_id = self.canvas.create_text((x, y), text=value_str,
                                               anchor=W, font='monospace 20', fill='cyan')
            value_bbox = self.canvas.bbox(value_id)
            x = value_bbox[2]

            equals_id = self.canvas.create_text((x, y), text='=',
                                                font='monospace 20', fill='white',
                                                anchor=W)
            equals_bbox = self.canvas.bbox(equals_id)
            x = equals_bbox[2]

        self.canvas.create_text((x, y), text=chromosome.decoded_str, anchor=W,
                                font='monospace 20', fill='cyan')

    def _draw_iteration(self, x=0, y=0):
        self.canvas.create_text(
            (x, y),
            text=f'Iteration #{self.iteration:,d}',
            fill='orange',
            font='monospace 30',
            anchor=NE,
        )

    def _restart_simulation(self, solution=None):
        if solution is None:
            solution = random.randint(10, 100000)

        self.sim = UISimulation(solution, population_size=self.population_size)
        self.iteration = 0
        self.drawn_solution = False
        self.solution_chromosome = None
        self.paused = False

    def on_new_simulation_button_pressed(self, event):
        self._restart_simulation()

    def on_pause_button_pressed(self, event):
        if self.solution_chromosome:
            self._restart_simulation()
        else:
            self.paused = not self.paused

    def on_step_button_pressed(self, event):
        if self.solution_chromosome:
            self._restart_simulation()
            self.paused = True
        else:
            self._iterate(step=True)

    def on_toggle_decoded_button_pressed(self, event):
        self.show_decoded = not self.show_decoded
        self._redraw()

    def on_restart_button_pressed(self, event):
        self._restart_simulation(solution=self.sim.solution)

    def on_change_pop_size_button_pressed(self, event):
        population_size = askinteger(
            title='Restart with different population size',
            prompt='Enter a new population size',
            initialvalue=self.population_size,
        )
        if population_size is not None:
            self.population_size = population_size
            self._restart_simulation(solution=self.sim.solution)
            self._resize()

    def on_enter_new_target_button_pressed(self, event):
        solution = askinteger(
            title='Restart with selected target',
            prompt='Enter a new target number',
            initialvalue=self.sim.solution,
        )
        if solution is not None:
            self._restart_simulation(solution=solution)


def start_interface():
    solution = random.randint(1, 1000)
    sim = UISimulation(solution)
    ui: Optional[GeneticExprUI] = None

    def start_ui():
        nonlocal ui
        ui = GeneticExprUI(sim)
        ui.run()

    def handle_sigint(sig, frame):
        if ui:
            ui.stop()

    signal.signal(signal.SIGINT, handle_sigint)
    ui_thread = Thread(target=start_ui)
    ui_thread.start()


if __name__ == '__main__':
    start_interface()
