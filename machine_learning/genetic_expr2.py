#!/usr/bin/python
# Uses a genetic algorithm to find an arithmetic expression which evaluates to
# a certain solution.
import argparse
import os
import random


# Levels of verbosity
VERB_NONE = 0   # No messages printed
VERB_RUN = 1    # Simulation summary
VERB_INFO = 2   # Information on each iteration is printed
VERB_DEBUG = 3  # PRINT ALL THE THINGS

DEFAULT_CHROMOSOME_SIZE = 30    # Number of genes in a chromosome
DEFAULT_CROSSOVER_RATE = 0.8    # Chance two chromosomes will swap their bits
DEFAULT_MUTATION_RATE = 0.01    # Chance each bit will be mutated
DEFAULT_POPULATION_SIZE = 30    # Number of chromosomes in a population


def split_n_chars(s, n):
    """Splits a string at every n chars"""
    return [s[i:i+n] for i in range(0, len(s), n)]


def ndigit_bin(n, digits=0):
    """Returns the binary representation of n, zero-filled to `digits` places"""
    binval = bin(n)[2:]  # strip '0b'
    return binval.zfill(digits)


def round_up_div(a, b):
    return (a + (-a % b)) // b


def string_to_bits(s):
    """Return a bitstring representing each char in s"""
    return ''.join(ndigit_bin(ord(c), 8) for c in s)


class Chromosome(object):
    """Represents a chromosome, i.e. a string of genes."""

    GENE_SIZE = 4  # Number of binary digits in a gene
    GENE_VALUE_DIGITS = '0123456789'
    GENE_VALUE_OPERATORS = '+-*/'
    GENE_VALUES = GENE_VALUE_DIGITS + GENE_VALUE_OPERATORS
    GENE_VALUE_BITS = dict((ndigit_bin(i, 4), value)
                           for i, value in enumerate(GENE_VALUES))

    def __init__(self, gene_string, solution):
        self.solution = solution
        self.gene_string = gene_string
        self.genes = split_n_chars(self.gene_string, 4)
        self.decoded = self._decode()
        self.evaluated = self._evaluate(self.decoded)
        self.fitness = self._calculate_fitness()
        self.is_solution = self.fitness is None

    def __repr__(self):
        return '%s(%r, %r)' % (self.__class__.__name__, self.gene_string,
                               self.solution)

    def __str__(self):
        return self.gene_string

    def __len__(self):
        return len(self.gene_string)

    def __getitem__(self, item):
        return self.gene_string.__getitem__(item)

    def __eq__(self, other):
        return (self.solution == other.solution
                and self.gene_string == other.gene_string)

    def __hash__(self):
        return hash((self.solution, self.gene_string))

    def _decode(self):
        expr = []
        need_digit = True
        for gene_value in self.genes:
            value = self._decode_gene(gene_value)
            if value is None:
                continue

            if need_digit:
                if value in self.GENE_VALUE_DIGITS:
                    expr.append(value)
                    need_digit = False
            elif value in self.GENE_VALUE_OPERATORS:
                expr.append(value)
                need_digit = True

        # Catches the case of an operator at the end of the chromosome (useless)
        if expr and need_digit:
            expr.pop()

        return ''.join(expr)

    def _decode_gene(self, gene_value):
        return self.GENE_VALUE_BITS.get(gene_value)

    def _calculate_fitness(self):
        if self.evaluated is None:
            return 0.0
        try:
            return 1.0 / (self.solution - self.evaluated)
        except ZeroDivisionError:
            return None

    def _evaluate(self, expr):
        """Returns None if the evaluation fails"""
        if not expr:
            return None

        try:
            return eval(expr)
        except ZeroDivisionError:
            return None

    def mutate(self, mutation_rate):
        mutated_bits = []
        for bit in self.gene_string:
            if random.random() < mutation_rate:
                bit = '0' if bit == '1' else '1'
            mutated_bits.append(bit)

        mutated_gene_string = ''.join(mutated_bits)
        return self.__class__(mutated_gene_string, self.solution)

    @classmethod
    def random(cls, solution, num_genes):
        bits_in_a_byte = 8
        total_bits_needed = num_genes * cls.GENE_SIZE
        total_bytes_needed = round_up_div(total_bits_needed, bits_in_a_byte)

        random_bits = os.urandom(total_bytes_needed)
        bitstring = string_to_bits(random_bits)
        chromosome_bits = bitstring[:total_bits_needed]

        return cls(chromosome_bits, solution)

    @classmethod
    def crossover(cls, a, b):
        fulcrum = random.randint(0, len(a)-1)
        new_x = cls(a[:fulcrum] + b[fulcrum:], a.solution)
        new_y = cls(b[:fulcrum] + a[fulcrum:], b.solution)
        return new_x, new_y


class Simulation(object):
    chromosome_class = Chromosome

    def __init__(self, solution, population_size=30, chromosome_size=30,
                 crossover_rate=0.8, mutation_rate=0.01, max_iterations=1000,
                 verbosity=VERB_NONE):
        self.verbosity = verbosity

        self.max_iterations = max_iterations
        self.solution = solution

        self.chromosome_size = chromosome_size
        self.population_size = population_size
        self.crossover_rate = crossover_rate
        self.mutation_rate = mutation_rate

        self.population = self._generate_random_population()

    def _generate_random_population(self):
        return [self.chromosome_class.random(self.solution,
                                             self.chromosome_size)
                for _ in xrange(self.population_size)]

    def step(self):
        solution_chromosome = self.check_for_solution()
        if solution_chromosome:
            return solution_chromosome

        self._iterate_population()

    def _iterate_population(self):
        self.population = self._generate_population_iteration()

    def _generate_population_iteration(self):
        new_population = []
        new_population_size = 0
        while new_population_size < self.population_size:
            new_population += self._new_children()
            new_population_size += 2  # _new_children always returns 2
        return new_population[:self.population_size]

    def _roulette_wheel(self):
        """Returns a random chromosome (random with respect to fitness)"""
        total_fitness = sum(abs(chromosome.fitness)
                            for chromosome in self.population)
        pick = random.uniform(0, total_fitness)
        current = 0.0
        for chromosome in self.population:
            current += abs(chromosome.fitness)
            if current > pick:
                return chromosome
        else:
            # If no chromosome is selected (which can happen if all fitness
            # values are 0.0), revert to a random choice.
            return random.choice(self.population)

    def _new_children(self):
        a = self._roulette_wheel()
        b = self._roulette_wheel()

        # See if we should crossover
        if random.random() <= self.crossover_rate:
            a, b = self.chromosome_class.crossover(a, b)

        a = a.mutate(self.mutation_rate)
        b = b.mutate(self.mutation_rate)

        return a, b

    def check_for_solution(self):
        for chromosome in self.population:
            if chromosome.is_solution:
                return chromosome

    def _run(self, max_iterations):
        for iteration in xrange(max_iterations):
            self._print('{:#^30}'.format(' ITERATION %d ' % iteration), VERB_INFO)
            self._print('{:*^30}'.format(' SOLUTION: %d ' % self.solution), VERB_INFO)
            self._print_population(VERB_INFO)
            self._print('', VERB_INFO)

            solution_chromosome = self.step()
            if solution_chromosome:
                return iteration, solution_chromosome
        else:
            return max_iterations, None

    def _str_population(self):
        population_str = []
        for chromosome in sorted(self.population, reverse=True,
                                 key=lambda c: abs(c.fitness) if c.fitness else None):
            population_str.append(self._get_chromosome_summary(chromosome))
        return '\n'.join(population_str)

    def _print_population(self, level=VERB_NONE):
        if self.verbosity >= level:
            print self._str_population()

    def _get_chromosome_summary(self, chromosome):
        if chromosome.fitness is None:
            fitness = ' SOLVE'
        else:
            fitness = '% 6.3f' % chromosome.fitness

        return ('{fitness}'
                '{0.evaluated:>6} = {0.decoded:<{self.chromosome_size}}'
                '{0:s}').format(chromosome, self=self, fitness=fitness)

    def run(self, max_iterations=None):
        if max_iterations is None:
            max_iterations = self.max_iterations

        self._print('Solution: %d\nBeginning simulation...' % self.solution,
                    VERB_RUN)
        iterations, solution_chromosome = self._run(max_iterations)

        if solution_chromosome:
            gene_symbols = '   '.join(
                self.chromosome_class.GENE_VALUE_BITS.get(gene_value, ' ')
                for gene_value in solution_chromosome.genes)

            summary = 'Solution found in %d iteration(s): %d = %s\n%s\n%s' % (
                iterations, self.solution, solution_chromosome.decoded,
                solution_chromosome,
                gene_symbols)
        else:
            summary = 'No solution found in %d iteration(s)' % iterations
        self._print(summary, VERB_RUN)

    def _print(self, msg, level):
        if self.verbosity >= level:
            print msg


def main(argv):
    parser = argparse.ArgumentParser(description='Use a genetic algorithm to '
                                                 'find an expression matching '
                                                 'a solution')
    parser.add_argument('solution', default=None, type=int, nargs='?',
                        help='Number to match with a generated expression')
    parser.add_argument('-i', '--max-iterations', default=1000, type=int,
                        help='Maximum number of iterations')
    parser.add_argument('-g', '--chromosome-size', default=DEFAULT_CHROMOSOME_SIZE,
                        type=int, help='Number of genes in a chromosome')
    parser.add_argument('-p', '--population-size', default=DEFAULT_POPULATION_SIZE,
                        type=int, help='Number of chromosomes in a population')
    parser.add_argument('-c', '--crossover-rate', default=DEFAULT_CROSSOVER_RATE,
                        type=float, help='Chance of swapping chromosome\'s '
                                         'bits')
    parser.add_argument('-m', '--mutation-rate', default=DEFAULT_MUTATION_RATE,
                        type=float, help='Chance an individual bit will be'
                                         'mutated')
    parser.add_argument('-v', action='count', default=VERB_RUN,
                        help='Level of verbosity', dest='verbosity')

    args = parser.parse_args(argv[1:])
    if args.solution is None:
        args.solution = random.randint(10, 1000)

    sim_args = vars(args)
    simulation = Simulation(**sim_args)
    simulation.run()


if __name__ == '__main__':
    import sys
    main(sys.argv)

