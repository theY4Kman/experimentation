package trie

import (
	"fmt"
	"strings"
)


type trieNode struct {
	symbol string
	value  interface{}
	next   *trieNode
	child  *trieNode
}


type Trie struct {
	root trieNode
}

func NewTrie() Trie {
	return Trie{}
}

func (tr *Trie) Set(k string, v interface{}) {
	start := 0

	Traversal:
	for node := &tr.root; node != nil; {
		s := k[start:]
		for i := 0; i < len(s) && i < len(node.symbol); i++ {
			kc, nc := s[i], node.symbol[i]
			if kc < nc {

				// key is lex. lower than symbol -- split node into children, us first
				// e.g. "car" into {"cat"}
				child_b := trieNode{
					symbol: node.symbol[i:],
					value: node.value,
					child: node.child,
				}

				child_a := trieNode{
					symbol: s[i:],
					value: v,
					next: &child_b,
				}

				node.child = &child_a
				node.symbol = node.symbol[:i]
				return
			} else if kc > nc {
				if node.next == nil {
					// there is no next node -- split node into children, us last
					// e.g. "cat" into {"car"}
					child_b := trieNode{
						symbol: s[i:],
						value: v,
					}

					child_a := trieNode{
						symbol: node.symbol[i:],
						value: node.value,
						next: &child_b,
						child: node.child,
					}

					node.child = &child_a
					node.symbol = node.symbol[:i]
					return
				}
				// e.g. "cat" into {"cad", "car"}
				node = node.next
				continue Traversal
			}
		}

		// if control reaches here, the symbol at least partially matched
		if len(s) > len(node.symbol) {
			// key is longer than symbol -- traverse children, or add new child
			if node.child != nil {
				// node has children -- continue traversal
				// e.g. "carts" into {"car" -> "t"}
				start += len(node.symbol)
				node = node.child
				continue Traversal
			} else {
				// node has no children -- create single new child node
				// e.g. "cart" into {"car"}
				node.child = &trieNode{
					symbol: s[len(node.symbol):],
					value: v,
				}
				return
			}
		} else if len(s) < len(node.symbol) {
			// key is shorter than symbol -- add a child, set the node value
			// e.g. "car" into {"cart"}
			child := trieNode{
				symbol: node.symbol[len(s):],
				value: node.value,
				child: node.child,
			}

			node.symbol = s
			node.value = v
			node.child = &child
			return
		} else {
			// symbol matched completely -- simply overwrite node value
			// e.g. "car" into {"car"}
			node.value = v
			return
		}
	}
}

func (tr *Trie) Get(k string) interface{} {
	start := 0

	Traversal:
	for node := tr.root.child; node != nil; {
		s := k[start:]
		for i := 0; i < len(s) && i < len(node.symbol); i++ {
			kc, nc := s[i], node.symbol[i]
			if kc < nc {
				break Traversal
			} else if kc > nc {
				node = node.next
				continue Traversal
			}
		}

		// if control reaches here, the symbol at least partially matched
		if len(s) > len(node.symbol) {
			if node.child != nil {
				start += len(node.symbol)
				node = node.child
				continue Traversal
			} else {
				break Traversal
			}
		} else if len(s) > len(node.symbol) {
			break Traversal
		} else {
			return node.value
		}
	}

	return nil
}

func (tr *Trie) Contains(k string) bool {
	return tr.Get(k) != nil
}

func (tr *Trie) Delete(k string) {
	// Fucking TODO
}

func debugPrint(node *trieNode, indent string) {
	for sibling := node; sibling != nil; sibling = sibling.next {
		fmt.Printf("%-30s%s\n", indent + node.symbol, node.value)
		if sibling.child != nil {
			debugPrint(sibling.child, indent + strings.Repeat(" ", len(node.symbol)))
		}
	}
}

// TODO: remove when implementation complete
func DebugPrint(tr *Trie) {
	fmt.Println("#root")
	debugPrint(tr.root.child, "  ")
}
