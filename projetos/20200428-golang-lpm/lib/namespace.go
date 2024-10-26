package definition

import "strings"

type Namespace struct {
	path []string
	i    int
}

func NewNamespace(s string) Namespace {
	return Namespace{path: strings.Split(s, "/")}
}

func (n Namespace) tail() Namespace { // Namespace without the first element
	if len(n.ScopedSlices()) <= 1 {
		panic("cant tail with 1 or less remaining elements")
	}
	return Namespace{path: n.path, i: n.i + 1}
}

func (n Namespace) ScopedPath() string {
	return strings.Join(n.path[n.i:], "/")
}

func (n Namespace) FullPath() string {
	return strings.Join(n.path, "/")
}

func (n Namespace) String() string {
	return n.FullPath()
}

func (n Namespace) ScopedSlices() []string {
	return n.path[n.i:]
}
