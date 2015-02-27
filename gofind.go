package main
// Author: Robert B Frangioso

import (
	"path/filepath"
	"os"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"
	"os/exec"
	"bytes"
)

var wg sync.WaitGroup
var exp string

func visit(path string, f os.FileInfo, err error) error {

	b, err := filepath.Match(exp, f.Name())
	if b == true {
		fmt.Printf("%s\n", path)
	}

	return err
}

func find(root string) error {

	defer wg.Done()
	var o bytes.Buffer

	cmd_out := &o

	cmd := exec.Command("find", root, "-name", exp, "-print")
	cmd.Stdout = cmd_out

	err := cmd.Run()

	if(cmd_out.Len() > 0) {
		fmt.Printf("%s\n", cmd_out.String())
	}
	// SLOWWWWW
	// err := filepath.Walk(root, visit)

	return err
}

func main() {

	flag.Parse()
	root := flag.Arg(0)
	exp = flag.Arg(1)

	basedirs, direrr := ioutil.ReadDir(root)

	if(direrr != nil) {
		fmt.Printf("ReadDir err %v \n", direrr)
	}

	for  dir := range basedirs {
		if basedirs[dir].IsDir() {
			wg.Add(1)
			go find(filepath.Join(root, basedirs[dir].Name()))
		}
	}

	wg.Wait()
}

