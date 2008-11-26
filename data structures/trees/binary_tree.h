/**
 * =============================================================================
 * Binary Tree
 *   An implementation of a rooted binary tree.
 * Copyright (C) 2008 Zach "theY4Kman" Kanzler
 * =============================================================================
 *
 * This program is free software; you can redistribute it and/or modify it under
 * the terms of the GNU General Public License, version 3.0, as published by the
 * Free Software Foundation.
 * 
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 * FOR A PARTICULAR PURPOSE.  See the GNU General Public License for more
 * details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program.  If not, see <http://www.gnu.org/licenses/>.
 */

/**
 * Returns true if the bit at the index in `bits` in `value` is 1, or false
 * otherwise.
 */
inline bool is_bit_set(long value, int bits)
{
    return (value & (1<<bits));
}

/**
 * A rooted binary tree is a tree structure that can have at most only two edges
 * per node: left and right. In this implementation, the left node is 0 and the
 * right node is 1. Therefore, we can traverse the tree using a string of bits.
 * For example, look at the following string of bits:
 *
 *                                   11010011011
 *      1      1     0      1     0     0      1      1     0      1      1
 *    right, right, left, right, left, left, right, right, left, right, right
 */
template<class K>
class RootedBinaryTree
{
public:
    
};
