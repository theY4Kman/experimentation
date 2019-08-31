import heapq
import struct
from collections import Counter
from dataclasses import dataclass
from functools import total_ordering
from typing import Optional, Tuple

from bitarray import bitarray


@total_ordering
@dataclass
class HuffNode:
    frequency: float = None
    left: 'HuffNode' = None
    right: 'HuffNode' = None
    sym: str = None

    bits: bitarray = None

    @classmethod
    def parent_for(cls, left: 'HuffNode', right: 'HuffNode'):
        return cls(
            left=left,
            right=right,
            frequency=left.frequency + right.frequency,
        )

    @property
    def is_leaf(self) -> bool:
        return self.left is None and self.right is None

    def __float__(self) -> float:
        return self.frequency

    def __lt__(self, other: 'HuffNode') -> bool:
        return self.frequency < other.frequency

    def __eq__(self, other: 'HuffNode') -> bool:
        return self.frequency == other.frequency

    def __getitem__(self, index: int) -> Optional['HuffNode']:
        if index == 0:
            return self.left
        elif index == 1:
            return self.right
        else:
            raise ValueError(
                f'Slicing on a binary tree must use either a 1 or 0. Found: {index!r}')


def encode(source: str) -> bytes:
    sym_total = len(source)
    sym_counts = Counter(source)

    symbols = {
        sym: HuffNode(frequency=sym_count / sym_total, sym=sym)
        for sym, sym_count in sym_counts.items()
    }
    nodes = list(symbols.values())
    heapq.heapify(nodes)

    while len(nodes) > 1:
        left = heapq.heappop(nodes)
        right = heapq.heappop(nodes)

        parent = HuffNode.parent_for(left, right)
        heapq.heappush(nodes, parent)

    root = nodes[0]

    height = 0
    to_visit = [(root, bitarray(), height)]
    while to_visit:
        node, bits, level = to_visit.pop()

        if node.is_leaf:
            node.bits = bits

            if level > height:
                height = level

        else:
            to_visit.append((node.left, bits + (False,), level + 1))
            to_visit.append((node.right, bits + (True,), level + 1))

    encoded = bitarray()
    for c in source:
        node = symbols[c]
        encoded += node.bits

    lookup_table = bitarray()
    for node in sorted(symbols.values(), key=lambda node: len(node.bits)):
        lookup_table += len(node.bits) * bitarray('1')
        lookup_table.append(False)
        lookup_table += node.bits
        lookup_table.frombytes(node.sym.encode('latin-1'))  # TODO: utf-8

    header = struct.pack('BB', height, len(symbols))
    body = lookup_table + encoded

    body_padlen = (8 - len(body) % 8) % 8
    padding_mark = body_padlen * bitarray('1')
    padding_mark.fill()

    compressed = header + (padding_mark + body)

    return compressed


def decode(encoded: bytes) -> str:
    height, num_syms = struct.unpack('BB', encoded[:2])

    bits = bitarray()
    bits.frombytes(encoded[2:])

    padding_mark, bits = bits[:8], bits[8:]
    while padding_mark.pop(0):
        bits = bits[:-1]

    nodes = []
    for i in range(num_syms):
        node = HuffNode()

        bitlen = 0
        while bits.pop(0):
            bitlen += 1

        node.bits = bits[:bitlen]
        del bits[:bitlen]

        node.sym = bits[:8].tobytes().decode('latin-1')  # TODO: utf-8
        del bits[:8]

        nodes.append(node)

    decoded = []
    buf = bitarray()
    while bits:
        buf += bits[:1]
        bits = bits[1:]

        for node in nodes:
            if node.bits == buf:
                decoded.append(node.sym)
                buf = bitarray('')
                break

    return ''.join(decoded)
