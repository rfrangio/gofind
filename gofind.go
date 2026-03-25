package main

// Author: Robert B Frangioso

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type msg_type int

const (
	OUTPUT msg_type = 1 + iota
	ERROR
	CLOSE
)

type output_msg struct {
	mtype  msg_type
	buffer bytes.Buffer
}

func find(root string, wg *sync.WaitGroup, flags []string, args []string, output chan output_msg) {

	defer wg.Done()
	var cmd_out, cmd_err bytes.Buffer
	var msg output_msg

	findstr := append(append(append([]string{}, flags...), []string{root}...), args...)
	cmd := exec.Command(findname, findstr...)
	cmd.Stdout = &cmd_out
	cmd.Stderr = &cmd_err
	err := cmd.Run()

	if cmd_out.Len() > 0 {
		msg.mtype = OUTPUT
		msg.buffer = cmd_out
		output <- msg
	}

	if err != nil && cmd_err.Len() > 0 {
		msg.mtype = ERROR
		msg.buffer = cmd_err
		output <- msg
	}

	return
}

func aggregator(wg *sync.WaitGroup, input chan output_msg, status chan bool) {

	defer wg.Done()
	var msg output_msg
	hadError := false

	for true {
		msg = <-input
		switch msg.mtype {
		case CLOSE:
			status <- hadError
			return
		case OUTPUT:
			if msg.buffer.Len() > 0 {
				fmt.Printf("%s", msg.buffer.String())
			}
		case ERROR:
			if msg.buffer.Len() > 0 {
				fmt.Fprintf(os.Stderr, "%s", msg.buffer.String())
			}
			hadError = true
		default:
		}
	}

}

func main() {
	flag.Usage = gofind_usage
	var wg, wga sync.WaitGroup
	msg_channel := make(chan output_msg)
	status_channel := make(chan bool, 1)

	set_flags := parseflags()
	argslice := flag.Args()
	basedirs := []string{}
	rootdirs, options := parseargs(argslice)

	for r := range rootdirs {
		excludelist := []string{}
		dirs, direrr := os.ReadDir(rootdirs[r])
		if direrr != nil {
			gofind_usage()
			return
		}
		for dirindex := range dirs {
			if dirs[dirindex].IsDir() {
				subdir := joinSearchRoot(rootdirs[r], dirs[dirindex].Name())
				basedirs = append(basedirs, subdir)
				if dirindex == 0 {
					excludelist = append([]string{"!", "-name"}, dirs[dirindex].Name())
				} else {
					excludelist = append(append(excludelist, []string{"-and", "!", "-name"}...), dirs[dirindex].Name())
				}
			}
		}
		if !isplan9 {
			shallowfind := append(append([]string{"-maxdepth", "1"}, excludelist...), options...)
			wg.Add(1)
			go find(rootdirs[r], &wg, set_flags, shallowfind, msg_channel)
		}
	}

	for dir := range basedirs {
		wg.Add(1)
		go find(basedirs[dir], &wg, set_flags, options, msg_channel)
	}

	wga.Add(1)
	go aggregator(&wga, msg_channel, status_channel)
	wg.Wait()

	msg_channel <- output_msg{CLOSE, bytes.Buffer{}}
	wga.Wait()
	if <-status_channel {
		os.Exit(1)
	}
}

var findname string = "find"
var isplan9 bool = false

func joinSearchRoot(root string, child string) string {
	if filepath.IsAbs(root) {
		return filepath.Join(root, child)
	}

	sep := string(os.PathSeparator)
	switch {
	case root == ".":
		return "." + sep + child
	case strings.HasPrefix(root, "."+sep):
		return root + sep + child
	default:
		return filepath.Join(root, child)
	}
}

func parseflags() []string {

	os_find_flags := getosfindflags()
	set_flags := []string{}

	for f := range os_find_flags {
		flag.Bool(os_find_flags[f], false, "bool")
	}

	flag.Parse()
	for f := range os_find_flags {
		flag_p := flag.Lookup(os_find_flags[f])
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
		if strings.HasPrefix(args[i], "-") {
			break
		}
		i++
	}

	rootdirs := append([]string{}, args[:i]...)
	options := append([]string{}, args[i:]...)
	return rootdirs, options
}

func gofind_usage() {
	fmt.Fprintf(os.Stderr, "Usage: gofind [find-flags] rootsearchdir[...] [find-options]\n(O & D find-flags for gnu find not supported atm)\n")
}

func getosfindflags() []string {
	var flags []string

	switch runtime.GOOS {
	case "darwin", "freebsd", "dragonfly":
		// osx find derived from freebsd and using same flags
		flags = []string{"L", "H", "P", "E", "X", "d", "s", "x", "f"}
	case "linux":
		// gnu find
		flags = []string{"L", "H", "P", "D", "O"}
	case "windows":
		// assuming gnu find for windows
		flags = []string{"L", "H", "P", "D", "O"}
	case "netbsd":
		flags = []string{"L", "H", "P", "E", "X", "d", "s", "x", "f", "h"}
	case "openbsd":
		flags = []string{"d", "H", "h", "L", "X", "x"}
	case "plan9":
		flags = []string{"a", "e", "f", "h", "n", "q", "s", "t", "u", "b", "p"}
		isplan9 = true
		findname = "du"
	default:
		// assume freebsd variant
		flags = []string{"L", "H", "P", "E", "X", "d", "s", "x", "f"}
	}

	return flags
}
