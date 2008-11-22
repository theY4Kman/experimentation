/**
 * =============================================================================
 * Linked List
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
