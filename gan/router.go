package gan

import (
	"fmt"
	"strings"
)

type trie struct {
	root *node
}

type router struct {
	roots    map[string]*trie
	handlers map[string]HandlerFunc
}

type node struct {
	pattern  string // only nonempty at leaf
	children map[string]*node
	wild     bool //true if starts with : or *
}

func newRouter() *router {
	return &router{
		roots:    make(map[string]*trie),
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *router) addRoute(method, pattern string, handler HandlerFunc) {
	if _, ok := r.roots[method]; !ok {
		r.roots[method] = &trie{root: &node{children: make(map[string]*node)}}
	}
	r.roots[method].insert(pattern)
	key := method + "-" + pattern
	r.handlers[key] = handler
}

func (r *router) getRoute(method, path string) (*node, map[string]string) {
	params := make(map[string]string)
	root, ok := r.roots[method]
	if !ok {
		return nil, nil
	}
	fmt.Printf("path: %v\n", path)
	n := root.search(path)
	fmt.Println(n)
	if n == nil {
		return nil, nil
	}
	parts := parsePattern(n.pattern)
	curParts := parsePattern(path)
	for i, part := range parts {
		if part[0] == ':' {
			params[part[1:]] = curParts[i]
		} else if part[0] == '*' {
			params[part[1:]] = strings.Join(curParts[i:], "/")
		}
	}
	return n, params
}

func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(404, "404 NOT FOUND: %v\n\n", c.Path)
		})
	}
	c.Next()
}

func parsePattern(pattern string) []string {
	parts := strings.Split(pattern, "/")
	ret := make([]string, 0)
	for _, part := range parts {
		if part != "" {
			ret = append(ret, part)
			if part[0] == '*' {
				break
			}
		}
	}
	return ret
}

func (t *trie) insert(pattern string) {
	parts := parsePattern(pattern)
	n := t.root
	for _, part := range parts {
		if n.children[part] == nil {
			n.children[part] = &node{
				children: make(map[string]*node),
				wild:     part[0] == ':' || part[0] == '*',
			}
		}
		n = n.children[part]
	}
	n.pattern = pattern
}

func (t *trie) search(pattern string) *node {
	searchParts := parsePattern(pattern)
	if len(searchParts) == 0 {
		if t.root.pattern != "" {
			return t.root
		}
		return nil
	}
	return doSearch(t.root, searchParts, 0)
}

func doSearch(n *node, parts []string, idx int) *node {
	var namedResult, wildResult, normalResult *node
	for part, child := range n.children {
		if child.wild || part == parts[idx] {
			if child.pattern != "" {
				switch {
				case strings.HasPrefix(part, ":"):
					namedResult = child
				case strings.HasPrefix(part, "*"):
					wildResult = child
				case idx == len(parts)-1:
					normalResult = child
				}
				if normalResult != nil {
					return normalResult
				}
			}
			res := doSearch(child, parts, idx+1)
			if res != nil {
				return res
			}
		}
	}
	if namedResult != nil {
		return namedResult
	}
	if wildResult != nil {
		return wildResult
	}
	return nil
}
