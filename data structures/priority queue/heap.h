/**
 * =============================================================================
 * Priority Queue
 *   An implementation of a priority queue using a heap structure. Holds a
 *   priority and value for each item.
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

template<class K>
class HeapPriorityQueue
{
private:
    struct node
    {
        /// Integer value of the node
        int priority;
        /// A value to store with the node
        K* value;
    };
    
    /// Our heap/array.
    node *m_pHeap;
    /// Length of our array/heap.
    unsigned int m_iMax;
    /// The largest used index in the heap at the moment
    unsigned int m_iN;
    
    /**
     * Uses the passed index to sort the item up the tree.
     * @param   k: index to the internal array
     */
    void upheap(unsigned int k)
    {
        if(k > m_iN) return;
        
        unsigned int j = k;
        node val = m_pHeap[j];
        
        while((k /= 2) > 0 && val.priority > m_pHeap[k].priority)
        {
            m_pHeap[j] = m_pHeap[k];
            m_pHeap[k] = val;
            j = k;
        }
    };
    
    /**
     * Uses the passed index to sort the item down the tree.
     * @param   k: index to the internal array
     */
    void downheap(unsigned int k)
    {
        if(k > m_iN) return;
        
        unsigned int j;
        node val = m_pHeap[k];
        
        unsigned int limit = m_iN / 2;
        
        while(k <= limit)
        {
            j = k + k;
            
            // Move to the right child if it is larger than the left child
            if(j < m_iN && m_pHeap[j].priority < m_pHeap[j+1].priority) j++;
            
            // If `val` is larger than both children of `k`, we're done.
            if(val.priority >= m_pHeap[j].priority) break;
            
            m_pHeap[k] = m_pHeap[j];
            k = j;
        }
        
        m_pHeap[k] = val;
    };

public:
    /**
     * @param   max: The amount of items to be stored in the queue.
     */
    HeapPriorityQueue(unsigned int max) : m_iMax(max), m_iN(0)
    {
        m_pHeap = new node[m_iMax];
    };
    
    ~HeapPriorityQueue()
    {
        delete [] m_pHeap;
    };
    
    /**
     * Insert an item into the queue.
     * @param   priority: Priority of the value to add.
     * @param   val: Value to add.
     */
    void insert(int priority, const K& val)
    {
        if(m_iN >= m_iMax-1) return;
        
        node item;
        item.priority = priority;
        item.value = (K*)&val;
        
        m_pHeap[++m_iN] = item;
        upheap(m_iN);
    }
    
    /**
     * Removes the largest item from the queue and returns it.
     * @return: Value of the item with the highest priority.
     */
    K* remove()
    {
        if(m_iN <= 1) return NULL;
        
        K *val = m_pHeap[1].value;
        m_pHeap[1] = m_pHeap[m_iN--];
        downheap(1);
        
        return val;
    }
};
