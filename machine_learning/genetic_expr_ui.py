from Tkinter import *
import random

from genetic_expr2 import Simulation, Chromosome


class UIChromosome(Chromosome):
    COLOR_VALID = 'green'
    COLOR_INVALID = 'yellow'
    COLOR_UNKNOWN = 'red'

    def __init__(self, gene_string, solution):
        self.gene_colors = []
        super(UIChromosome, self).__init__(gene_string, solution)

    def _decode(self):
        expr = []
        need_digit = True
        for gene_value in self.genes:
            value = self._decode_gene(gene_value)
            if value is None:
                self.gene_colors.append(self.COLOR_UNKNOWN)
                continue

            if need_digit:
                if value in self.GENE_VALUE_DIGITS:
                    expr.append(value)
                    self.gene_colors.append(self.COLOR_VALID)
                    need_digit = False
                    continue
            elif value in self.GENE_VALUE_OPERATORS:
                expr.append(value)
                self.gene_colors.append(self.COLOR_VALID)
                need_digit = True
                continue

            self.gene_colors.append(self.COLOR_INVALID)

        # Catches the case of an operator at the end of the chromosome (useless)
        if expr and need_digit:
            expr.pop()

        return ''.join(expr)



class UISimulation(Simulation):
    chromosome_class = UIChromosome


class GeneticExprUI(object):
    FRAMES_PER_SECOND = 60
    MILLISECONDS_PER_FRAME = 1000/FRAMES_PER_SECOND

    def __init__(self, simulation):
        """
        @type   simulation: UISimulation
        """
        self.sim = simulation

        self.tk = Tk()
        self.tk.minsize(1145, 250)
        self.tk.geometry('1145x550')
        self.tk.configure(background='black')

        self.new_button = Button(self.tk, text='New Simulation')
        self.new_button.bind('<Button-1>', self._restart_simulation_evt)
        self.new_button.pack()

        self.canvas = Canvas(self.tk, height=700)
        self.canvas.configure(background='black')
        self.canvas.pack(fill=BOTH)

        self.iteration = 0
        self.solution_chromosome = None
        self.drawn_solution = False

    def _run(self):
        if not self.solution_chromosome or not self.drawn_solution:
            self.canvas.delete(ALL)
            self.chromosome_ids = self._draw_lines(20, y=50)
            self._draw_top_status(x=10, y=5)
            if self.solution_chromosome:
                self.drawn_solution = True

        if not self.solution_chromosome:
            self.iteration += 1
            self.solution_chromosome = self.sim.step()

        self.tk.after(self.MILLISECONDS_PER_FRAME, self._run)

    def run(self):
        self.tk.after(self.MILLISECONDS_PER_FRAME, self._run)
        self.iteration = 0
        self.tk.mainloop()

    def _draw_lines(self, x=0, y=0):
        ids = []
        for i,chromosome in enumerate(self.sim.population):
            ids += self._draw_line(x, y + i*15, chromosome)
        return ids

    def _draw_line(self, x, y, chromosome):
        ids = []
        orig_bbox = bbox = (x, y, x-5, y)
        for gene_string,gene_color in zip(chromosome.genes,
                                          chromosome.gene_colors):
            text_id = self.canvas.create_text((bbox[2] + 5, bbox[1]),
                                              text=gene_string, anchor=NW,
                                              fill=gene_color)
            bbox = self.canvas.bbox(text_id)
            ids.append(text_id)
        if chromosome.is_solution:
            value_ids = []
            gene_width = bbox[2] - bbox[0]
            val_bbox = (x + gene_width / 2, bbox[1] + 5) * 2
            for gene_value in chromosome.genes:
                value_id = self.canvas.create_text(
                    (val_bbox[2] + gene_width, val_bbox[1]), text=gene_value,
                    fill='white')
                val_bbox = self.canvas.bbox(value_id)
                value_ids.append(value_id)
            self.canvas.create_rectangle(
                ((orig_bbox[0] - 18,) + (y,) + val_bbox[2:]),
                fill='', outline='white')
            ids += value_ids
        return ids

    def _draw_top_status(self, x=0, y=0):
        sol_bbox = self._draw_solution(x, y)
        equals_bbox = self._draw_equals_sign(sol_bbox[2], y)

        right_x = self.tk.winfo_width()
        self._draw_iteration(right_x, y)

        middle_y = float(equals_bbox[1] + equals_bbox[3]) / 2

        top_chromosome = sorted(self.sim.population, key=lambda c: c.fitness)[0]
        self._draw_top_chromosome(top_chromosome, equals_bbox[2], middle_y)

    def _draw_solution(self, x=0, y=0):
        text_id = self.canvas.create_text((x, y), text=str(self.sim.solution),
                                          anchor=NW, font='monospace 30',
                                          fill='white')
        return self.canvas.bbox(text_id)

    def _draw_equals_sign(self, x=0, y=0):
        symbol = '=' if self.solution_chromosome else u'\u2260'
        text_id = self.canvas.create_text((x, y), text=symbol,
                                          font='monospace 30', fill='white',
                                          anchor=NW)
        return self.canvas.bbox(text_id)

    def _draw_top_chromosome(self, chromosome, x=0, y=0):
        self.canvas.create_text((x, y), text=chromosome.decoded, anchor=W,
                                font='monospace 20', fill='cyan')

    def _draw_iteration(self, x=0, y=0):
        self.canvas.create_text((x, y), text='Iteration #%d' % self.iteration,
                                fill='orange', font='monospace 30', anchor=NE)

    def _restart_simulation(self):
        solution = random.randint(10, 1000)
        self.sim = UISimulation(solution)
        self.iteration = 0
        self.drawn_solution = False
        self.solution_chromosome = None

    def _restart_simulation_evt(self, event):
        self._restart_simulation()


def start_interface():
    solution = random.randint(1, 1000)
    sim = UISimulation(solution)
    ui = GeneticExprUI(sim)
    ui.run()


if __name__ == '__main__':
    start_interface()
