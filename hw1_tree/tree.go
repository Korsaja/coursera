package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := tree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func tree(w io.Writer, path string, show bool) error {

	nodes, err := read(path, []N{}, show)

	printTree(w,nodes,[]string{})

	return err

}

func printTree(w io.Writer, n []N, p []string) {
	if len(n) == 0{
		return
	}
	fmt.Fprintf(w, "%s", strings.Join(p, ""))

	node := n[0]

	if len(n) == 1{
		fmt.Fprintf(w, "%s%s\n", "└───", node)
		if dir, ok := node.(D);ok{printTree(w,dir.child,append(p,"\t"))}
		return
	}


	fmt.Fprintf(w, "%s%s\n", "├───", node)
	if dir, ok := node.(D);ok{
		printTree(w,dir.child,append(p,"│\t"))
	}

	printTree(w,n[1:],p)
}

func read(path string, nodes []N, showFiles bool) ([]N, error) {

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if !(f.IsDir() || showFiles) {
			continue
		}

		var n N
		if f.IsDir() {
			nodes, _ := read(filepath.Join(path, f.Name()), []N{}, showFiles)
			n = D{f.Name(), nodes}
		} else {
			n = F{f.Name(), f.Size()}
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

type N interface{ fmt.Stringer }

//dir
type D struct {
	name  string
	child []N
}

func (d D) String() string {
	return d.name
}

//file
type F struct {
	name string
	size int64
}

func (f F) String() string {
	if f.size == 0 {
		return fmt.Sprintf("%s (empty)",f.name)
	}
	return fmt.Sprintf("%s (%db)", f.name, f.size)
}
