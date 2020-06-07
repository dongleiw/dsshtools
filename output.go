package main

import (
	"fmt"
	"github.com/dongleiw/dssh"
	"strings"
	"time"
)

type Output struct {
	task       Task
	status     string
	output     string
	finish_seq int
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
	}
	if ret.IsSuccess() {
		output.output = ret.Stdout
		output.status = green(ret.GetStatus())
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
		output.status = red(ret.GetStatus())
	}

	self.map_task_output[task.UniqKey()] = output

	if !self.opt_group {
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
				status:     red("interrupt"),
				output:     "",
			}
			self.map_task_output[task.UniqKey()] = output
			if !self.opt_group {
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

	// 按照状态和输出分组, 确保不同情况不会混淆在一起
	var group = map[string][]*Output{}
	for _, output := range self.map_task_output {
		var key = output.status + output.output //
		group[key] = append(group[key], output)
	}

	if self.opt_group {
		for _, output_group := range group {
			self.print_group_title(output_group)
			print_newline(output_group[0].output)
		}
	}
	if self.opt_output_same {
		if len(group) != 1 {
			logerr("output-same-check fail")
			return false
		}
	}
	if len(self.opt_output_equal) > 0 {
		if len(group) != 1 {
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

func (self *DSSH) print_group_title(output_group []*Output) {
	var t = time.Now()
	var curtime = fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second())
	fmt.Printf("[%d/%d] [%s] [%s]\n", len(output_group), len(self.tasks), curtime, output_group[0].status)
	if !self.opt_quiet {
		for _, output := range output_group {
			fmt.Printf("  -> %s\n", output.task.addr)
		}
	}
}

// 打印一个指令的执行结果
func (self *DSSH) print_output(output *Output) {
	if !self.opt_quiet {
		var t = time.Now()
		var curtime = fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second())
		fmt.Printf("%d/%d %s %s %s\n", output.finish_seq, len(self.tasks), curtime, output.status, output.task.addr)
	}
	print_newline(output.output)
}
