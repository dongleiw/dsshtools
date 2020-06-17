/*
	dssh
	提供在多个目标机器批量执行指令的功能
	可以指定账户执行(sudo方式)
	可以按照输出结果分组
	支持从stdin输入指令
*/
package main

import (
	"github.com/dongleiw/dssh"
	"os"
)

type DSSH struct {
	// opts
	opt_quiet        bool
	opt_verbose      bool
	opt_group        bool
	opt_diff         bool
	opt_timeout      uint32
	opt_conntimeout  uint32
	opt_parallel     int
	opt_output_same  bool
	opt_output_equal string
	opt_noopt        bool

	tasks []Task

	map_task_output map[string]*Output // task.key -> Output
	output_groups   [][]*Output
}

func main() {
	var err = dssh.Initialize(true)
	if err != nil {
		panic(err)
	}

	var arg_parser = new_arg_parser()
	var dssh = arg_parser.ParseArgs(os.Args)
	if !dssh.Execute() {
		os.Exit(1)
	}
}
