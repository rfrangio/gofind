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

type msg_type int

const (
	OUTPUT msg_type = 1 + iota
	ERROR
	CLOSE
) 

type output_msg struct {
	mtype	msg_type
	buffer	bytes.Buffer
}

func find(root string, wg *sync.WaitGroup, exp []string, output chan output_msg) {

	defer wg.Done()
	var cmd_out, cmd_err bytes.Buffer
	var msg output_msg 

	findargs := append([]string{root}, exp... )
	cmd := exec.Command("find", findargs... )
	cmd.Stdout = &cmd_out
	cmd.Stderr = &cmd_err
	err := cmd.Run()

	if err == nil {
		msg.mtype = OUTPUT
		msg.buffer = cmd_out
		output <- msg
	} else {
		msg.mtype = ERROR
		msg.buffer = cmd_err
		output <- msg
	}

	return
}

func aggregate(wg *sync.WaitGroup, input chan output_msg) {

	defer wg.Done()
	var msg output_msg

	for true {
		msg = <- input
		switch msg.mtype {
		case CLOSE:
			return
		case OUTPUT:
			if msg.buffer.Len() > 0 {
				fmt.Printf("%s", msg.buffer.String())
			}	
		case ERROR:
			if msg.buffer.Len() > 0 {
				fmt.Fprintf(os.Stderr, "%s", msg.buffer.String())
			}
		default:
		}
	}

}

func main() {

	var wg, wga sync.WaitGroup
	msg_channel := make(chan output_msg)

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
	go find(root, &wg, shallowfind, msg_channel) 

	for  dir := range basedirs {
		if basedirs[dir].IsDir() {
			wg.Add(1)
			go find(filepath.Join(root, basedirs[dir].Name()), &wg, argslice[1:], msg_channel)
		}
	}

	wga.Add(1)
	go aggregate(&wga, msg_channel)
	wg.Wait()

	msg_channel <- output_msg{CLOSE, bytes.Buffer{}}
	wga.Wait()
}

