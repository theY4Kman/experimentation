#!/usr/bin/python
# Uses a genetic algorithm to find an arithmetic expression which evaluates to
# a certain solution.

import argparse
import math
import os
import random
from typing import Optional, Union


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
    if isinstance(s, str):
        s = [ord(c) for c in s]
    return ''.join(f'{c:08b}' for c in s)


class Chromosome:
    """Represents a chromosome, i.e. a string of genes."""

    GENE_SIZE = 4  # Number of binary digits in a gene
    GENE_VALUE_DIGITS = '0123456789'
    GENE_VALUE_OPERATORS = '+-*/'
    GENE_VALUES = GENE_VALUE_DIGITS + GENE_VALUE_OPERATORS
    GENE_VALUE_BITS = {
        f'{i:04b}': value
        for i, value in enumerate(GENE_VALUES)
    }

    def __init__(self, gene_string, solution):
        self.solution = solution
        self.gene_string = gene_string
        self.genes = split_n_chars(self.gene_string, 4)
        self.decoded = self.decode()
        self.evaluated = self.evaluate(self.decoded)
        self.is_solution = self.evaluated == solution
        self.fitness = self.calculate_fitness()

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

    def decode(self):
        expr = []
        for gene_value in self.genes:
            value = self.decode_gene(gene_value)
            if value is None:
                continue

            if value in self.GENE_VALUE_DIGITS or value in self.GENE_VALUE_OPERATORS:
                expr.append(value)

        # Catches the case of an operator at the end of the chromosome (useless)
        while expr and expr[-1] in self.GENE_VALUE_OPERATORS:
            expr.pop()

        return ''.join(expr)

    def decode_gene(self, gene_value):
        return self.GENE_VALUE_BITS.get(gene_value)

    def calculate_fitness(self) -> float:
        if self.evaluated is None:
            return 0.0

        is_integer = self.evaluated.is_integer() if isinstance(self.evaluated, float) else True
        int_bias = 1 if is_integer else 0.5

        if is_integer and self.solution == self.evaluated:
            return 1.0

        # Without a maximum possible non-solution score, 1 / (goal - eval) would be 1 / 1
        # if `eval` was only 1 away from `goal`.
        max_score = 0.96

        try:
            return max_score * 1.0 / int(abs(self.solution - self.evaluated)) * int_bias
        except ZeroDivisionError:
            return 0.0

    def evaluate(self, expr) -> Optional[Union[float, int]]:
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


class Simulation:
    chromosome_class = Chromosome

    def __init__(self, solution, population_size=30, chromosome_size=30,
                 crossover_rate=0.8, base_mutation_rate=0.01, max_iterations=1000,
                 verbosity=VERB_NONE):
        self.verbosity = verbosity

        self.iteration = 1
        self.max_iterations = max_iterations
        self.solution = solution

        self.chromosome_size = chromosome_size
        self.population_size = population_size
        self.crossover_rate = crossover_rate
        self.base_mutation_rate = base_mutation_rate

        self.population = self._generate_random_population()

    def _generate_random_population(self):
        return [self.chromosome_class.random(self.solution,
                                             self.chromosome_size)
                for _ in range(self.population_size)]

    def step(self):
        self._iterate_population()

        solution_chromosome = self.check_for_solution()
        if solution_chromosome:
            return solution_chromosome

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
        total_fitness = self._get_total_fitness()
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

    def _select_chromosomes(self, n=1):
        """Return a number of chromosomes, with respect to fitness
        """
        chromosomes = self.population

        for i in range(n):
            chromosomes = set(sorted(chromosomes, key=lambda c: c.fitness, reverse=i % 2 == 1))

            total_fitness = sum(abs(chromosome.fitness) for chromosome in chromosomes)
            pick = random.uniform(0, total_fitness)
            current = 0.0
            for chromosome in chromosomes:
                current += abs(chromosome.fitness)
                if current > pick:
                    chromosomes.remove(chromosome)
                    yield chromosome
                    break
            else:
                # If no chromosome is selected (which can happen if all fitness
                # values are 0.0), revert to a random choice.
                yield chromosomes.pop()

            if not chromosomes:
                break

    def _get_total_fitness(self):
        return sum(abs(chromosome.fitness) for chromosome in self.population)

    def _new_children(self):
        a, b = self._select_chromosomes(2)

        generation_multiplier = 2 - math.log(self.iteration % 100 + 1, 100)
        generation_multiplier_alt = 2 - math.log(101 - self.iteration % 100, 100)

        # See if we should crossover
        if random.random() <= self.crossover_rate:
            a, b = self.chromosome_class.crossover(a, b)

        # time_multiplier = (1 + math.log(self.iteration))
        mutation_rate = self.base_mutation_rate * generation_multiplier - random.random() * self.base_mutation_rate * generation_multiplier

        a_mutation_rate = mutation_rate * (1 - abs(a.fitness) + random.random() * generation_multiplier)
        b_mutation_rate = mutation_rate * (1 - abs(b.fitness) + random.random() * generation_multiplier)

        a = a.mutate(a_mutation_rate)
        b = b.mutate(b_mutation_rate)

        return a, b

    def check_for_solution(self):
        for chromosome in self.population:
            if chromosome.is_solution:
                return chromosome

    def _run(self, max_iterations):
        for iteration in range(max_iterations):
            self.iteration = iteration

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
            print(self._str_population())

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
            print(msg)


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

