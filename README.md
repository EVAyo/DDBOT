# DDBOT

[<img src="https://github.com/Sora233/DDBOT/actions/workflows/ci.yml/badge.svg"/>](https://github.com/Sora233/DDBOT/actions/workflows/ci.yml)
[<img src="https://coveralls.io/repos/github/Sora233/DDBOT/badge.svg?branch=master"/>](https://coveralls.io/github/Sora233/DDBOT)

DDBOT是一个基于 [MiraiGO](https://github.com/Mrs4s/MiraiGo) 的QQ群推送机器人，支持b站直播/动态，斗鱼直播，YTB直播/预约直播，虎牙直播。

*DDBOT不是一个聊天机器人。*

[Bilibili专栏](https://www.bilibili.com/read/cv10602230)

-----

## 设计理念

制作bot的本意是为了减轻一些重复的工作负担，bot只会做好bot份内的工作：

- ddbot的交互被刻意设计成最小程度，正常交流时永远不必担心会误触ddbot。
- ddbot只有两种情况会主动发言，更新动态和直播，以及答复命令结果。

## **基本功能：**

- **B站直播/动态推送**
    - 让阁下在DD的时候不错过任何一场突击。
- **斗鱼直播推送**
    - 没什么用，主要用来看爽哥。
- **油管直播/视频推送**
    - 支持推送预约直播信息及视频更新。
- **虎牙直播推送** *新增*
    - 不知道能看谁。
- 可配置的 **@全体成员** *新增*
    - 只建议单推群开启。
- **人脸识别**
    - 主要用来玩，支持二次元人脸。
- **倒放**
    - 主要用来玩。
- **Roll**
    - 没什么用的roll点。
- **签到**
    - 没什么用的签到。
- **权限管理**
    - 可配置整个命令的启用和禁用，也可对单个用户配置命令权限，防止滥用。
- **帮助**
    - 输出一些没什么帮助的信息。

<details>
  <summary>里命令</summary>

以下命令默认禁用，使用enable命令后才能使用

- **随机图片**
    - 由 [api.olicon.app](https://api.lolicon.app/#/) 提供

</details>

### 推送效果

<img src="https://user-images.githubusercontent.com/11474360/111737379-78fbe200-88ba-11eb-9e7e-ecc9f2440dd8.jpg" width="300">

### 用法示例

详细介绍及示例请查看：[详细示例](/EXAMPLE.md)

阁下可添加官方Demo机器人体验，QQ号：

- ~~368236249 （二号机）~~
- 1561991863 （初号机）

<details>
<summary>点此扫码二号机</summary>
<img src="https://user-images.githubusercontent.com/11474360/122684719-a8afe280-d239-11eb-9089-b8ce6613c819.jpg" width="300" height="450">
</details>

<details>
<summary>点此扫码初号机</summary>
<img src="https://user-images.githubusercontent.com/11474360/108590360-150afa00-739e-11eb-86f7-77f68d845505.jpeg" width="300" height="450">
</details>

~~推荐您优先选择二号机，由于目前初号机负载较高。~~

二号机暂时关闭，请选择初号机或者私人部署。

**尝试同时使用多个官方Demo机器人会导致您被暂时加入黑名单**

## 使用与部署

对于普通用户，推荐您选择使用开放的官方Demo机器人。

您也可以选择私人部署，[详见部署指南](/INSTALL.md)。

私人部署的好处：

- 保护您的隐私，bot完全属于您，我无法得知您bot的任何信息（我甚至无法知道您部署了一个私人bot）
- 稳定的@全体成员功能
- 可定制BOT账号的头像、名字、签名
- 减轻我的服务器负担
- 很cool

如果您遇到任何问题，或者有任何建议，可以加入**唯一指定交流群：755612788**

## 最近更新

请参考[更新文档](/UPDATE.md)。

NOTE：DDBOT正在进行重构，目前已经支持为DDBOT编写插件，来支持新的网站和订阅类型，如果您对此有兴趣，请查看[开发版本](https://github.com/Sora233/DDBOT/tree/refactor/concern#%E5%A2%9E%E5%8A%A0%E6%8E%A8%E9%80%81%E6%9D%A5%E6%BA%90)
。

**警告：开发版本目前尚处于实验阶段，无法兼容稳定版本，稳定版本下的订阅无法迁移到开发版本，如果您之前运行了稳定版本，请注意备份数据。**

## 常见问题FAQ

提问前请先查看[FAQ文档](/FAQ.md)，如果仍然未能解决，请咨询唯一指定交流群。

## 注意事项

- **bot只在群聊内工作，但命令可以私聊使用，以避免在群内刷屏**（少数次要娱乐命令暂不支持，详细列表请看用法指南）
- **建议bot秘密码设置足够强，同时不建议把bot设置为QQ群管理员，因为存在密码被恶意爆破的可能（包括但不限于盗号、广告等）**
- **您应当知道，bot账号可以人工登陆，请注意个人隐私**
- bot掉线无法重连时将自动退出，请自行实现保活机制
- bot使用 [buntdb](https://github.com/tidwall/buntdb) 作为embed database，会在当前目录生成文件`.lsp.db`
  ，删除该文件将导致bot恢复出厂设置，可以使用 [buntdb-cli](https://github.com/Sora233/buntdb-cli) 作为运维工具，但注意不要在bot运行的时候使用（buntdb不支持多写）

## 声明

- 您可以免费使用DDBOT进行其他商业活动，但不允许通过出租、出售DDBOT等方式进行商业活动。
- 如果您运营了私人部署的BOT，可以接受他人对您私人部署的BOT进行捐赠以帮助BOT运行，但该过程必须本着自愿的原则，不允许用BOT使用权来强制他人进行捐赠。
- 如果您使用了DDBOT的源代码，或者对DDBOT源代码进行修改，您应该用相同的开源许可（AGPL3.0）进行开源，并标明著作权。

## 贡献

*Feel free to make your first pull request.*

DDBOT使用 [MiraiGO-Template](https://github.com/Logiase/MiraiGo-Template) 进行开发，如果您使用了该框架，您可以将DDBOT嵌入您的程序里。

DDBOT提供一些可以使用的module：

- github.com/Sora233/DDBOT/lsp

该module即是DDBOT，嵌入DDBOT时需要引入这个module，同时需要在`bot.RefreshList()`这一行后面增加`lsp.Instance.PostStart(bot.Instance)`

- github.com/Sora233/DDBOT/msg-marker

该module可以自动把群聊和私聊消息标记为已读，go-cqhttp使用该机制来减少bot被识别。

- github.com/Sora233/DDBOT/logging

该module可以自动把消息内容打印到标准输出和`qq-logs`文件夹内。

- github.com/Sora233/DDBOT/miraigo-logging

该module可以自动把miraigo的日志打印到`miraigo-logs`内（用于帮助定位miraigo内部问题，所以不输出到标准输出）。

想要为开源做一点微小的贡献？

[Golang点我入门！](https://github.com/justjavac/free-programming-books-zh_CN#go)

您也可以选择点一下右上角的⭐星⭐

发现问题或功能建议请到 [issues](https://github.com/Sora233/DDBOT/issues)

其他用法问题请到**唯一指定交流群：755612788**

## 赞助

（排名按时间先后顺序）

|赞助者|渠道|金额|
|-----|----|----|
|VE-H Maw|爱发电|￥30.00|
|饱受突击的3737民|爱发电|￥168.00|
|刀光流水|爱发电|￥5.00|
|爱发电用户_4QBx|爱发电|￥5.00|
|XDMrSmile_鸟鸟|爱发电|￥30.00|

## 鸣谢

> [Goland](https://www.jetbrains.com/go/) 是一个非常适合Gopher的智能IDE，它极大地提高了开发人员的效率。

特别感谢 [JetBrains](https://jb.gg/OpenSource) 为本项目提供免费的 [Goland](https://www.jetbrains.com/go/) 等一系列IDE的授权

[<img src="https://user-images.githubusercontent.com/11474360/112592917-baa00600-8e41-11eb-9da4-ecb53bb3c2fa.png" width="200"/>](https://jb.gg/OpenSource)

## DDBOT:star:趋势图

[![Stargazers over time](https://starchart.cc/Sora233/DDBOT.svg)](https://starchart.cc/Sora233/DDBOT)
