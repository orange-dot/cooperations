// File: main.go
package main

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound = errors.New("not found")
)

// Ordered is a minimal constraint for types that can be ordered.
// (Avoids external deps like golang.org/x/exp/constraints)
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

type node[T Ordered] struct {
	value       T
	left, right *node[T]
}

type BST[T Ordered] struct {
	root *node[T]
	size int
}

// NewBST creates an empty binary search tree.
func NewBST[T Ordered]() *BST[T] {
	return &BST[T]{}
}

func (t *BST[T]) Size() int {
	return t.size
}

func (t *BST[T]) IsEmpty() bool {
	return t.size == 0
}

// Insert inserts v into the tree. If v already exists, it does nothing and returns false.
func (t *BST[T]) Insert(v T) bool {
	if t.root == nil {
		t.root = &node[T]{value: v}
		t.size = 1
		return true
	}

	cur := t.root
	for {
		if v == cur.value {
			return false
		}
		if v < cur.value {
			if cur.left == nil {
				cur.left = &node[T]{value: v}
				t.size++
				return true
			}
			cur = cur.left
			continue
		}
		if cur.right == nil {
			cur.right = &node[T]{value: v}
			t.size++
			return true
		}
		cur = cur.right
	}
}

// Contains reports whether v exists in the tree.
func (t *BST[T]) Contains(v T) bool {
	_, ok := t.Find(v)
	return ok
}

// Find returns v if present and ok=true; otherwise ok=false with the zero value.
func (t *BST[T]) Find(v T) (T, bool) {
	cur := t.root
	for cur != nil {
		if v == cur.value {
			return cur.value, true
		}
		if v < cur.value {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}
	var zero T
	return zero, false
}

// Min returns the minimum element.
func (t *BST[T]) Min() (T, error) {
	if t.root == nil {
		var zero T
		return zero, ErrNotFound
	}
	n := t.root
	for n.left != nil {
		n = n.left
	}
	return n.value, nil
}

// Max returns the maximum element.
func (t *BST[T]) Max() (T, error) {
	if t.root == nil {
		var zero T
		return zero, ErrNotFound
	}
	n := t.root
	for n.right != nil {
		n = n.right
	}
	return n.value, nil
}

// Delete removes v from the tree. Returns true if an element was removed.
func (t *BST[T]) Delete(v T) bool {
	var deleted bool
	t.root, deleted = deleteNode(t.root, v)
	if deleted {
		t.size--
	}
	return deleted
}

func deleteNode[T Ordered](n *node[T], v T) (*node[T], bool) {
	if n == nil {
		return nil, false
	}

	if v < n.value {
		var deleted bool
		n.left, deleted = deleteNode(n.left, v)
		return n, deleted
	}
	if v > n.value {
		var deleted bool
		n.right, deleted = deleteNode(n.right, v)
		return n, deleted
	}

	// n.value == v: delete this node
	if n.left == nil && n.right == nil {
		return nil, true
	}
	if n.left == nil {
		return n.right, true
	}
	if n.right == nil {
		return n.left, true
	}

	// Two children: replace with inorder successor (min of right subtree)
	successor := n.right
	for successor.left != nil {
		successor = successor.left
	}
	n.value = successor.value
	var deleted bool
	n.right, deleted = deleteNode(n.right, successor.value)
	// deleted must be true here
	return n, deleted
}

// InOrder returns values in ascending order.
func (t *BST[T]) InOrder() []T {
	out := make([]T, 0, t.size)
	inOrder(t.root, &out)
	return out
}

func inOrder[T Ordered](n *node[T], out *[]T) {
	if n == nil {
		return
	}
	inOrder(n.left, out)
	*out = append(*out, n.value)
	inOrder(n.right, out)
}

// PreOrder returns values in root-left-right order.
func (t *BST[T]) PreOrder() []T {
	out := make([]T, 0, t.size)
	preOrder(t.root, &out)
	return out
}

func preOrder[T Ordered](n *node[T], out *[]T) {
	if n == nil {
		return
	}
	*out = append(*out, n.value)
	preOrder(n.left, out)
	preOrder(n.right, out)
}

// PostOrder returns values in left-right-root order.
func (t *BST[T]) PostOrder() []T {
	out := make([]T, 0, t.size)
	postOrder(t.root, &out)
	return out
}

func postOrder[T Ordered](n *node[T], out *[]T) {
	if n == nil {
		return
	}
	postOrder(n.left, out)
	postOrder(n.right, out)
	*out = append(*out, n.value)
}

// Height returns the height of the tree measured in edges.
// Empty tree: -1, single node: 0.
func (t *BST[T]) Height() int {
	return height(t.root)
}

func height[T Ordered](n *node[T]) int {
	if n == nil {
		return -1
	}
	lh := height(n.left)
	rh := height(n.right)
	if lh > rh {
		return lh + 1
	}
	return rh + 1
}

// String prints the in-order traversal.
func (t *BST[T]) String() string {
	vals := t.InOrder()
	var b strings.Builder
	b.WriteString("BST{")
	for i, v := range vals {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprint(&b, v)
	}
	b.WriteString("}")
	return b.String()
}

func demoBinaryTree() {
	t := NewBST[int]()
	for _, v := range []int{5, 3, 7, 2, 4, 6, 8} {
		t.Insert(v)
	}

	fmt.Println("size:", t.Size())
	fmt.Println("in-order:", t.InOrder())
	fmt.Println("contains 4:", t.Contains(4))
	fmt.Println("contains 10:", t.Contains(10))

	min, _ := t.Min()
	max, _ := t.Max()
	fmt.Println("min:", min, "max:", max)

	fmt.Println("delete 7:", t.Delete(7))
	fmt.Println("in-order after delete:", t.InOrder())
	fmt.Println("height:", t.Height())
	fmt.Println(t.String())
}
