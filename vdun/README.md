# 微盾

HTTP Flood 防火墙

## 安装运行

```sh
$ mvn clean install
$ storm jar target/vdun-0.0.1-SNAPSHOT.jar vdun.VDunTopology vdun.yml
```

目前运行在 storm local cluster 上，仅供验证功能用，可修改运行在集群模式。

## Topylogy 说明

input（日志输入）-> filter（提取有效日志特征）
   -> detect（计算请求特征） -> brain（风险检测）-> output（回调拦截）

## 输入日志格式

日志通过 kafka 输入。

配置日志格式如下，日志中需包含以下字段（一条日志一行，下面为了方便阅读做了 prettify）：

```json
{
   "request_length" : "$request_length",
   "body_bytes_sent" : "$body_bytes_sent",
   "http_referer" : "$http_referer",
   "http_user_agent" : "$http_user_agent",
   "request_uri" : "$request_uri",
   "hostname" : "$hostname",
   "upstream_response_time" : "$upstream_response_time",
   "upstream_addr" : "$upstream_addr",
   "request_time" : "$request_time",
   "remote_addr" : "$remote_addr",
   "remote_user" : "$remote_user",
   "scheme" : "$scheme",
   "http_host" : "$host",
   "http_x_forwarded_for" : "$http_x_forwarded_for",
   "status" : "$status",
   "method" : "$request_method",
   "time_local" : "$time_iso8601",
   "upstream_status" : "$upstream_status"
}
```

然后创建一个 kafka topic，通过 kafkacat 等客户端将日志实时写入 kafka 中。

```sh
$ kafka-topics.sh --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic web-log
$ tail -F –q /path/to/web.log | kafkacat -b localhost:6667 -t web-log -z snappy
```

微盾程序配置方法见 [vdun.yml](vdun.yml) 。
