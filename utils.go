package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func new_tmp_dir() string {
	const datadir = "/tmp/dssh.tmp"
	if err := os.Mkdir(datadir, 0777); err != nil {
		if !os.IsExist(err) {
			logerr("failed to create datadir[%v]", datadir)
			panic(err)
		}
	}

	var tmpdir, err = ioutil.TempDir(datadir, "")
	if err != nil {
		logerr("failed to create tmpdir under[%v]", datadir)
		panic(err)
	}
	return tmpdir
}
func calldiff(filelist []string) {
	var cmd = exec.Command("vimdiff", filelist...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	var err = cmd.Run()
	if err != nil {
		panic(fmt.Sprintf("failed to run diff: %v", err))
	}
}
func green(str string) string {
	return fmt.Sprintf("\033[0;32m%s\033[0m", str)
}
func red(str string) string {
	return fmt.Sprintf("\033[0;31m%s\033[0m", str)
}
