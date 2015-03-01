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
	"strconv"
	"strings"
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

func find(root string, wg *sync.WaitGroup, flags []string, args []string, output chan output_msg) {

	defer wg.Done()
	var cmd_out, cmd_err bytes.Buffer
	var msg output_msg 

	findstr := append(append(append([]string{}, flags... ), []string{root}... ), args... )
	cmd := exec.Command("find", findstr... )
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

func aggregator(wg *sync.WaitGroup, input chan output_msg) {

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

func parseflags() []string {

	osx_find_flags := []string{"L", "H", "P", "E", "X", "d", "s", "x", "f"}
	set_flags := []string{}

	for f := range osx_find_flags {
		flag.Bool(osx_find_flags[f], false, "bool")
	}

	flag.Parse()
	for f := range osx_find_flags {
		flag_p := flag.Lookup(osx_find_flags[f])
		val, err := strconv.ParseBool(flag_p.Value.String())
		if err == nil && val == true {
			set_flags = append(set_flags, "-"+flag_p.Name)
		}
	}

	return set_flags
}

func parseargs(args []string) ([]string, []string) {

	var i int

	for i = range args {
		if strings.HasPrefix(args[i], "-")  { break }
		i++
	}

	rootdirs := append([]string{}, args[:i]... )
	options := append([]string{}, args[i:]... )
	return rootdirs, options
}

func gofind_usage() {
        fmt.Fprintf(os.Stderr, "Usage: gofind [find-flags] rootsearchdir[...] [find-options]\n(osx only find-flags atm)\n")
}

func main() {

	flag.Usage = gofind_usage
	var wg, wga sync.WaitGroup
	msg_channel := make(chan output_msg)

	set_flags := parseflags()
	argslice := flag.Args()
	basedirs := []string{}
	rootdirs, options := parseargs(argslice)

	for r := range rootdirs {
		dirs, direrr := ioutil.ReadDir(rootdirs[r])
		if(direrr != nil) {
			gofind_usage()
			//	fmt.Printf("Usage: gofind [find-flags] rootsearchdir[...] [find-options] \n")
			return
		}
		for dirindex := range dirs {
			if dirs[dirindex].IsDir() {
				basedirs = append(basedirs, filepath.Join(rootdirs[r], dirs[dirindex].Name()))
			}
		}
		shallowfind := append(append([]string{},[]string{"-maxdepth", "1"}... ), options... )
		wg.Add(1)
		go find(rootdirs[r], &wg, set_flags, shallowfind, msg_channel) 
	}

	for  dir := range basedirs {
		wg.Add(1)
		go find(basedirs[dir], &wg, set_flags, options, msg_channel)
	}

	wga.Add(1)
	go aggregator(&wga, msg_channel)
	wg.Wait()

	msg_channel <- output_msg{CLOSE, bytes.Buffer{}}
	wga.Wait()
}

