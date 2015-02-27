package main
// Author: Robert B Frangioso

import (
	"path/filepath"
//	"os"
	"flag"
	"fmt"
	"io/ioutil"
	"sync"
	"os/exec"
	"bytes"
)

func find(root string, wg *sync.WaitGroup, exp string) error {

	defer wg.Done()
	var cmd_out bytes.Buffer

	cmd := exec.Command("find", root, "-name", exp, "-print")
	cmd.Stdout = &cmd_out
	err := cmd.Run()

	if(cmd_out.Len() > 0) {
		fmt.Printf("%s\n", cmd_out.String())
	}
	return err
}

func main() {

	var wg sync.WaitGroup
	var exp string

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
			go find(filepath.Join(root, basedirs[dir].Name()), &wg, exp)
		}
	}

	wg.Wait()
}

