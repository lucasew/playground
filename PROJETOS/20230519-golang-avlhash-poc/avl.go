package avlhash

import (
	"sync"

	"golang.org/x/exp/constraints"
)

// AVLTree represents a thread-safe AVL tree implementation using generics.
// It uses a mutex to ensure thread safety for concurrent access.
//
// The AVL tree is a self-balancing binary search tree. In an AVL tree, the heights
// of the two child subtrees of any node differ by at most one; if at any time they
// differ by more than one, rebalancing is done to restore this property.
// This ensures that lookups, insertions, and deletions take O(log n) time in both
// average and worst cases, where n is the number of nodes in the tree prior to the operation.
type AVLTree[AVLKey constraints.Ordered, AVLValue any] struct {
	sync.Mutex
	root *AVLNode[AVLKey, AVLValue]
}

// Add inserts a key-value pair into the AVL tree.
// If the key already exists, its value is updated.
// This operation is thread-safe and has a time complexity of O(log n).
func (t *AVLTree[AVLKey, AVLValue]) Add(key AVLKey, value AVLValue) {
	t.Lock()
	defer t.Unlock()
	t.root = t.root.add(key, value)
}

// Remove deletes a node with the specified key from the AVL tree.
// If the key does not exist, the tree remains unchanged.
// This operation is thread-safe and has a time complexity of O(log n).
func (t *AVLTree[AVLKey, AVLValue]) Remove(key AVLKey) {
	t.Lock()
	defer t.Unlock()
	t.root = t.root.remove(key)
}

// Search finds the value associated with the given key.
// It returns a pointer to the value if found, or nil if the key does not exist.
// This operation is thread-safe and has a time complexity of O(log n).
func (t *AVLTree[AVLKey, AVLValue]) Search(key AVLKey) *AVLValue {
	t.Lock()
	defer t.Unlock()
	return t.root.search(key)
}

// Clear removes all nodes from the AVL tree, making it empty.
// This operation is thread-safe.
func (t *AVLTree[AVLKey, AVLValue]) Clear() {
	t.Lock()
	defer t.Unlock()
	t.root = nil
}

// AVLNode represents a node in the AVL tree.
// It holds the key, value, references to left and right children, and the height of the node.
type AVLNode[AVLKey constraints.Ordered, AVLValue any] struct {
	key    AVLKey
	value  AVLValue
	left   *AVLNode[AVLKey, AVLValue]
	right  *AVLNode[AVLKey, AVLValue]
	height int
}

// add inserts a new node or updates an existing one in the subtree rooted at n.
// It returns the new root of the subtree after rebalancing.
func (n *AVLNode[AVLKey, AVLValue]) add(key AVLKey, value AVLValue) *AVLNode[AVLKey, AVLValue] {
	if n == nil {
		return &AVLNode[AVLKey, AVLValue]{
			key:    key,
			value:  value,
			left:   nil,
			right:  nil,
			height: 1,
		}
	}
	if key < n.key {
		n.left = n.left.add(key, value)
	} else if key > n.key {
		n.right = n.right.add(key, value)
	} else {
		n.value = value
	}
	return n.rebalanceTree()
}

// remove deletes a node with the specified key from the subtree rooted at n.
// It returns the new root of the subtree after rebalancing.
func (n *AVLNode[AVLKey, AVLValue]) remove(key AVLKey) *AVLNode[AVLKey, AVLValue] {
	if n == nil {
		return nil
	}
	if key < n.key {
		n.left = n.left.remove(key)
	} else if key > n.key {
		n.right = n.right.remove(key)
	} else {
		// Node found
		if n.left != nil && n.right != nil {
			// Node has two children: Find the smallest node in the right subtree (successor),
			// replace current node's content with successor's content, and remove successor.
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			n.value = rightMinNode.value
			n.right = n.right.remove(rightMinNode.key)
		} else if n.left != nil {
			// Node has only left child
			n = n.left
		} else if n.right != nil {
			// Node has only right child
			n = n.right
		} else {
			// Node is a leaf
			n = nil
			return n
		}
	}
	return n.rebalanceTree()
}

// search looks for a key in the subtree rooted at n.
func (n *AVLNode[AVLKey, AVLValue]) search(key AVLKey) *AVLValue {
	if n == nil {
		return nil
	}
	if key < n.key {
		return n.left.search(key)
	} else if key > n.key {
		return n.right.search(key)
	} else {
		return &n.value
	}
}

// getHeight returns the height of the node, handling nil nodes (height 0).
func (n *AVLNode[AVLKey, AVLValue]) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

// recalculateHeight updates the height of the node based on its children's heights.
func (n *AVLNode[AVLKey, AVLValue]) recalculateHeight() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

// rebalanceTree checks the balance factor of the node and performs rotations if necessary
// to restore the AVL property.
func (n *AVLNode[AVLKey, AVLValue]) rebalanceTree() *AVLNode[AVLKey, AVLValue] {
	if n == nil {
		return nil
	}
	n.recalculateHeight()
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor == -2 {
		// Right heavy
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	} else if balanceFactor == 2 {
		// Left heavy
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	}
	return n
}

// rotateLeft performs a left rotation on the subtree rooted at n.
// This is used when the right subtree is too tall.
func (n *AVLNode[AVLKey, AVLValue]) rotateLeft() *AVLNode[AVLKey, AVLValue] {
	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// rotateRight performs a right rotation on the subtree rooted at n.
// This is used when the left subtree is too tall.
func (n *AVLNode[AVLKey, AVLValue]) rotateRight() *AVLNode[AVLKey, AVLValue] {
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// findSmallest finds the node with the minimum key in the subtree.
func (n *AVLNode[AVLKey, AVLValue]) findSmallest() *AVLNode[AVLKey, AVLValue] {
	if n.left == nil {
		return n
	}
	return n.left.findSmallest()
}

// max returns the larger of two values.
func max[Value constraints.Ordered](x, y Value) Value {
	if x > y {
		return x
	} else {
		return y
	}
}
