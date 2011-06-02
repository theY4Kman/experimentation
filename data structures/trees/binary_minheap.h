/**
 * =============================================================================
 *   A minheap structure using arrays. It can be turned into a maxheap by
 *       writing a custom bheap_comparator function. Includes a function that
 *       outputs a DOT file representation of the heap for easy visualization.
 *   Copyright (C) 2011 Zach "theY4Kman" Kanzler
 * =============================================================================
 * 
 * A combination of a Patricia tree and a de la Briandias tree, meaning it fol-
 * lows partial nodes, and those nodes are stored in linked lists.
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

#include <malloc.h>

#ifndef true
#  define true 1
#endif
#ifndef false
#  define false 0
#endif

#define BHEAP_PARENT(i) ((i)==0?0:((i)-1)/2)
#define BHEAP_LCHILD(i) (2*(i)+1)
#define BHEAP_RCHILD(i) (2*(i)+2)

typedef char bool;
typedef unsigned int uint;

/* bool bheap_comparator(uint key1, void *value1, uint key2, void *value2):
 *    Tests whether key1/value1 is of a higher priority than key2/value2.
 *    Should return true if it is.
 */
typedef bool (*bheap_comparator)(uint, void *, uint, void *);

/* void bheap_dot_key(char *buffer, int maxlength, uint key, void *value)
 *    Write a string to the buffer, no more than maxlength characters long,
 *    that is a representation of the key/value pair -- for use in a DOT graph.
 */
typedef void (*bheap_dot_key)(char *, int, uint, void *);

typedef struct bheap_node
{
  uint key;
  void *value;
} bheap_node_t;

typedef struct bheap
{
  bheap_node_t *items;
  uint n;
  uint next;
} bheap_t;

/**
 * Creates a new binary heap structure
 * @param n: Number of elements the heap can hold
 */
bheap_t *
bheap_create(uint n)
{
  if (n == 0)
    return NULL;
  
  bheap_t *b = malloc(sizeof(bheap_t));
  b->n = n;
  b->next = 0;
  b->items = calloc(n, sizeof(bheap_node_t));
  
  return b;
}

/**
 * Frees a binary heap structure
 * @param b: The binary heap structure to free
 */
void
bheap_free(bheap_t *b)
{
  if (b == NULL)
    return;
  
  if (b->items != NULL)
    free(b->items);
  
  free(b);
}

/**
 * Returns the number of elements placed in the binary heap
 * @param b: The binary heap structure
 * @return: The number of nodes inserted into the binary heap.
 */
uint
bheap_size(bheap_t *b)
{
  if (b == NULL)
    return 0;
  
  return b->next;
}

/**
 * Inserts `value` into the binary heap associated with `key`
 * @param b: The binary heap structure
 * @param key: The key to associate the value with.
 * @param value: Arbitrary data to enter.
 * @return: true on success, false if invalid bheap struct or bheap is full
 */
bool
bheap_insert(bheap_t *b, uint key, void *value, bheap_comparator comp)
{
  uint parent;
  uint node_idx;
  
  if (b == NULL || b->items == NULL)
    return false;
  
  if (b->next >= b->n)
    return false;
  
  node_idx = b->next;
  b->items[node_idx].key = key;
  b->items[node_idx].value = value;
  b->next++;
  
  /* Upheap:
   *   If the new node has a higher priority than its parent, swap them.
   *   Continue.
   */
  while ((parent = BHEAP_PARENT(node_idx)) != node_idx && comp(key, value,
      b->items[parent].key, b->items[parent].value))
  {
    b->items[node_idx].key = b->items[parent].key;
    b->items[node_idx].value = b->items[parent].value;
    
    b->items[parent].key = key;
    b->items[parent].value = value;
    
    node_idx = parent;
  }
  
  return true;
}

/**
 * Gives back the value of the root node and removes it from the heap.
 * @param b: The binary heap structure
 * @param value: Where to store the value, if successfull.
 * @return: true on success, false if invalid bheap struct or bheap is empty
 * @note: value is not modified upon case of failure
 */
bool
bheap_rootpop(bheap_t *b, void **value, bheap_comparator comp)
{
  uint cur_idx;
  
  if (b == NULL || b->items == NULL)
    return false;
  
  if (b->next == 0)
    return false;
  
  *value = b->items[0].value;
  
  /* Downheap:
   *   First, replace the root node with the last element of the last level.
   *   Then, if the root node has a lower priority than its children, replace
   *   it with the child that has the higher priority. Continue down each level
   */
  b->next--;
  
  b->items[0].key = b->items[b->next].key;
  b->items[0].value = b->items[b->next].value;
  
  cur_idx = 0;
  
  while ((BHEAP_LCHILD(cur_idx) < b->next &&
        comp(b->items[BHEAP_LCHILD(cur_idx)].key,
          b->items[BHEAP_LCHILD(cur_idx)].value,
          b->items[cur_idx].key, b->items[cur_idx].value)) ||
      
      (BHEAP_RCHILD(cur_idx) < b->next &&
        comp(b->items[BHEAP_RCHILD(cur_idx)].key,
          b->items[BHEAP_RCHILD(cur_idx)].value,
          b->items[cur_idx].key, b->items[cur_idx].value)))
  {
    uint tempkey;
    void *tempvalue;
    
    uint min_child = BHEAP_LCHILD(cur_idx);
    if (BHEAP_RCHILD(cur_idx) < b->next &&
        comp(b->items[BHEAP_RCHILD(cur_idx)].key,
          b->items[BHEAP_RCHILD(cur_idx)].value, b->items[min_child].key,
          b->items[min_child].value))
    {
        min_child = BHEAP_RCHILD(cur_idx);
    }
    
    tempkey = b->items[min_child].key;
    tempvalue = b->items[min_child].value;
    
    b->items[min_child].key = b->items[cur_idx].key;
    b->items[min_child].value = b->items[cur_idx].value;
    
    b->items[cur_idx].key = tempkey;
    b->items[cur_idx].value = tempvalue;
    
    cur_idx = min_child;
  }
  
  return true;
}

void
_bheap_output_indent(FILE *ofp, uint level)
{
  uint i;
  for (i=0; i<level; i++)
    fprintf(ofp, "  ");
}

void
_bheap_output_node(FILE *ofp, bheap_t *b, uint idx, uint indent,
    bheap_dot_key keyfunc)
{
  char buffer[1024];
  
  if (b == NULL)
    return;
  if (b->items == NULL)
    return;
  if (idx >= b->next)
    return;
  
  keyfunc((char *)&buffer, sizeof(buffer), b->items[idx].key,
      b->items[idx].value);
  
  /* Output the representation of the node idx */
  _bheap_output_indent(ofp, indent);
  fprintf(ofp, "%u [label=\"%s\"];\n", idx, (char *)&buffer);
  
  if (BHEAP_LCHILD(idx) < b->next)
  {
    uint lidx = BHEAP_LCHILD(idx);
    
    _bheap_output_indent(ofp, indent);
    /* Output the connection between current and left child */
    fprintf(ofp, "%u -> %u;\n", idx, lidx);
    
    /* Children */
    _bheap_output_node(ofp, b, lidx, indent+1, keyfunc);
  }
  
  if (BHEAP_RCHILD(idx) < b->next)
  {
    uint ridx = BHEAP_RCHILD(idx);
    
    _bheap_output_indent(ofp, indent);
    /* Output the connection between current and left child */
    fprintf(ofp, "%u -> %u;\n", idx, ridx);
    
    /* Children */
    _bheap_output_node(ofp, b, ridx, indent+1, keyfunc);
  }
}

/**
 * Outputs a graph diagram of the binary heap in DOT format, for visualization
 * using GraphViz.
 * @param b: The binary heap to output
 * @param outfile: The filename of the file to output the DOT graph to.
 * @param keyfunc: A function that is used to generate the string that will be
 *     printed for each node in the graph.
 * @return: true on success, false if outfile could not be opened for writing
 *     or the bheap is invalid
 */
bool
bheap_output_dot(bheap_t *b, char const *outfile, bheap_dot_key keyfunc)
{
  FILE *ofp;
  
  if (b == NULL)
    return false;
  
  ofp = fopen(outfile, "w");
  if (ofp == NULL)
    return false;
  
  fprintf(ofp, "digraph BST {\n");
  fprintf(ofp, "  node [fontname=\"Arial\"];\n");
  
  /* Empty tree */
  if (bheap_size(b) == 0)
    fprintf(ofp, "\n");
  
  /* Tree with single node */
  else if (BHEAP_RCHILD(0) >= b->next && BHEAP_LCHILD(0) >= b->next)
  {
    char buffer[1024];
    keyfunc((char *)&buffer, sizeof(buffer), b->items[0].key,
        b->items[0].value);
    
    fprintf(ofp, "  %u [label=\"%s\"];\n", b->items[0].key, (char *)&buffer);
    fprintf(ofp, "  %u;\n", b->items[0].key);
  }
  
  /* Regular tree */
  else
    _bheap_output_node(ofp, b, 0, 1, keyfunc);
  
  fprintf(ofp, "}\n");
  fclose(ofp);
  
  return true;
}

