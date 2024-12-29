package xmlcomparator

import (
	"bytes"
	"encoding/xml"
	"strconv"
	"strings"
)

// Abstract XML node presentation
type Node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:"-"`
	Content  []byte     `xml:",innerxml"`
	CharData string     `xml:",chardata"`
	Children []Node     `xml:",any"`
	Parent   *Node      `xml:"-"`
	Hash     int        `xml:"-"`
}

// Walks depth-first through the XML tree calling the function for iteslef and then for each child node
//   - f - function to call for each node; should return `false` to stop traversiong
func (node *Node) Walk(f func(*Node) bool) {
	if !f(node) {
		return
	}

	for i := range node.Children {
		node.Children[i].Walk(f)
	}
}

// Unmarshals XML data into a Node structure - "encoding/xml" package compatible
func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node Node

	return d.DecodeElement((*node)(n), &start)
}

// Unmarshals XML string into a Node structure
//   - xmlString - XML string to unmarshal
//
// Returns: root node of the XML tree and error if any
func UnmarshalXML(xmlString string) (*Node, error) {
	buf := bytes.NewBuffer([]byte(xmlString))
	dec := xml.NewDecoder(buf)

	var root Node
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	root.Walk(func(n *Node) bool {
		n.Hash = hashCode(n)

		for i := range n.Children {
			n.Children[i].Parent = n
		}
		return true
	})

	return &root, nil
}

// Creates a string representation of the XML path to the node.
//
// Path elements are node names separated by slashes.
//
// Child element might have its index, unless it is the only child - handy for dealing with arrays.
func (node *Node) Path() string {
	path := make([]string, 0)
	currNode := node

	for currNode.Parent != nil {
		siblings := currNode.Parent.Children
		if len(siblings) == 1 {
			path = append(path, "/"+currNode.XMLName.Local)
		} else {
			for i := 0; i < len(siblings); i++ {
				if &siblings[i] == currNode {
					path = append(path, "/"+currNode.XMLName.Local+"["+strconv.Itoa(i)+"]")
					break
				}
			}
		}
		currNode = currNode.Parent
	}
	path = append(path, "/"+currNode.XMLName.Local)

	// Why `slices.Reverse(path)` does not work?`
	size := len(path)
	for i := 0; i < size/2; i++ {
		path[i], path[size-i-1] = path[size-i-1], path[i]
	}

	return strings.Join(path, "")
}

// Converts XML node to a string that includes node name and attribites.
func (node *Node) String() string {
	attStr := ""
	for i, a := range node.Attrs {
		attStr += a.Name.Local + "=" + a.Value
		if i < len(node.Attrs)-1 {
			attStr += ", "
		}
	}

	ret := node.XMLName.Local + "[" + attStr + "]"

	if len(node.Children) == 0 {
		ret += " = " + string(node.Content)
	}

	return ret
}

// Compares two slices of comparable elements (insteaed)
//   - a - first slice
//   - b - second slice
//
// Returns: `true` if slices are identical, `false` otherwise
func slicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Finds index of an element in a slice
func findIndex[T comparable](slice []T, elem T) int {
	for i := range slice {
		if slice[i] == elem {
			return i
		}
	}
	return -1
}

//------- hash code generation -------

func hashCodeS(s string) int {
	hash := 0
	for _, c := range s {
		hash = 31*hash + int(c)
	}
	return hash
}

func hashCode(node *Node) int {
	hash := 31*hashCodeS(node.XMLName.Local) + hashCodeS(strings.TrimSpace(node.CharData))
	for _, attr := range node.Attrs {
		hash = 31 * (31*hash + hashCodeS(attr.Name.Local) + hashCodeS(attr.Value))
	}
	hash = 31*hash + len(node.Children)

	return hash
}
