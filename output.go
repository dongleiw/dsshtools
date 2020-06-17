package main

import (
	"fmt"
	"github.com/dongleiw/dssh"
	"io/ioutil"
	"strings"
	"time"
)

type Output struct {
	task       Task
	status     string
	succ       bool
	output     string
	finish_seq int
}

func (self Output) GetStatus(color bool) string {
	if color {
		if self.succ {
			return green(self.status)
		} else {
			return red(self.status)
		}
	} else {
		if self.succ {
			return self.status
		} else {
			return self.status
		}
	}
}

// 连接失败 或者 有一个指令出结果了
// 返回指令是否成功执行: 连接成功, 指令有exitcode, 而且exitcode=0.
func (self *DSSH) CmdFinished(task Task, ret *dssh.CmdResult, finish_seq int) bool {
	if task.addr != ret.Addr || task.sudo != ret.Sudo {
		panic(fmt.Sprintf("bug: addr [%v][%v] sudo[%v][%v]", task.addr, ret.Addr, task.sudo, ret.Sudo))
	}
	logdebug("cmd finished: addr[%s@%s] cost[%v]ns output[%v]", task.addr, task.sudo, ret.End_time.UnixNano()-ret.Beg_time.UnixNano(), ret.GetStdout())

	var output = &Output{
		task:       task,
		finish_seq: finish_seq,
		status:     ret.GetStatus(),
		succ:       ret.IsSuccess(),
	}
	if ret.IsSuccess() {
		output.output = ret.Stdout
	} else {
		if len(ret.Stdout) > 0 {
			output.output += ret.Stdout
		}
		if len(ret.Stderr) > 0 {
			output.output += ret.Stderr
		}
		if ret.Err != nil {
			output.output += fmt.Sprintf("%v\n", ret.Err)
		}
	}

	self.map_task_output[task.UniqKey()] = output

	if !self.opt_group && !self.opt_diff {
		self.print_output(output)
	}
	return ret.IsSuccess()
}

// 被打断了(比如ctrl-c)
func (self *DSSH) Interrupt(finish_seq int) {
	for _, task := range self.tasks {
		if nil == self.map_task_output[task.UniqKey()] {
			finish_seq++
			logdebug("interrupt. addr[%s@%s]", task.sudo, task.addr)
			var output = &Output{
				task:       task,
				finish_seq: finish_seq,
				status:     "interrupt",
				succ:       false,
				output:     "",
			}
			self.map_task_output[task.UniqKey()] = output
			if !self.opt_group && !self.opt_diff {
				self.print_output(output)
			}
		}
	}
}

// 都结束了. 处理输出
func (self *DSSH) Finished() bool {
	if len(self.map_task_output) != len(self.tasks) {
		panic("bug")
	}

	var m = map[string][]*Output{}
	for _, output := range self.map_task_output {
		var key = output.status + output.output //
		m[key] = append(m[key], output)
	}
	for _, group := range m {
		self.output_groups = append(self.output_groups, group)
	}

	if self.opt_diff {
		self.diff()
	}
	if self.opt_group {
		self.group()
	}
	if self.opt_output_same {
		if len(self.output_groups) != 1 {
			logerr("output-same-check fail")
			return false
		}
	}
	if len(self.opt_output_equal) > 0 {
		if len(self.output_groups) != 1 {
			logerr("output-equal-check fail")
			return false
		}
		for _, output := range self.map_task_output {
			if strings.TrimSpace(output.output) != strings.TrimSpace(self.opt_output_equal) {
				logerr("output-equal-check fail: [%v]", output.output)
				return false
			}
			break
		}
	}
	return true
}

// 如果为空. 不输出
// 否则如果没有空行, 就自动补一个空行
func print_newline(s string) {
	if len(s) > 0 {
		if strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\n\r") {
			fmt.Print(s)
		} else {
			fmt.Println(s)
		}
	}
}

/*
	拼接出一组状态和输出相同的机器信息
*/
func (self *DSSH) get_group_title(output_group []*Output, color bool) string {
	var curtime = time.Now().Format("15:04:05")
	var title = fmt.Sprintf("%d/%d %s %s\n", len(output_group), len(self.tasks), curtime, output_group[0].GetStatus(color))
	if !self.opt_quiet {
		for _, output := range output_group {
			title += fmt.Sprintf("  -> %s\n", output.task.addr)
		}
	}
	return title
}

// 打印一个指令的执行结果
func (self *DSSH) print_output(output *Output) {
	if !self.opt_quiet {
		var t = time.Now()
		var curtime = fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second())
		fmt.Printf("%d/%d %s %s %s\n", output.finish_seq, len(self.tasks), curtime, output.GetStatus(true), output.task.addr)
	}
	print_newline(output.output)
}

func (self *DSSH) diff() {
	// 存储
	var tmpdir = new_tmp_dir()
	var idx = 1
	var filelist []string
	for _, output_group := range self.output_groups {
		var filename = fmt.Sprintf("%v/%v", tmpdir, idx)
		var title = self.get_group_title(output_group, false)
		var content = fmt.Sprintf("%v\n-------------------------------------\n%v", title, output_group[0].output)
		var err = ioutil.WriteFile(filename, []byte(content), 0664)
		if err != nil {
			logerr("failed to write to file[%v]", filename)
			panic(err)
		}
		filelist = append(filelist, filename)
		idx++
	}
	fmt.Println(strings.Join(filelist, " "))

	// diff
	calldiff(filelist)
}

func (self *DSSH) group() {
	for _, output_group := range self.output_groups {
		fmt.Printf(self.get_group_title(output_group, true))
		print_newline(output_group[0].output)
	}
}
