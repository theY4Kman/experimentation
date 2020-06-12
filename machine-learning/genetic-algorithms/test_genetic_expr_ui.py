import pytest

from genetic_expr_ui import UIChromosome

EXPR_GENES = {
    expr: bits
    for bits, expr in UIChromosome.GENE_VALUE_BITS.items()
}
EXPR_GENES['?'] = '1111'

COLORS = {
    '+': UIChromosome.COLOR_VALID,
    '-': UIChromosome.COLOR_INVALID,
    '?': UIChromosome.COLOR_UNKNOWN,
}


def encode_gene_expression(gene_expression: str) -> str:
    """Translate a gene expression to its 0s and 1s gene string representation

    >>> encode_gene_expression('1+2')
    '000110100010'
    >>> encode_gene_expression('1+2')
    '1100110000100100111110100011'
    """
    return ''.join(EXPR_GENES[c] for c in gene_expression)


@pytest.mark.parametrize('gene_expression,colors,decoded', [
    ('1',
     '+', '1'),

    ('1+',
     '+-', '1'),

    ('1+2',
     '+++', '1+2'),

    ('+1+2',
     '++++', '+1+2'),

    ('+1?+2',
     '++?++', '+1+2'),

    ('+?1?+2',
     '+?+?++', '+1+2'),

    ('02',
     '-+', '2'),

    ('20',
     '++', '20'),

    ('0',
     '+', '0'),
])
def test_decode(gene_expression: str, colors: str, decoded: str):
    gene_string = encode_gene_expression(gene_expression)
    chromosome = UIChromosome(gene_string, 123)

    expected_colors = [COLORS[c] for c in colors]

    expected = (decoded, expected_colors)
    actual = (chromosome.decoded, chromosome.gene_colors)
    assert expected == actual

