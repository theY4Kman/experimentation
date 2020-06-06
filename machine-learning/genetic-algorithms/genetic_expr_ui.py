from __future__ import print_function

import random
from bisect import bisect

try:
    from Tkinter import *
    from Tkinter.tkSimpleDialog import askinteger
except ImportError:  # Python 3
    from tkinter import *
    from tkinter.simpledialog import askinteger

from genetic_expr2 import Simulation, Chromosome


class UIChromosome(Chromosome):
    COLOR_VALID = 'green'
    COLOR_INVALID = 'yellow'
    COLOR_UNKNOWN = 'red'

    def __init__(self, gene_string, solution):
        self.gene_colors = []
        super(UIChromosome, self).__init__(gene_string, solution)

        self.decoded_str = '?'
        if self.decoded:
            self.decoded_str = self.decoded.replace('*', '×').replace('/', '÷')

    def decode(self):
        expr = []
        num_length = 0
        for gene_value in self.genes:
            value = self.decode_gene(gene_value)
            if value is None:
                self.gene_colors.append(self.COLOR_UNKNOWN)
                continue

            is_digit = value in self.GENE_VALUE_DIGITS
            is_operator = value in self.GENE_VALUE_OPERATORS

            was_operator = expr and expr[-1] in self.GENE_VALUE_OPERATORS
            was_digit = expr and expr[-1] in self.GENE_VALUE_DIGITS

            if is_operator:
                num_length = 0

            if is_digit:
                num_length += 1

            if (
                (is_operator and expr and not was_operator)
                or (is_digit and not (value == '0' and not was_digit) and num_length <= 2)
            ):
                expr.append(value)
                self.gene_colors.append(self.COLOR_VALID)
                continue

            self.gene_colors.append(self.COLOR_INVALID)

        # Catches the case of an operator at the end of the chromosome (useless)
        while expr and expr[-1] in self.GENE_VALUE_OPERATORS:
            expr.pop()

        return ''.join(expr)

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


class GeneticExprUI(object):
    FRAMES_PER_SECOND = 60
    MILLISECONDS_PER_FRAME = int(1000 / FRAMES_PER_SECOND)

    def __init__(self, simulation: UISimulation, *, population_size: int = 30):
        self.sim = simulation
        self.population_size = population_size

        self.tk = Tk()
        self.tk.minsize(1175, 250)
        self.tk.configure(background='black')

        self.button_frame = Frame(self.tk, relief=FLAT, bg='black')

        self.pause_button = Button(self.button_frame, text='Pause')
        self.pause_button.bind('<Button-1>', self.on_pause_button_pressed)
        self.pause_button.pack(side=LEFT)
        self.tk.bind('<space>', self.on_pause_button_pressed)

        self.step_button = Button(self.button_frame, text='⏽︎⏵︎')
        self.step_button.bind('<Button-1>', self.on_step_button_pressed)
        self.step_button.pack(side=LEFT)
        self.tk.bind('<space>', self.on_step_button_pressed)

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
            self.canvas.delete(ALL)
            self.chromosome_ids = self._draw_lines(20, 50)
            self._draw_top_status(10, 5)
            if self.solution_chromosome:
                self.paused = True
                self.drawn_solution = True

        if (not self.paused or step) and not self.solution_chromosome:
            self.iteration += 1
            self.sim.iteration = self.iteration
            self.solution_chromosome = self.sim.step()

        self.tk.update()

    def stop(self):
        self.tk.quit()

    def _resize(self):
        pop_minheight = 15 * self.sim.population_size
        status_height = 50
        canvas_height = pop_minheight + status_height

        button_frame_height = self.button_frame.winfo_height()

        bottom_margin = 20
        height = canvas_height + button_frame_height + bottom_margin

        self.canvas.config(height=canvas_height)
        self.tk.geometry(f'1475x{height}')

    @property
    def paused(self):
        return self._paused

    @paused.setter
    def paused(self, is_paused: bool):
        self._paused = is_paused
        self.pause_button.config(text='‣' if is_paused else '⏸')
        self.step_button.config(state=NORMAL if is_paused else DISABLED)

    def _draw_lines(self, x=0, y=0):
        ids = []
        for i, chromosome in enumerate(self.sim.population):
            ids += self._draw_line(x, y + i*15, chromosome)
        return ids

    def _draw_line(self, x, y, chromosome: UIChromosome):
        ids = []
        orig_bbox = bbox = (x, y, x-5, y)
        for gene_string, gene_color in zip(chromosome.genes, chromosome.gene_colors):
            text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                              text=gene_string, anchor=NW,
                                              fill=gene_color)
            bbox = self.canvas.bbox(text_id)
            ids.append(text_id)

        if chromosome.is_solution:
            self.canvas.create_rectangle(
                (orig_bbox[0] - 5, y, bbox[2] + 4, bbox[3] - 2),
                fill='', outline='white')

        # Draw expression
        value = chromosome.evaluated
        if value:
            value_str = f'{value: .2f}'
        else:
            value_str = '?'
        value_str = f'{value_str:>10}'

        font = 'monospace 10'
        value_color = chromosome.get_value_color()
        static_color = 'white'

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text='=', anchor=NW,
                                          fill=static_color, font=font)
        bbox = self.canvas.bbox(text_id)

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text=value_str, anchor=NW,
                                          fill=value_color, font=font)
        bbox = self.canvas.bbox(text_id)

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text='=', anchor=NW,
                                          fill=static_color, font=font)
        bbox = self.canvas.bbox(text_id)

        text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                          text=chromosome.decoded_str, anchor=NW,
                                          fill=value_color, font=font)
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
            top_chromosome = sorted(self.sim.population, key=lambda c: c.fitness)[0]

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
            solution = random.randint(10, 1000)

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
    ui = GeneticExprUI(sim)
    try:
        ui.run()
    except KeyboardInterrupt:
        ui.stop()


if __name__ == '__main__':
    start_interface()
