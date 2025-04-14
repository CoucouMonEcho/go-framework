package web

import (
	"fmt"
	"regexp"
	"strings"
)

type router struct {
	// method -> root
	trees map[string]*node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

func (r *router) addRoute(method string, path string, handlerFunc HandlerFunc, middlewares ...Middleware) {
	if path == "" {
		panic("web: empty path")
	}
	if path[0] != '/' {
		panic("web: path must begin with '/'")
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: path can not end with '/'")
	}

	root, ok := r.trees[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}
	if path == "/" {
		if root.handler != nil {
			panic("web: '/' already has a handler")
		}
		root.handler = handlerFunc
		root.route = "/"
		root.middlewares = middlewares
		return
	}
	for _, seg := range strings.Split(path, "/")[1:] {
		if seg == "" {
			panic("web: path can not contains '//'")
		}
		child := root.childOrCreate(seg)
		// do not edit node here
		root = child
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: path '%s' already exist", path))
	}
	root.handler = handlerFunc
	root.route = path
	root.middlewares = middlewares
}

func (r *router) route(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	mi := &matchInfo{
		node: root,
	}
	if path == "/" {
		return mi, mi.node.handler != nil
	}

	var pathParams map[string]string
	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segs {
		if seg == "" {
			panic("web: path can not contains '//'")
		}
		child, ok := mi.node.childOf(seg)
		if !ok {
			if mi.node.nodeType == nodeTypeWildcard {
				return mi, mi.node.handler != nil
			}
			return nil, false
		}
		if child.nodeType == nodeTypePathParam || child.nodeType == nodeTypeRegular {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			pathParams[child.path[1:]] = seg
		}
		mi.node = child
	}
	mi.pathParams = pathParams
	mi.middlewares = findMiddlewares(root, segs)
	return mi, mi.node.handler != nil
}

func findMiddlewares(root *node, segs []string) []Middleware {
	// use queue to level-order traversal
	queue := []*node{root}
	res := make([]Middleware, 0, 16)
	for i := 0; i < len(segs); i++ {
		seg := segs[i]
		var children []*node
		for _, cur := range queue {
			if len(cur.children) > 0 {
				res = append(res, cur.middlewares...)
			}
			curChildren := []*node{}
			for _, curChild := range cur.children {
				if curChild.path == seg || curChild.nodeType != nodeTypeStatic {
					curChildren = append(children, curChild)
				}
			}
			children = append(children, curChildren...)
		}
		queue = children
	}
	for _, cur := range queue {
		if len(cur.children) > 0 {
			res = append(res, cur.middlewares...)
		}
	}
	return res
}

type node struct {
	route       string
	path        string
	children    []*node
	nodeType    nodeType
	handler     HandlerFunc
	middlewares []Middleware
}

type nodeType int

const (
	nodeTypeStatic nodeType = iota
	nodeTypeRegular
	nodeTypePathParam
	nodeTypeWildcard
)

var regCompileMap map[int]*regexp.Regexp

func (n *node) childOrCreate(path string) *node {
	if n.children == nil {
		n.children = []*node{}
	}
	var needChildType nodeType
	var regIndexL int
	var regCompile *regexp.Regexp

	if regIndexL = strings.Index(path, "("); regIndexL > 0 && regIndexL < len(path)-1 && strings.HasSuffix(path, ")") {
		needChildType = nodeTypeRegular
		if c, err := regexp.Compile(path[regIndexL+1 : len(path)-1]); err != nil {
			panic("web: invalid regular expression")
		} else {
			regCompile = c
		}
	} else if path[0] == ':' {
		needChildType = nodeTypePathParam
	} else if path == "*" {
		needChildType = nodeTypeWildcard
	} else {
		needChildType = nodeTypeStatic
	}

	for index, child := range n.children {
		if nodeTypeRegular == needChildType && nodeTypeRegular == child.nodeType {
			isPathEqual := child.path == path[:regIndexL]
			isRegEqual := regCompileMap != nil && regCompileMap[index] != nil && regCompileMap[index].String() == path[regIndexL+1:len(path)-1]
			if !isPathEqual && isRegEqual || isPathEqual && !isRegEqual {
				panic("web: duplicate regular router")
			} else if isPathEqual && isRegEqual {
				return child
			}
			continue
		} else if nodeTypeStatic != needChildType && nodeTypeStatic != child.nodeType && needChildType != child.nodeType {
			panic("web: can not register too many non static nodes")
		}
		if child.nodeType == needChildType && child.path == path {
			return child
		}
	}
	if nodeTypeRegular == needChildType {
		if regCompileMap == nil {
			regCompileMap = make(map[int]*regexp.Regexp)
		}
		regCompileMap[len(n.children)] = regCompile
		path = path[:regIndexL]
	}
	child := &node{path: path, nodeType: needChildType}
	n.children = append(n.children, child)
	return child
}

func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return nil, false
	}
	nonStaticChildren := make(map[int]*node)
	for index, child := range n.children {
		if child.nodeType == nodeTypeStatic && child.path == path {
			return child, true
		}
		if child.nodeType != nodeTypeStatic {
			nonStaticChildren[index] = child
		}
	}
	for index, nonStaticChild := range nonStaticChildren {
		if nonStaticChild == nil {
			continue
		}
		if nonStaticChild.nodeType != nodeTypeRegular {
			return nonStaticChild, true
		} else {
			regCompile := regCompileMap[index]
			if regCompile.Match([]byte(path)) {
				return nonStaticChild, true
			}
		}
	}
	return nil, false
}

type matchInfo struct {
	node        *node
	pathParams  map[string]string
	middlewares []Middleware
}
