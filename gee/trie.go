package gee

import (
	"fmt"
	"strings"
)

type node struct {
	pattern  string  // 待匹配路由
	part     string  //路由一部分 如 :lang
	children []*node // 子节点
	isWild   bool    // 是否精确匹配  如 part含有 :  或者 * 时，为true
}

func (n *node) String() string {
	return fmt.Sprintf("node{pattern = %s, part = %s, isWild = %t}", n.pattern, n.part, n.isWild)
}

func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil { // 找不到就 创建
		child = &node{part: part, isWild: part[0] == ':' || part[0] == '*'}
		n.children = append(n.children, child)
	}

	child.insert(pattern, parts, height+1)
}

// 找到 parts 最后一个节点， 即路径的最后一个节点
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	// 找到 n的子节点中 所有 part = parts[height] 的 节点
	children := n.matchChildren(part)

	for _, child := range children {
		// 递归上一步的节点集合中的 每一个节点
		result := child.search(parts, height+1)

		if result != nil {
			return result
		}
	}

	return nil
}

//  list存放 每个 pattern 的 最后节点 因为insert()中，每个pattern的最后一个节点才 n.pattern = pattern
func (n *node) travel(list *([]*node)) {
	if n.pattern != "" {
		*list = append(*list, n)
	}

	for _, child := range n.children {
		child.travel(list)
	}
}

// 第一个匹配成功的节点， 用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}

	return nil
}

// 所有匹配成功的节点， 用于查询
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)

	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}

	return nodes
}
