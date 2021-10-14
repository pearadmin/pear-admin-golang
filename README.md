<div align="center">
<br/>
<br/>
  <h1 align="center">
    Pear Admin Go
  </h1>
</div>

#### 项目简介
>Pear Admin Go 基于 Gin框架  的后台管理系统
> 
>众人拾柴火焰高，欢迎参与项目~
>
>	go1.16	+	gin	+	mysql	+	权限验证	

>  注： demo分支只为演示使用


####  项目结构

```
Pear Admin Golang
├─app  # 应用
├─database  # 数据库预设文件
├─static  # 前端css、js、img文件
├─template # 前端html文件
├─go.mod # go mod文件
├─config.toml.examole # 配置文件样例
└─main.go # 项目主入口执行文件

```



#### 项目安装

```bash
# 下 载
git clone https://gitee.com/pear-admin/pear-admin-golang

# 修 改 配 置
cp config.toml.example config.toml


```

#### 运行项目

```bash
go mod tidy
go run main.go
```
#### 备注: 因数据库使用sqlite3，使用及编译时需要gcc。如果不希望额外配置gcc，可以使用发行版。 

#### 未完成工作
- [x] 修改路由结构
- [x] 修改配置方式
- [x] 去除多余函数
- [x] 优化文件层级
- [x] 本地、服务器文件备份迁移功能
- [ ] 完善操作日志 

![登陆日志](doc/image/login-log.png)
![操作日志](doc/image/oper-log.png)
![添加服务器-密码登陆](doc/image/server1.png)
![添加服务器-密钥登陆](doc/image/server2.png)
![添加任务](doc/image/task1.png)
![添加任务](doc/image/task2.png)
![添加任务](doc/image/task3.png)
![执行任务](doc/image/task4.png)