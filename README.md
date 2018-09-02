# 项目介绍

luckybot 是使用 GoLang 语言开发的通用 [Telegram](https://telegram.org/) 加密货币红包机器人框架。封装了充值、提现、发红包、抢红包、操作纪录等诸多功能，开发者只需简单修改配置文件即可运行工作。脚本系统使用 Lua 语言编写，目的是帮助用户快速定制开发自己的红包机器人，无需重新修改、编译程序。开发者也不必关心红包机器人的实现细节，只需对接加密货币的充值提现即可实现自己的红包机器人了。

# 开发环境
* golang 1.8+
* python 2/3
* glide

# 快速开始
### 1. 拉取代码
```bash
git clone https://github.com/zhangpanyi/luckybot.git
```

### 2. 安装依赖
```bash
go get -u github.com/Masterminds/glide
glide install
```
> 如果无法翻墙请将 `.glide` 文件夹拷贝到 `%HOME%` 目录

### 3. 编译程序
```bash
go build
```

### 4. 初始配置
```bash
python init_config.py
```
Telegram 机器人必须开启 [Inline mode](https://core.telegram.org/bots/inline) ，再将 server.yml 配置文件中 **token** 字段的值填写为你 Telegram 机器人 Token。 

### 5. 运行服务

**Linux**

```bash
./luckybot
```

**Windows**

```bash
luckybot.exe
```

# 配置文件

# 充值接口

# 脚本系统

# 管理接口
