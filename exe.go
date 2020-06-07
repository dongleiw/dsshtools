package main

import (
	"fmt"
	"github.com/dongleiw/dssh"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ResultPack struct {
	ret *dssh.CmdResult
	idx int
}

func (self *DSSH) run_nb(idx int, task Task, ch chan<- ResultPack) {
	var opt = dssh.NewOption()
	opt.SetConnTimeout(time.Duration(self.opt_conntimeout) * time.Second)
	opt.SetExecTimeout(time.Duration(self.opt_timeout) * time.Second)
	opt.SetBashOpt(self.opt_noopt == false)
	var ret = dssh.SudoRun(task.addr, task.sudo, task.cmd, opt)
	ch <- ResultPack{
		ret: ret,
		idx: idx,
	}
}

// 执行指令
func (self *DSSH) Execute() bool {

	// sigint信号
	var signalchan = make(chan os.Signal, 1)
	signal.Notify(signalchan, syscall.SIGINT)

	var wait_run_idx = 0 // 等待执行的开始位置
	var running_num = 0  // 执行中的数量
	var finish_num = 0   // 完成的数量
	var finishchan = make(chan ResultPack)
	var has_err = false

	logdebug("begin to execute cmd")
	// 最大并发<=指定最大并发数
	for wait_run_idx < len(self.tasks) || running_num > 0 {
		for wait_run_idx < len(self.tasks) && running_num < self.opt_parallel {
			go self.run_nb(wait_run_idx, self.tasks[wait_run_idx], finishchan)
			wait_run_idx++
			running_num++
			logdebug("parallel=%d wait_run_idx=%d, running_num=%d\n", self.opt_parallel, wait_run_idx, running_num)
		}
		select {
		case rp := <-finishchan: // 有指令执行完了
			running_num--
			finish_num++
			var task = self.tasks[rp.idx]
			if !self.CmdFinished(task, rp.ret, finish_num) {
				has_err = true
			}
		case <-signalchan: // interrupt signal
			self.Interrupt(finish_num)
			finish_num = len(self.tasks)
			self.Finished()
			return false
		}
	}

	if wait_run_idx != len(self.tasks) || running_num != 0 || finish_num != len(self.tasks) {
		panic(fmt.Sprintf("bug: wait_run_idx[%v] running_num[%v] finish_num[%v] tasks[%v]", wait_run_idx, running_num, finish_num, len(self.tasks)))
	}

	if !self.Finished() {
		has_err = true
	}
	return has_err
}
