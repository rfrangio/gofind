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

	findargs := make([]string, 1)
	findargs[0] = root 
	findargs = append(findargs, exp... )
	cmd := exec.Command("find", findargs... )
	cmd.Stdout = &cmd_out
	err := cmd.Run()

	if(cmd_out.Len() > 0) {
		fmt.Printf("%s\n", cmd_out.String())
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
	}

	for  dir := range basedirs {
		if basedirs[dir].IsDir() {
			fmt.Printf("dir: %v \n", basedirs[dir])
			wg.Add(1)
			go find(filepath.Join(root, basedirs[dir].Name()), &wg, argslice[1:])
		}
	}

	wg.Wait()
}

