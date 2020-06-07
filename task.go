package main

import (
	"fmt"
	"os"
)

/*
	一个任务
*/
type Task struct {
	// 任务信息
	id   int // 可能存在addr+sudo+cmd重复的情况, 添加一个唯一id
	addr string
	sudo string
	cmd  string
}

func (self *Task) String() string {
	return fmt.Sprintf("%s@%s\n\t######\n\t%s\n\t######\n", self.sudo, self.addr, self.cmd)
}
func (self *Task) UniqKey() string {
	return fmt.Sprintf("id[%v] sudo[%v] addr[%v] cmd[%v]", self.id, self.sudo, self.addr, self.cmd)
}
func (self *Task) Key() string {
	return fmt.Sprintf("sudo[%v] addr[%v] cmd[%v]", self.sudo, self.addr, self.cmd)
}

func create_tasks_from_args(addrs []string, sudo string, cmd string) []Task {
	var tasks = []Task{}
	for idx, addr := range addrs {
		tasks = append(tasks, Task{addr: addr, sudo: sudo, cmd: cmd, id: idx + 1})
	}
	return tasks
}

func uniq_check(tasks []Task) {
	var uniq = make(map[string]bool, 0)

	for _, task := range tasks {
		var k = task.Key()
		if _, e := uniq[k]; e {
			logerr("dup task: %v\n", task.Key())
			os.Exit(1)
		}
		uniq[k] = true
	}
}
