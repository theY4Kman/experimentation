/**
 * =============================================================================
 * YakHash
 *   A hashtable! (Using linked lists)
 * Copyright (C) 2009 Zach "theY4Kman" Kanzler
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

#include <malloc.h>
#include <new>
#include <cstring>
#include <assert.h>

typedef unsigned int uint;

template<typename E>
class YakHash
{
public:
    YakHash() : m_iNumNodes(50)
    {
        _initial_allocate();
    }
    
    YakHash(uint numNodes) : m_iNumNodes(numNodes)
    {
        _initial_allocate();
    }

private:
    enum YakHashNodeType
    {
        NodeType_Element = 0,
        NodeType_Bucket,
    };
    
    struct YakHashNode
    {
        YakHashNodeType type;
        char const *key;
        E value;
        bool valset;
        /* If this is a bucket, this will point to the first element contained */
        YakHashNode *next;
    };


public:
    void put(char const *key, const E& value)
    {
        YakHashNode *node = _get_node(key);
        node->value = value;
        node->valset = value;
    }
    
    E *get(char const *key)
    {
        YakHashNode *node = _get_node(key, false);
        if (node == NULL)
            return NULL;
        
        return &node->value;
    }

private:
    inline void _initial_allocate()
    {
        uint size = sizeof(YakHashNode*) * m_iNumNodes;
        m_Nodes = (YakHashNode **)malloc(size);
        memset(m_Nodes, NULL, size);
    }
    
    /** Converts a string to an integer from 0 .. m_iNumNodes-1 */
    uint hash(char const *key)
    {
        uint idx = 1;
        while ((idx *= *key) && *(++key) != '\0');
        
        return idx % (m_iNumNodes - 1);
    }
    
    /**
     * @brief Returns the element node for the specified key
     * If no node exists for that key, it creates a new node.
     *
     * @return A YakHashNode of type NodeType_Element
     */
    YakHashNode *_get_node(char const *key, bool create=true)
    {
        uint idx = hash(key);
        YakHashNode *node = m_Nodes[idx];
        if (node == NULL)
        {
            if (!create)
                return NULL;
            
            node = new YakHashNode;
            node->type = NodeType_Element;
            node->key = _strdup(key);
            node->valset = false;
            node->next = NULL;
            
            m_Nodes[idx] = node;
            
            return node;
        }
        
        if (node->type == NodeType_Element)
        {
            /* The keys match, we want this node */
            if (strcmp(key, node->key) == 0)
                return node;
            
            if (!create)
                return NULL;
            
            /* The keys do not match, so we need to convert this node into a
             * bucket, placing itself and a new node of |key| inside it.
             */
            YakHashNode *bucket = new YakHashNode;
            bucket->type = NodeType_Bucket;
            bucket->key = NULL; // Just to be safe
            bucket->valset = true;
            bucket->next = node;
            
            m_Nodes[idx] = bucket;
            
            YakHashNode *newnode = new YakHashNode;
            newnode->type = NodeType_Element;
            newnode->key = _strdup(key);
            newnode->valset = false;
            newnode->next = NULL;
            
            node->next = newnode;
            
            return newnode;
        }
        
        if (node->type == NodeType_Bucket)
        {
            /* Search through the elements of the bucket until we find a match */
            YakHashNode *elem = node;
            do
            {
                elem = elem->next;
                if (strcmp(key, elem->key) == 0)
                    return elem;
            } while (elem->next != NULL);
            
            if (!create)
                return NULL;
            
            /* If no match is found, create a new element and tack it on the list */
            YakHashNode *newnode = new YakHashNode;
            newnode->type = NodeType_Element;
            newnode->key = _strdup(key);
            newnode->valset = false;
            newnode->next = NULL;
            
            elem->next = newnode;
            
            return newnode;
        }
        
        /* We should never reach this point */
        assert(false);
        
        return NULL;
    }
    
    char *_strdup(char const *string)
    {
        return strcpy((char *)malloc(sizeof(char) * strlen(string)), string);
    }

private:
    YakHashNode **m_Nodes;
    uint m_iNumNodes;
};

