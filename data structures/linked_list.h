template<class K>
class LinkedList
{
private:
    struct listNode
    {
        K* item;
        listNode *next;
    };
    
    listNode *m_pHead;
    listNode *m_pButt;
    

public:
    LinkedList()
    {
        m_pHead = new listNode();
        m_pButt = new listNode();
        m_pButt->next = m_pButt;
        m_pHead->next = m_pButt;
    };
    
    /**
     * Deletes all elements in the linked list.
     */
    ~LinkedList()
    {
        listNode *this_node = m_pHead;
        listNode *next_node;
        
        do
        {
            next_node = this_node->next;
            delete this_node;
            this_node = next_node;
        } while(next_node != m_pButt);
        
        delete m_pButt;
    };
    
    /**
     * Prepends an item to the end of the list.
     */
    void prepend(const K& item)
    {
        listNode *first = m_pHead->next;
        
        listNode *new_node = new listNode();
        new_node->item = item;
        new_node->next = first;
        m_pHead->next = new_node;
    };
};
