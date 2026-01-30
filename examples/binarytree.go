package main

import (
	"fmt"
)

// TreeNode represents each node in the binary search tree
type TreeNode struct {
	Value int
	Left  *TreeNode
	Right *TreeNode
}

// BinarySearchTree represents the binary search tree
type BinarySearchTree struct {
	Root *TreeNode
}

// Insert adds a new value to the binary search tree
func (bst *BinarySearchTree) Insert(value int) {
	newNode := &TreeNode{Value: value}
	if bst.Root == nil {
		bst.Root = newNode
	} else {
		insertNode(bst.Root, newNode)
	}
}

// insertNode places a new node in the correct position in the tree
func insertNode(current, newNode *TreeNode) {
	if newNode.Value < current.Value {
		if current.Left == nil {
			current.Left = newNode
		} else {
			insertNode(current.Left, newNode)
		}
	} else {
		if current.Right == nil {
			current.Right = newNode
		} else {
			insertNode(current.Right, newNode)
		}
	}
}

// Search looks for a value in the binary search tree and returns true if found
func (bst *BinarySearchTree) Search(value int) bool {
	return searchNode(bst.Root, value)
}

// searchNode checks a node for the specified value
func searchNode(node *TreeNode, value int) bool {
	if node == nil {
		return false
	}
	if value < node.Value {
		return searchNode(node.Left, value)
	} else if value > node.Value {
		return searchNode(node.Right, value)
	}
	return true
}

// demoBinaryTree demonstrates insert and search functionalities.
func demoBinaryTree() {
	bst := BinarySearchTree{}
	valuesToInsert := []int{50, 30, 70, 20, 40, 60, 80}
	for _, value := range valuesToInsert {
		bst.Insert(value)
	}

	searchValues := []int{25, 70}
	for _, value := range searchValues {
		found := bst.Search(value)
		if found {
			fmt.Printf("Value %d found in BST\n", value)
		} else {
			fmt.Printf("Value %d not found in BST\n", value)
		}
	}
}
