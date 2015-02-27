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

func find(root string, wg *sync.WaitGroup, exp []string) error {

	defer wg.Done()
	var cmd_out bytes.Buffer

	findargs := append([]string{root}, exp... )
	cmd := exec.Command("find", findargs... )
	cmd.Stdout = &cmd_out
	err := cmd.Run()

	if(cmd_out.Len() > 0) {
		fmt.Printf("%s", cmd_out.String())
	}
	return err
}

func main() {

	var wg sync.WaitGroup

	flag.Parse()
	root := flag.Arg(0)
	basedirs, direrr := ioutil.ReadDir(root)

	argslice := flag.Args() 
 
	if(direrr != nil) {
		fmt.Printf("ReadDir err %v \n", direrr)
		fmt.Printf("Usage: gofind rootsearchdir <other-find-args> \n")
		return
	}

	shallowfind := append(append([]string{},[]string{"-maxdepth", "1"}... ), argslice[1:]... )
	wg.Add(1)
	go find(root, &wg, shallowfind) 

	for  dir := range basedirs {
		if basedirs[dir].IsDir() {
			wg.Add(1)
			go find(filepath.Join(root, basedirs[dir].Name()), &wg, argslice[1:])
		}
	}

	wg.Wait()
}

