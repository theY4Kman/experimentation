#!/usr/bin/python2.7
# Implements a genetic algorithm to find a simple mathematical expression which
# is equivalent to the solution number. The expression must be of the form:
#   digit | operator | digit | operator | digit ...
from __future__ import division
import argparse
from random import SystemRandom
random = SystemRandom()


class SolutionFound(Exception):
    pass


GENE_BIT_LENGTH = 4 # Assumes 16 max gene values to encode (into 4 bits each)

def ndigit_bin(n, digits=GENE_BIT_LENGTH):
    binval = bin(n)[2:] # strip '0b'
    return binval.zfill(digits)

GENE_VALUE_DIGITS = '0123456789'
GENE_VALUE_OPERATORS = '+-*/'
GENE_VALUES = GENE_VALUE_DIGITS + GENE_VALUE_OPERATORS
GENE_VALUE_BITS = dict((ndigit_bin(i), value)
                       for i,value in enumerate(GENE_VALUES))

# Number of genes in a chromosome
CHROMOSOME_LENGTH = 30
# Chance two selected chromosomes will swap their bits
CROSSOVER_RATE = 0.8
# Chance each bit will be flipped when mutating a chromosome
MUTATION_RATE = 0.01
# Number of chromosomes in a population
POPULATION_SIZE = 30


def split_n_chars(s, n=GENE_BIT_LENGTH):
    return [s[i:i+n] for i in range(0, len(s), n)]


def roulette_wheel(popfit):
    """Takes an iterable of (chromosome, fitness) and returns a random choice,
    taking fitness values into account."""
    assert popfit
    maxval = sum(abs(fitness) for _,fitness in popfit)
    pick = random.uniform(0, maxval)
    current = 0
    for chromosome,fitness in popfit:
        current += abs(fitness)
        if current > pick:
            return chromosome
    return chromosome


def decode_chromosome(chromosome):
    """Decodes a chromosome's bitstring, producing the expression represented
    by it. It first looks for a digit, ignoring genes until one has been found.
    Then it searches for an operator, ignoring digits... till the end of the
    bitstring."""
    gene_values = filter(None, map(lambda gv: GENE_VALUE_BITS.get(gv),
                                   split_n_chars(chromosome)))
    expr = []
    need_digit = True
    for value in gene_values:
        if need_digit:
            if value in GENE_VALUE_DIGITS:
                expr.append(value)
                need_digit = False
        elif value in GENE_VALUE_OPERATORS:
            expr.append(value)
            need_digit = True

    # Catches the case of an operator at the end of the chromosome (useless)
    if expr and need_digit:
        expr.pop()

    return ''.join(expr)


def evaluate_expression(expr):
    return eval(expr)


def evaluate_chromosome(chromosome):
    """Decodes a chromosome into a mathematical expression and evaluates it."""
    # eval() is safe because it is limited to the chars defined in GENE_VALUES
    return evaluate_expression(decode_chromosome(chromosome))


def chromosome_fitness(solution, chromosome):
    """Determines the fitness value of a chromosome. This is the inverse of the
    difference between the solution and the value of the chromosome's
    expression when evaluated."""
    expr = decode_chromosome(chromosome)
    if not expr:
        return 0

    try:
        expr_eval = evaluate_expression(expr)
    except ZeroDivisionError:
        return 0

    try:
        return 1.0 / (solution - expr_eval)
    except ZeroDivisionError:
        raise SolutionFound(chromosome, expr)


def crossover_chromosomes(a, b):
    """ Swaps parts of a & b at a random point """
    # XXX: normalize lengths? (r: don't need the protection in this instance)
    assert len(a) == len(b)
    length = len(a)
    fulcrum = random.randint(0, length-1)
    return (a[:fulcrum] + b[fulcrum:], b[:fulcrum] + a[fulcrum:])


def mutate_chromosome(chromosome, rate=MUTATION_RATE):
    """ Flips bits in the chromosome dependent on the mutation `rate` """
    mutated = []
    for c in chromosome:
        if random.random() <= rate:
            c = '0' if c == '1' else '1'
        mutated.append(c)
    return ''.join(mutated)


def random_chromosome(genes=CHROMOSOME_LENGTH):
    num_chars = genes * GENE_BIT_LENGTH
    return ''.join(random.choice('01') for _ in xrange(num_chars))


def build_random_chromosomes(genes=CHROMOSOME_LENGTH, size=POPULATION_SIZE):
    """ Builds a population filled with random chromosomes."""
    return [random_chromosome(genes) for _ in xrange(size)]


def population_fitness(solution, population):
    """ Evaluates the fitness of each chromosome in the population, returning a
    list of (chromosome, fitness) """
    popfit = []
    for chromosome in population:
        try:
            popfit.append((chromosome,
                           chromosome_fitness(solution, chromosome)))
        except SyntaxError:
            pass
    return popfit


def population_new_children(popfit, crossover_rate=CROSSOVER_RATE,
                            mutation_rate=MUTATION_RATE):
    """ Creates a new child chromosome from two selected in `popfit`.
    `popfit` should be a list of (chromosome, fitness) """
    a = roulette_wheel(popfit)
    b = roulette_wheel(popfit)

    # See if we should crossover
    if random.random() <= crossover_rate:
        a,b = crossover_chromosomes(a, b)

    a = mutate_chromosome(a, mutation_rate)
    b = mutate_chromosome(b, mutation_rate)

    return a,b


def population_iterate(popfit, size=POPULATION_SIZE,
                       crossover_rate=CROSSOVER_RATE,
                       mutation_rate=MUTATION_RATE):
    """ Generates a new population based on a previous population. """
    new_chromosomes = []
    num_new = 0
    while num_new < size:
        new_chromosomes += population_new_children(popfit, crossover_rate,
                                                   mutation_rate)
        num_new += 2
    # XXX: assumes an even population size. Otherwise, this should return
    # new_chromosomes[:size], to respect the size requested
    return new_chromosomes


def run_genetic_expr_finder(solution, max_iterations=1000,
                            chromosome_size=CHROMOSOME_LENGTH,
                            population_size=POPULATION_SIZE,
                            crossover_rate=CROSSOVER_RATE,
                            mutation_rate=MUTATION_RATE, debug=False):
    print 'SOLUTION:', solution
    print

    chromosomes = build_random_chromosomes(chromosome_size, population_size)
    for iteration in xrange(max_iterations):
        if debug:
            print 'ITERATION %4d   SOLUTION %d' % (iteration, solution)

        try:
            popfit = population_fitness(solution, chromosomes)
        except SolutionFound as solution:
            chromosome,expr = solution.args
            print 'ITERATION:', iteration
            print 'SOLUTION FOUND:', expr
            print '  %s' % chromosome
            break

        chromosomes = population_iterate(popfit, population_size,
                                         crossover_rate, mutation_rate)

        # Current status printing
        if debug:
            for chromosome,fitness in sorted(popfit, key=lambda o: o[1],
                                             reverse=True):
                expr = decode_chromosome(chromosome)
                try:
                    expr_eval = '%3d' % evaluate_expression(expr)
                except (ZeroDivisionError, SyntaxError):
                    expr_eval = 'NaN'
                print '  %s %.4f  %s = %s' % (chromosome, fitness, expr_eval,
                                              expr)

            print
            print
    else:
        print 'NO SOLUTION FOUND IN %d ITERATIONS' % max_iterations


def main(argv=()):
    parser = argparse.ArgumentParser(description='Use a genetic algorithm to '
                                                 'find an expression matching '
                                                 'a solution')
    parser.add_argument('solution', default=None, type=int, nargs='?',
                        help='Number to match with a generated expression')
    parser.add_argument('-i', '--max-iterations', default=1000, type=int,
                        help='Maximum number of iterations')
    parser.add_argument('-g', '--chromosome-size', default=CHROMOSOME_LENGTH,
                        type=int, help='Number of genes in a chromosome')
    parser.add_argument('-p', '--population-size', default=POPULATION_SIZE,
                        type=int, help='Number of chromosomes in a population')
    parser.add_argument('-c', '--crossover-rate', default=CROSSOVER_RATE,
                        type=float, help='Chance of swapping chromosome\'s '
                                         'bits')
    parser.add_argument('-m', '--mutation-rate', default=MUTATION_RATE,
                        type=float, help='Chance an individual bit will be'
                                         'mutated')
    parser.add_argument('-d', '--debug', action='store_true', default=False,
                        help='Print more information')

    args = parser.parse_args(argv[1:])
    if args.solution is None:
        args.solution = random.randint(10, 1000)

    run_genetic_expr_finder(args.solution, args.max_iterations,
                            args.chromosome_size, args.population_size,
                            args.crossover_rate, args.mutation_rate, args.debug)

if __name__ == '__main__':
    import sys
    main(sys.argv)
