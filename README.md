# osmonitor
不用部署复杂运维工具。通过该程序，可尽可能简单的实时监控分布自不同区域的服务器情况  
尊崇个人执念`大道至简`：  
本身是为了解决这个问题：有大量的设备（几十乃至数百台），每台设备部署了不同的程序在运行，进程运行期间，某些进程会掉线，此时逐个排查很不现实，想及时知道这些设备是否有异常的，也很麻烦。而为了这么些需求，去部署一套复杂的运维工具，也不现实。
为此，本工具诞生，作为非专业的运维人员，也能很快上手。

![start](/logo.png)

## 功能
1. 每个服务器节点名称不得重复
2. 监控各服务器的指定进程是否掉线，同时监控服务器是否掉线，并定时汇发送该信息到邮箱
3. server和client强稳定性，可持续稳定运行，降低了运维复杂度，差不多就是个守护进程，要是总停止，三天两头去重启服务，很烦人的。
4. 可通过命令行传入参或者通过配置文件启动`client、server`，不建议同时使用两种方式，选择其中一种即可
5. 本项目没有前端页面，主要是不会用前端语言，也设计不了。。。本项目在server/app/service/api中提供了socket数据出口，只要前端使用socket调用，即可渲染在前端。
6. 其余功能会根据个人需要，陆续开发，比如定时发送各设备内存、硬盘空间、温度等情况（或者是达到阈值则发送邮件提醒） 

## 使用方式
分为客户端和服务器端，客户端安装在每台需要监控的节点上，服务器端找台有ip的稳定机子部署就行。
需要先在根目录执行：
```shell
go mod tidy
```

### client 客户端
```shell
cd client
go build -o client ./client.go 
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o client client.go

# 配置方式启动
client start -c setting.yml

# 命令行方式启动
## --name 节点名称
## --secret 和服务器端交互使用的私钥
## --server-url 服务端地址
./client start --name test-client --secret 123456 --server-url ws://127.0.0.1:8003
```

### server 服务端
```shell
cd server
go build server.go -o server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o server server.go

# 配置方式启动
server start -c setting.yml

# 命令行方式启动
## --name 服务端名称
## --secret 和客户端交互使用的私钥
## --host 服务端地址
## --port 端口
## --email-subject-prefix 邮件前缀
## --email-host 邮箱服务地址
## --email-port 邮箱端口
## --email-username 发件邮箱账户 
## --email-password 邮箱密钥 
## --email-from 发件邮箱账户 
## --email-to 收件邮箱账户(多个逗号隔开) 
## --email-monitor-time 邮件发送间隔，单位秒
./server start --name test-server --secret 123456 --host 0.0.0.0 --port 3000 --email-subject-prefix test --email-host 邮箱服务地址 --email-port 465 --email-username 发件邮箱账户 --email-password 邮箱密钥 --email-from 发件邮箱账户 --email-to 收件邮箱账户(多个逗号隔开) --email-monitor-time 7200
```
