package avlhash

import (
	"golang.org/x/exp/constraints"
	"sync"
)

type AVLTree[AVLKey constraints.Ordered, AVLValue any] struct {
	sync.Mutex
	root *AVLNode[AVLKey, AVLValue]
}

func (t *AVLTree[AVLKey, AVLValue]) Add(key AVLKey, value AVLValue) {
	t.Lock()
	defer t.Unlock()
	t.root = t.root.add(key, value)
}

func (t *AVLTree[AVLKey, AVLValue]) Remove(key AVLKey) {
	t.Lock()
	defer t.Unlock()
	t.root = t.root.remove(key)
}

func (t *AVLTree[AVLKey, AVLValue]) Search(key AVLKey) *AVLValue {
	t.Lock()
	defer t.Unlock()
	return t.root.search(key)
}

func (t *AVLTree[AVLKey, AVLValue]) Clear() {
	t.Lock()
	defer t.Unlock()
	t.root = nil
}

type AVLNode[AVLKey constraints.Ordered, AVLValue any] struct {
	key    AVLKey
	value  AVLValue
	left   *AVLNode[AVLKey, AVLValue]
	right  *AVLNode[AVLKey, AVLValue]
	height int
}

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

func (n *AVLNode[AVLKey, AVLValue]) remove(key AVLKey) *AVLNode[AVLKey, AVLValue] {
	if n == nil {
		return nil
	}
	if key < n.key {
		n.left = n.left.remove(key)
	} else if key > n.key {
		n.right = n.right.remove(key)
	} else {
		if n.left != nil && n.right != nil {
			rightMinNode := n.right.findSmallest()
			n.key = rightMinNode.key
			n.value = rightMinNode.value
			n.right = n.right.remove(rightMinNode.key)
		} else if n.left != nil {
			n = n.left
		} else if n.right != nil {
			n = n.right
		} else {
			n = nil
			return n
		}
	}
	return n.rebalanceTree()
}

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

func (n *AVLNode[AVLKey, AVLValue]) getHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *AVLNode[AVLKey, AVLValue]) recalculateHeight() {
	n.height = 1 + max(n.left.getHeight(), n.right.getHeight())
}

func (n *AVLNode[AVLKey, AVLValue]) rebalanceTree() *AVLNode[AVLKey, AVLValue] {
	if n == nil {
		return nil
	}
	n.recalculateHeight()
	balanceFactor := n.left.getHeight() - n.right.getHeight()
	if balanceFactor == -2 {
		if n.right.left.getHeight() > n.right.right.getHeight() {
			n.right = n.right.rotateRight()
		}
		return n.rotateLeft()
	} else if balanceFactor == 2 {
		// check if child is right-heavy and rotateLeft first
		if n.left.right.getHeight() > n.left.left.getHeight() {
			n.left = n.left.rotateLeft()
		}
		return n.rotateRight()
	}
	return n
}

func (n *AVLNode[AVLKey, AVLValue]) rotateLeft() *AVLNode[AVLKey, AVLValue] {
	newRoot := n.right
	n.right = newRoot.left
	newRoot.left = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

func (n *AVLNode[AVLKey, AVLValue]) rotateRight() *AVLNode[AVLKey, AVLValue] {
	newRoot := n.left
	n.left = newRoot.right
	newRoot.right = n

	n.recalculateHeight()
	newRoot.recalculateHeight()
	return newRoot
}

// Finds the smallest child (based on the key) for the current node
func (n *AVLNode[AVLKey, AVLValue]) findSmallest() *AVLNode[AVLKey, AVLValue] {
	if n.left == nil {
		return n
	}
	return n.left.findSmallest()
}

func max[Value constraints.Ordered](x, y Value) Value {
	if x > y {
		return x
	} else {
		return y
	}
}
