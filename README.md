### mo visualizer 的作用

mo visualizer 提前在代码里埋点，统计相关的 trace 信息到 system.span_info 表中。待相关的负载完成后，再读表中的数据加以分析，
最后以 web、文本等方式可视化分析结果，以此排查系统中可能存在的性能问题。

目前支持的埋点有：

```golang
trace.SpanKindS3FSVis
trace.SpanKindLocalFSVis
```

### 如何使用 mo visualizer
基本步骤如下：
1. 向 mo 发送命令启用 trace
```MySQL
# 启动 CN 的trace，注意，不要有多余的空格，下同
select mo_ctl("cn", "TraceSpan", "all:enable:local,s3");

# 启动 DN 的trace
select mo_ctl("dn", "TraceSpan", "enable:local,s3");
```
2. 跑负载，如 TPCC 100 warehouse * 100 terminal
3. 负载结束后，执行 mo visualizer 脚本。visualizer 有两种使用方式
   1. 直接读 mo 的表，这种方式，需要指定 mo 的地址、用户名和密码。如果不指定，会使用默认值
      1. -h host: 默认值 127.0.0.1
      2. -P port: 默认值 6001
      3. -p pwd:  默认值 111
      4. -u user:  默认值 dump
   2. 提前将 mo 的数据 dump 到 CSV 文件中，visualizer 再分析该 CSV 文件。这种方式不能省略参数
      1. -f file

注意，
1. 上述两种方式的参数是互斥的，如果指定了 -f，就不能指定 -h,-P,-p,-u，反之也一样。
2. 参数与值之间必须有空格，如 -P 6001, 不能是 -P6001
3. 参数顺序可以随意

另外，如果需要使用 web 展示功能还可以指定端口号 -http port，如果不指定，那么 visualizer 只会生成文字报告，不会提供 http 服务

一些运行脚本的例子：
```
go run main.go -h 10.0.183.106 -p 111
go run main.go -P 6001 -u dump -p 111
# 会启动 http 服务
go run main.go -f src.csv -http 11235 -p 222 
```

4. 待 visualizer 执行完成后，再向 mo 发送命令关闭 trace
```mysql
select mo_ctl("cn", "TraceSpan", "all:disable:local,s3");
select mo_ctl("dn", "TraceSpan", "disable:local,s3");
```

### daily 流程相关
1. 目前只需要分析 TPCC 的 100 仓 * 100 线程
   1. 在其他负载结束后，跑 TPCC 100 * 100 之前，启用 trace（命令参见上面）
   2. 跑 TPCC 100 * 100 负载
   3. 负载结束后，运行该 visualizer 脚本。运行脚本之前需要先获取到任意一个 CN 的 IP 地址，使用 -h 传递给脚本
   4. 脚本运行结束后，相关数据会保存在 reports 和 src_data 目录，希望能从 本次 GitHub daily 地址处下载
   5. 由于 trace 可能会影响性能，所以最后需要向 mo 发送命令关闭 trace（命令见上面）




