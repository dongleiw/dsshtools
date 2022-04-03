### 功能简介

类似pssh. 但是添加了一些常用的功能

- 默认指定pipefail, errexit, nounset, 可以屏蔽
- 可以按照结果分组显示
- 可以调用vimdiff工具查看多个输出之间的不同
- 支持指定用户执行
- 支持从stdin读取指令执行
- 支持从stdin读入机器地址	
- 打印将要执行的指令, 然后退出
- 处理SIGINT信号
- 指定执行一个本地的脚本
- 等待指令结束的连接超时时间,执行超时时间
- 指定并发数
- 按照指定的目标机器的顺序执行

### 编译

首先需要下载依赖
```
go get -u github.com/pborman/getopt/v2
go get -u golang.org/x/crypto/ssh # https://github.com/golang/crypto.git
```

然后编译. 可执行文件在build/下
```
make
```

### 示例
	
假设有3台机器: A B C

#### 获取系统版本号
```
$ ./dssh -H A,B,C cat /etc/redhat-release
1/3 14:52:0 success A
CentOS Linux release 7.2.1511 (Core) 
2/3 14:52:0 success B
CentOS release 6.5 (Final)
3/3 14:52:0 success C
CentOS release 6.5 (Final)
```

#### 分组 (相同输出的合并显示. 在一些场景下非常有用
```
# 按照系统版本号将机器分组
$ ./dssh -g -H A,B,C cat /etc/redhat-release
[2/3] [14:52:29] [success]
  -> C
  -> B
CentOS release 6.5 (Final)
[1/3] [14:52:29] [success]
  -> A
CentOS Linux release 7.2.1511 (Core) 


# 显示每种系统有多少台.
$ ./dssh -q -g -H A,B,C cat /etc/redhat-release
[2/3] [14:53:40] [success]
CentOS release 6.5 (Final)
[1/3] [14:53:40] [success]
CentOS Linux release 7.2.1511 (Core) 
```

#### 管道 (可以从管道中读取iplist, 在串联多个步骤时非常有用
```
# 管道传输机器列表
cat iplist | dssh df

# 管道传输指令
echo df | dssh -h iplist
```
	
#### 切换账号
```
# 登录到目标机器, sudo到root执行
dssh -H A -H B -u root 'id && pwd'
```

#### 执行本地脚本
```
dssh -H A,B -s './do_something.sh arg1 arg2'
```

#### 查看sysctl.conf的区别  (有时候不仅要看多少种输出结果, 还需要对比输出的差异. 可以通过指定'-d'. 
```
dssh -d -H A,B,C 'cat /etc/sysctl.conf'
```

### 文档
```
./dssh # 无任何参数则打印help信息
```

