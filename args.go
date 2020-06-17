package main

import (
	"bufio"
	"fmt"
	"github.com/pborman/getopt/v2"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type HostArg struct {
	isfile bool
	value  string
}

type ArgParser struct {

	// opts
	opt              *getopt.Set
	opt_quiet        *bool
	opt_verbose      *bool
	opt_group        *bool
	opt_sudo         *string
	opt_uniq         *bool
	opt_test         *bool
	opt_diff         *bool
	opt_script_file  *string
	opt_timeout      *uint32
	opt_conntimeout  *uint32
	opt_parallel     *int
	opt_noopt        *bool // 是否去掉默认添加的pipefail, errexit, nounset
	opt_output_same  *bool
	opt_output_equal *string

	// parameters
	host_args []HostArg // 指定的所有目标机器信息. 有序
}

func new_arg_parser() ArgParser {
	var opt = getopt.New()
	opt.SetParameters("指令")
	opt.String('h', "", "包含目标地址的文件,每行一个. 多个文件用`,`隔开. 可以和-H同时使用. 可以指定多个. 如果没有通过`-h`或者`-H`指定机器, 则从stdin读取", "[file1,file2,...]")
	opt.String('H', "", "目标地址. 多个地址用`,`隔开. 可以和-h同时使用. 可以指定多个. 如果没有通过`-h`或者`-H`指定机器, 则从stdin读取", "[host1,host2,..]")
	return ArgParser{
		opt:              opt,
		opt_quiet:        opt.BoolLong("quiet", 'q', "keep quiet"),
		opt_verbose:      opt.BoolLong("verbose", 'v', "显示更多日志"),
		opt_group:        opt.BoolLong("group", 'g', "相同执行状态和输出结果的机器合为一组, 使用-q屏蔽具体机器列表"),
		opt_sudo:         opt.StringLong("user", 'u', "", "在目标机器以${user}运行指令", "[user]"),
		opt_uniq:         opt.BoolLong("uniq", 'U', "检查任务(addr+sudo+cmd)不能重复"),
		opt_test:         opt.BoolLong("test", 0, "", "将每台机器需要执行的指令显示(但不执行), 然后退出"),
		opt_diff:         opt.Bool('d', "分组输出到文件. 调用diff工具查看差异. 默认为vimdiff. -q屏蔽具体机器列表"),
		opt_script_file:  opt.String('s', "", "指定本地shell脚本. 脚本参数空格隔开", "[scriptfile]"),
		opt_timeout:      opt.Uint32Long("timeout", 't', 0, "等待指令完成的超时时间(包括连接建立的时间). 默认0表示不超时", "[uint32]"),
		opt_conntimeout:  opt.Uint32Long("conntimeout", 0, 3, "连接建立的超时时间. 默认3", "[uint32]"),
		opt_parallel:     opt.Int('p', 100, "指定并发数. 默认. 如果指定为1, 则为串行有序执行", ""),
		opt_noopt:        opt.Bool('n', "屏蔽默认添加的pipefail, errexit, nounset"),
		opt_output_same:  opt.BoolLong("output-same", 0, "所有输出应该相同(且成功)"),
		opt_output_equal: opt.StringLong("output-equal", 0, "", "所有输出应该等于value(且成功). 忽略前后空白字符", "[value]"),

		host_args: make([]HostArg, 0),
	}
}

func (self *ArgParser) print_help() {
	getopt.DisplayWidth = 140
	getopt.HelpColumn = 30

	fmt.Fprintf(os.Stderr, "批量ssh工具\n")
	fmt.Fprintf(os.Stderr, "exitcode:\n")
	fmt.Fprintf(os.Stderr, "    如果所有指令都执行成功(返回0), 则返回0\n")
	fmt.Fprintf(os.Stderr, "    如果任何指令执行失败, 或者连接失败, 或者参数错误, 或者超时, 或者被打断, 都返回非0\n")

	self.opt.PrintUsage(os.Stderr)
	fmt.Fprintf(os.Stderr, "\n")

	fmt.Fprintf(os.Stderr, `
例子:
    dssh -h iplist-H 127.0.0.1,localhost -H 3.4.5.6 df   # 指定多个地址执行
    cat iplist | dssh df                                 # stdin读取iplist执行
    echo df | dssh -h iplist                             # stdin读取指令执行
    dssh -h iplist -s 'remote.sh 1 2 3'                  # 执行本地脚本
`)
}

// 从指定的本地脚本中读取指令
// 空格分割后, 第一个为脚本 后面都是脚本参数
// 定义参数:
//		$1 = 参数1
//		$2 = 参数2
//		...
func read_cmd_from_scriptfile(script string) string {
	var cmd string

	var splits = strings.Fields(script)
	if len(splits) == 0 {
		fmt.Fprintf(os.Stderr, "unknown script format[%s]\n", script)
		os.Exit(1)
	}
	var scriptfile = splits[0]
	var args = splits[1:]

	// 添加cmd
	cmd = fmt.Sprintf("set -- %v\n", strings.Join(args, " "))

	// 脚本内容
	var content, err = ioutil.ReadFile(scriptfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read scriptfile[%s]: %+v\n", scriptfile, err)
		os.Exit(1)
	}
	cmd += string(content)

	return cmd
}

// 从stdin读取指令
func read_cmd_from_stdin() string {
	var cmd string

	var reader = bufio.NewReader(os.Stdin)
	for {
		var line, err = reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				cmd += line
				break
			} else {
				fmt.Fprintf(os.Stderr, "failed to read from stdin: %s\n", err)
				os.Exit(1)
			}
		}
		cmd += line
	}
	return cmd
}

// 从stdin读入目标机器地址
// 按照读入顺序返回
func read_address_from_stdin() []string {
	var hosts = []string{}

	var reader = bufio.NewReader(os.Stdin)
	for {
		var line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "failed to read from stdin: %+v\n", err)
			os.Exit(1)
		}
		var host = strings.TrimSpace(line)
		if len(host) > 0 {
			hosts = append(hosts, host)
		}
		if err == io.EOF {
			break
		}
	}
	return hosts
}

// 从输入参数收集地址
// 按照顺序返回
func collect_address_from_host_args(host_args []HostArg) []string {
	var hosts = []string{}

	for _, hg := range host_args {
		if hg.isfile {
			// 从-h指定的文件读取
			for _, filepath := range strings.Split(hg.value, ",") {
				var sfilepath = strings.TrimSpace(filepath)
				if len(sfilepath) > 0 {
					var content, err = ioutil.ReadFile(sfilepath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s\n", err)
						os.Exit(1)
					}
					for _, line := range strings.Split(string(content), "\n") {
						var host = strings.TrimSpace(line)
						if len(host) == 0 || strings.HasPrefix(host, "#") {
							continue
						}
						hosts = append(hosts, host)
					}
				}
			}
		} else {
			// -H指定的地址
			for _, host := range strings.Split(hg.value, ",") {
				var shost = strings.TrimSpace(host)
				if len(shost) > 0 {
					hosts = append(hosts, shost)
				}
			}
		}
	}

	return hosts
}

// getopt 解析回调. 用来收集指定的所有目标机器信息(从左到右)
func (self *ArgParser) opt_parse_callback(o getopt.Option) bool {
	if o.Name() == "-h" {
		self.host_args = append(self.host_args, HostArg{isfile: true, value: o.String()})
	} else if o.Name() == "-H" {
		self.host_args = append(self.host_args, HostArg{isfile: false, value: o.String()})
	}
	return true
}

// 解析参数, 获取
func (self *ArgParser) ParseArgs(args []string) *DSSH {
	if len(args) <= 1 {
		self.print_help()
		os.Exit(1)
	}

	if err := self.opt.Getopt(args, self.opt_parse_callback); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if *self.opt_verbose {
		enable_debug()
	}
	logdebug("%v", self.host_args)

	if *self.opt_verbose && *self.opt_quiet {
		*self.opt_verbose = false
	}

	if *self.opt_parallel <= 0 {
		fmt.Fprintf(os.Stderr, "parallel <= 0\n")
		os.Exit(1)
	}

	var tasks = []Task{}
	// 如果没有指定目标地址, 也没有指定指令, 报错 (没法从stdin同时读取两个内容)
	if len(self.host_args) <= 0 && len(self.opt.Args()) <= 0 && len(*self.opt_script_file) <= 0 {
		fmt.Fprintf(os.Stderr, "both cmd and address are empty!\n")
		os.Exit(1)
	}
	// 不能同时指定指令和脚本
	if len(self.opt.Args()) > 0 && len(*self.opt_script_file) > 0 {
		fmt.Fprintf(os.Stderr, "both scriptfile and cmdargs are not empty\n")
		os.Exit(1)
	}

	// 读取host
	var hosts = []string{}
	if len(self.host_args) > 0 {
		hosts = collect_address_from_host_args(self.host_args)
	} else {
		hosts = read_address_from_stdin()
	}

	// 读取cmd
	var cmd string
	if len(self.opt.Args()) > 0 {
		cmd = strings.Join(self.opt.Args(), " ")
	} else if len(*self.opt_script_file) > 0 {
		cmd = read_cmd_from_scriptfile(*self.opt_script_file)
	} else {
		cmd = read_cmd_from_stdin()
	}
	tasks = create_tasks_from_args(hosts, *self.opt_sudo, cmd)

	if *self.opt_uniq {
		uniq_check(tasks)
	}
	if *self.opt_test {
		for _, task := range tasks {
			fmt.Printf("[%v/%v] %v %v\n", task.id, len(tasks), task.sudo, task.addr)
			fmt.Printf("%v\n", task.cmd)
		}
		os.Exit(0)
	}
	return &DSSH{
		opt_quiet:        *self.opt_quiet,
		opt_verbose:      *self.opt_verbose,
		opt_group:        *self.opt_group,
		opt_diff:         *self.opt_diff,
		opt_timeout:      *self.opt_timeout,
		opt_conntimeout:  *self.opt_conntimeout,
		opt_parallel:     *self.opt_parallel,
		opt_output_same:  *self.opt_output_same,
		opt_output_equal: *self.opt_output_equal,
		opt_noopt:        *self.opt_noopt,
		tasks:            tasks,
		map_task_output:  map[string]*Output{},
	}
}
