# Mint

Mint (**M**inimal **in**i**t** system) 是一个单文件的极简进程管理工具。适用于容器等环境。

在系统的 /etc/ 目录下添加一个配置文件，配置文件名： Procfile，格式：

```
进程名：启动命令
```

配置示例：

```
web: go run web.go -a :$PORT
worker: bundle exec ruby worker.rb
```

支持如下命令：

```
$ mint master       -- 提供给容器的启动命令，作为容器的 1 号进程
$ mint reload       -- 重新读取配置文件，并使用新的配置启动进程
$ mint start XXX    -- 启动进程名为 XXX 的进程
$ mint stop  XXX    -- 停掉进程为 XXX 的进程
$ mint restart XXX  -- 重启进程为 XXX 的进程
$ mint status       -- 打印所有运行的进程运行状况，格式参考 supervisord
```

实现参考了以下项目：

- https://github.com/ddollar/forego
- https://github.com/mattn/goreman
- https://github.com/ddollar/foreman/wiki
- http://www.supervisord.org/

其他：

- 能替不是自己起的挂掉的进程收尸，防止容器中出现僵尸进程。
- 捕获管理的进程的 stdout/stderr 输出到 docker 日志中并打上进程名 tag。
- 本工具编译为纯静态程序，通过 docker volume 映射为容器中的 /bin/mint ，除了管理自己的进程外。
