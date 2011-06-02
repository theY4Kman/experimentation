/**
 * =============================================================================
 *   Trie. Yak Trie.
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
#include <string.h>

#ifndef NULL
#  define NULL ((void *)0)
#endif
#ifndef true
#  define true 1
#endif
#ifndef false
#  define false 0
#endif

typedef char bool;
typedef unsigned int uint;

struct trienode;
typedef struct trienode
{
  bool has_value;
  void *value;
  char const *str;
  uint strlen;
  struct trienode *children;
  struct trienode *next;
} trienode_t;

typedef struct trie
{
  trienode_t *root;
  uint size;
} trie_t;

/**
 * Creates a new trie structure
 */
trie_t *
trie_create()
{
  return calloc(1, sizeof(trie_t));
}

void
_trie_free(trienode_t *root)
{
  trienode_t *cur;
  trienode_t *temp;
  
  if (root == NULL)
    return;
  
  if (root->children != NULL)
    _trie_free(root->children);
  
  cur = root->next;
  while (cur)
  {
    temp = cur->next;
    free((void *)cur->str);
    free(cur);
    cur = temp;
  }
}

/**
 * Frees a trie structure
 * @param t: The trie structure to free
 */
void
trie_free(trie_t *t)
{
  if (t == NULL)
    return;
  
  _trie_free(t->root);
}

/**
 * Internal trie_find function. If closest is true, also returns the closest
 * node to `key` (NULL signifies root is closest). fkey is written with the
 * place in the `key` that was checked against the closest node.
 */
trienode_t *
_trie_find(trienode_t *root, char const *key, bool closest, char const **fkey)
{
  trienode_t *node;
  uint keylen = strlen(key);
  if (root == NULL)
    return NULL;
  
  node = root;
  while (node)
  {
    /* Compare our key up to the length of the current node's key */
    int cmp = strncmp(node->str, key, node->strlen);
    
    /* A match */
    if (cmp == 0)
    {
      /* An exact match */
      if (keylen == node->strlen)
      {
        if (!node->has_value)
          return NULL;
        
        if (fkey != NULL)
          *fkey = key;
        return node;
      }
      
      /* We've matched, but we have more characters in our key */
      else if (keylen > node->strlen)
      {
        /* Continue matching using the rest of the characters */
        trienode_t *result =  _trie_find(node->children, &key[node->strlen],
            closest, NULL);
        
        /* Found a result in our children */
        if (result)
          return result;
        
        else if (closest)
        {
          if (fkey != NULL)
            *fkey = key;
          /* Found no match, so we'll return the last node to match our key */
          return node;
        }
        else
          return NULL;
      }
    }
    
    /* Next sibling */
    node = node->next;
  }
  
  return NULL;
}

/**
 * Inserts `key` into the trie, associated with `value`
 * @param t: The trie structure
 * @param key: The string to associate the value with.
 * @param value: The value that will be returned when key is found
 * @param replace: Whether to replace if the `key` already exists in the trie
 * @return: true on success, false on invalid trie structure or existing key
 *     found and replace is false.
 */
bool
trie_insert(trie_t *t, char const *key, void *value, bool replace)
{
  trienode_t *closest;
  char const *fkey;
  if (t == NULL)
    return false;
  
  /* Obviously, the toughest case: the tree is empty */
  if (t->root == NULL)
  {
    trienode_t *root = (trienode_t *)calloc(1, sizeof(trienode_t));
    root->has_value = true;
    root->value = value;
    root->str = strdup(key);
    root->strlen = strlen(root->str);
    
    t->root = root;
    t->size = 1;
    
    return true;
  }
  
  closest = _trie_find(t->root, key, true, &fkey);
  if (closest != NULL)
  {
    trienode_t **temp;
    trienode_t *node;
    
    char *c = (char *)closest->str;
    char *f = (char *)fkey;
    while (*c != '\0' && *f != '\0' && *c == *f)
    {
      c++;
      f++;
    }
    
    /* Found key that exactly matches fkey */
    if (*c == *f)
    {
      /* The node already has a value. */
      if (closest->has_value)
      {
        if (replace)
        {
          closest->value = value;
          return true;
        }
        
        return false;
      }
      
      /* The node is purely a branching node and has no value */
      else
      {
        closest->has_value = true;
        closest->value = value;
        t->size++;
        return true;
      }
    }
    
    /* The comparison between our new key and the closest node's key ended mid
     * compare, meaning we can split the closest node's key. For example, this
     * case will happen when comparing our new key "tardis" against the closest
     * node's key "tarbaby". This case would cut off the closest node's key to
     * "tar" and create two children, "dis" and "baby"
     */
    if (strlen(closest->str) > (c - closest->str))
    {
      trienode_t *new_nodes = (trienode_t *)calloc(2, sizeof(trienode_t));
      
      /* A new node split off from the closest node's key */
      node = &new_nodes[0];
      node->has_value = closest->has_value;
      node->value = closest->value;
      node->str = strdup(c);
      node->strlen = strlen(node->str);
      node->children = closest->children;
      
      /* Cut off the key from the closest node at the point of separation */
      *c = '\0';
      closest->has_value = false;
      closest->value = NULL;
      closest->strlen = strlen(closest->str);
      
      /* Create our new node from `key` */
      node = &new_nodes[1];
      node->has_value = true;
      node->value = value;
      node->str = strdup(f);
      node->strlen = strlen(node->str);
      node->next = &new_nodes[0];
      
      closest->children = node;
      
      return true;
    }
    
    /* Add to the closest node's siblings. */
    else
    {
      temp = &closest->children;
      while (*temp != NULL)
        temp = &(*temp)->next;
      
      node = (trienode_t *)calloc(1, sizeof(trienode_t));
      node->has_value = true;
      node->value = value;
      node->str = strdup(&fkey[closest->strlen]);
      node->strlen = strlen(node->str);
      
      *temp = node;
      return true;
    }
  }
  
  /* Add to root node siblings */
  else
  {
    trienode_t *temp;
    trienode_t *node = (trienode_t *)calloc(1, sizeof(trienode_t));
    node->has_value = true;
    node->value = value;
    node->str = strdup(key);
    node->strlen = strlen(node->str);
    
    temp = t->root;
    while (temp->next)
      temp = temp->next;
    temp->next = node;
    
    t->size++;
    
    return true;
  }
}

/**
 * Finds `key` in the trie and gives back its value
 * @param t: The trie structure
 * @param key: The key to search for
 * @param value: A pointer to a place to store the key's value
 * @return: true if key found, false on invalid trie structure or key not found
 */
bool
trie_find(trie_t *t, char const *key, void **value)
{
  trienode_t *node;
  if (t == NULL || t->root == NULL)
    return false;
  
  node = _trie_find(t->root, key, false, NULL);
  if (node == NULL)
    return false;
  
  *value = node->value;
  return true;
}

