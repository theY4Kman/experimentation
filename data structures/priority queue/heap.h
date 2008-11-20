class HeapPriorityQueue
{
private:
    /// Our heap/array.
    int *m_pHeap;
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
        int val = m_pHeap[j];
        
        while((k /= 2) > 0 && val > m_pHeap[k])
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
        int val = m_pHeap[k];
        
        unsigned int limit = m_iN / 2;
        
        while(k <= limit)
        {
            j = k + k;
            
            // Move to the right child if it is larger than the left child
            if(j < m_iN && m_pHeap[j] < m_pHeap[j+1]) j++;
            
            // If `val` is larger than both children of `k`, we're done.
            if(val >= m_pHeap[j]) break;
            
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
        m_pHeap = new int[m_iMax];
    };
    
    ~HeapPriorityQueue()
    {
        delete [] m_pHeap;
    };
    
    /**
     * Insert an item into the queue.
     * @param   val: Item to add.
     */
    void insert(int val)
    {
        if(m_iN >= m_iMax-1) return;
        
        m_pHeap[++m_iN] = val;
        upheap(m_iN);
    }
    
    /**
     * Removes the largest item from the queue and returns it.
     * @return: Largest item from the queue.
     */
    int remove()
    {
        if(m_iN <= 1) return -1;
        
        int val = m_pHeap[1];
        m_pHeap[1] = m_pHeap[m_iN--];
        downheap(1);
        
        return val;
    }
};
