<div style="text-align: center"><img src="/img/dapr_logo.svg" height="120px">
<h2>构建安全可靠微服务的 API</h2>
</div>

[![Go Report][go-report-badge]][go-report-url] [![OpenSSF][openssf-badge]][openssf-url] [![Docker Pulls][docker-badge]][docker-url] [![Build Status][actions-badge]][actions-url] [![Test Status][e2e-badge]][e2e-url] [![Code Coverage][codecov-badge]][codecov-url] [![License: Apache 2.0][apache-badge]][apache-url] [![FOSSA Status][fossa-badge]][fossa-url] [![TODOs][todo-badge]][todo-url] [![Good First Issues][gfi-badge]][gfi-url] [![discord][discord-badge]][discord-url] [![YouTube][youtube-badge]][youtube-link] [![Bluesky][bluesky-badge]][bluesky-link] [![X/Twitter][x-badge]][x-link]

[go-report-badge]: https://goreportcard.com/badge/github.com/dapr/dapr
[go-report-url]: https://goreportcard.com/report/github.com/dapr/dapr
[openssf-badge]: https://www.bestpractices.dev/projects/5044/badge
[openssf-url]: https://www.bestpractices.dev/projects/5044
[docker-badge]: https://img.shields.io/docker/pulls/daprio/daprd?style=flat&logo=docker
[docker-url]: https://hub.docker.com/r/daprio/dapr
[apache-badge]: https://img.shields.io/github/license/dapr/dapr?style=flat&label=License&logo=github
[apache-url]: https://github.com/dapr/dapr/blob/master/LICENSE
[actions-badge]: https://github.com/dapr/dapr/workflows/dapr/badge.svg?event=push&branch=master
[actions-url]: https://github.com/dapr/dapr/actions?workflow=dapr
[e2e-badge]: https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/dapr-bot/14e974e8fd6c6eab03a2475beb1d547a/raw/dapr-test-badge.json
[e2e-url]: https://github.com/dapr/dapr/actions?workflow=dapr-test&event=schedule
[codecov-badge]: https://codecov.io/gh/dapr/dapr/branch/master/graph/badge.svg
[codecov-url]: https://codecov.io/gh/dapr/dapr
[fossa-badge]: https://app.fossa.com/api/projects/custom%2B162%2Fgithub.com%2Fdapr%2Fdapr.svg?type=shield
[fossa-url]: https://app.fossa.com/projects/custom%2B162%2Fgithub.com%2Fdapr%2Fdapr?ref=badge_shield
[todo-badge]: https://badgen.net/https/api.tickgit.com/badgen/github.com/dapr/dapr
[todo-url]: https://www.tickgit.com/browse?repo=github.com/dapr/dapr
[gfi-badge]:https://img.shields.io/github/issues-search/dapr/dapr?query=type%3Aissue%20is%3Aopen%20label%3A%22good%20first%20issue%22&label=Good%20first%20issues&style=flat&logo=github
[gfi-url]:https://github.com/dapr/dapr/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22
[discord-badge]: https://img.shields.io/discord/778680217417809931?label=Discord&style=flat&logo=discord
[discord-url]: http://bit.ly/dapr-discord
[youtube-badge]:https://img.shields.io/youtube/channel/views/UCtpSQ9BLB_3EXdWAUQYwnRA?style=flat&label=YouTube%20views&logo=youtube
[youtube-link]:https://youtube.com/@daprdev
[bluesky-badge]:https://img.shields.io/badge/Follow-%40daprdev.bsky.social-0056A1?logo=bluesky
[bluesky-link]:https://bsky.app/profile/daprdev.bsky.social
[x-badge]:https://img.shields.io/twitter/follow/daprdev?logo=x&style=flat
[x-link]:https://twitter.com/daprdev

[开发环境与编译指南](./docs/development/developing-dapr_cn.md)  
[body解析中间件功能](./examples/middleware-body-demo/README.md)

Dapr 是一套集成的 API，内置了构建分布式应用程序的最佳实践和模式。Dapr 通过提供开箱即用的功能（如工作流、发布/订阅、状态管理、密钥存储、外部配置、绑定、角色、分布式锁和密码学）将您的开发效率提高 20-40%。您可以从内置的安全性、可靠性和可观察性功能中受益，因此无需编写样板代码即可实现生产就绪的应用程序。

借助 Dapr（一个已毕业的 CNCF 项目），平台团队可以配置复杂的设置，同时向应用程序开发团队公开简单的接口，使他们更容易构建高度可扩展的分布式应用程序。许多平台团队已经采用 Dapr 来为基于 API 的基础设施交互提供治理和黄金路径。

![Dapr 概览](./img/overview.png)

我们是云原生计算基金会（CNCF）的毕业项目。
<p align="center"><img src="https://raw.githubusercontent.com/kedacore/keda/main/images/logo-cncf.svg" height="75px"></p>

## 目标

- 使使用*任何*语言或框架的开发人员都能编写分布式应用程序
- 通过提供最佳实践构建块来解决开发人员在构建微服务应用程序时面临的难题
- 以社区驱动、开放和供应商中立为原则
- 获得新的贡献者
- 通过开放 API 提供一致性和可移植性
- 在云和边缘环境中保持平台无关性
- 拥抱可扩展性并提供可插拔组件，避免供应商锁定
- 通过高性能和轻量级特性支持物联网和边缘场景
- 可从现有代码逐步采用，无运行时依赖

## 工作原理

Dapr 向每个计算单元注入一个边车（容器或进程）。边车与事件触发器交互，并通过标准 HTTP 或 gRPC 协议与计算单元通信。这使得 Dapr 能够支持所有现有和未来的编程语言，而无需您导入框架或库。

Dapr 通过标准 HTTP 动词或 gRPC 接口提供内置状态管理、可靠消息传递（至少一次传递）、触发器和绑定。这允许您遵循相同的编程范式编写无状态、有状态和类似角色的服务。您可以自由选择一致性模型、线程模型和消息传递模式。

Dapr 在 Kubernetes 上原生运行，作为您机器上的自托管二进制文件，在 IoT 设备上，或作为可注入任何系统的容器，无论是在云中还是本地。

Dapr 使用可插拔的组件状态存储和消息总线（如 Redis）以及 gRPC 来提供广泛的通信方法，包括使用 gRPC 的直接 dapr-to-dapr 通信和具有保证传递和至少一次语义的异步发布-订阅。

## 为什么选择 Dapr？

编写高性能、可扩展和可靠的分布式应用程序是困难的。Dapr 为您带来了经过验证的模式和实践。它将事件驱动和角色语义统一到一个简单、一致的编程模型中。它支持所有编程语言而不会被框架锁定。您不会暴露于低级原语，如线程、并发控制、分区和扩展。相反，您可以使用您选择的熟悉的 Web 框架实现简单的 Web 服务器来编写代码。

Dapr 在线程和状态一致性模型方面具有灵活性。如果您选择，可以利用多线程，并且可以在不同的一致性模型中进行选择。这种灵活性使您能够在没有人为约束的情况下实现高级场景。Dapr 是独特的，因为您可以在平台和底层实现之间无缝转换，而无需重写代码。

## 功能特性

* 具有可插拔提供程序和至少一次语义的事件驱动发布-订阅系统
* 具有可插拔提供程序的输入和输出绑定
* 具有可插拔数据存储的状态管理
* 一致的服务到服务发现和调用
* 可选的有状态模型：强/最终一致性，首次写入/最后写入获胜
* 跨平台虚拟角色
* 从安全密钥保管库检索密钥的密钥管理
* 速率限制
* 内置[可观察性](https://docs.dapr.io/concepts/observability-concept/)支持
* 使用专用操作器和 CRD 在 Kubernetes 上原生运行
* 通过 HTTP 和 gRPC 支持所有编程语言
* 多云、开放组件（绑定、发布-订阅、状态）来自 Azure、AWS、GCP
* 可在任何地方运行，作为进程或容器化
* 轻量级（58MB 二进制文件，4MB 物理内存）
* 作为边车运行 - 无需特殊 SDK 或库
* 专用 CLI - 开发人员友好的体验，易于调试
* 支持 Java、.NET Core、Go、Javascript、Python、Rust 和 C++ 的客户端

## 开始使用 Dapr

请查看我们文档中的[入门指南](https://docs.dapr.io/getting-started/)。

## 快速入门和示例

* 查看[快速入门仓库](https://github.com/dapr/quickstarts)中的代码示例，这些示例可以帮助您开始使用 Dapr。
* 在 Dapr [示例仓库](https://github.com/dapr/samples)中探索更多示例。

## 社区
我们希望得到您的贡献和建议！最简单的贡献方式之一是参与邮件列表上的讨论、在即时通讯中聊天或参加双周社区电话会议。
有关社区参与、开发人员和贡献指南等更多信息，请前往 [Dapr 社区仓库](https://github.com/dapr/community#dapr-community)。

### 联系我们

如果您有任何问题，请随时联系我们，我们将确保尽快回答！

| 平台  | 链接        |
|:----------|:------------|
| 💬 Discord（首选） | [![Discord Banner](https://discord.com/api/guilds/778680217417809931/widget.png?style=banner2)](https://aka.ms/dapr-discord)
| 💭 LinkedIn | [@daprdev](https://www.linkedin.com/company/daprdev)
| 🦋 BlueSky | [@daprdev.bsky.social](https://bsky.app/profile/daprdev.bsky.social)
| 🐤 Twitter | [@daprdev](https://twitter.com/daprdev)

### 社区电话会议

每两周我们举办一次社区电话会议，展示新功能、回顾即将到来的里程碑，并进行问答。欢迎所有人参加！

📞 访问[即将到来的 Dapr 社区电话会议](https://github.com/dapr/community/issues?q=is%3Aissue%20state%3Aopen%20label%3A%22community%20call%22)了解即将到来的日期和会议链接。

📺 访问 https://www.youtube.com/@DaprDev/streams 观看以前的社区电话会议直播。

### 视频和播客

我们有各种主题演讲、播客和演示文稿可供参考和学习。

📺 访问 https://docs.dapr.io/contributing/presentations/ 查看以前的演讲和幻灯片，或访问我们的 YouTube 频道 https://www.youtube.com/@DaprDev/videos。

### 为 Dapr 做贡献

请查看[开发指南](https://docs.dapr.io/contributing/)开始构建和开发。

## 仓库

| 仓库 | 描述 |
|:-----|:------------|
| [Dapr](https://github.com/dapr/dapr) | 您当前所在的主仓库。包含 Dapr 运行时代码和概述文档。
| [CLI](https://github.com/dapr/cli) | Dapr CLI 允许您在本地开发机器或 Kubernetes 集群上设置 Dapr，提供调试支持，启动和管理 Dapr 实例。
| [文档](https://docs.dapr.io) | Dapr 的文档。
| [快速入门](https://github.com/dapr/quickstarts) | 此仓库包含一系列简单的代码示例，突出了 Dapr 的主要功能。
| [示例](https://github.com/dapr/samples) | 此仓库包含社区维护的各种 Dapr 用例示例。
| [组件贡献](https://github.com/dapr/components-contrib) | 组件贡献的目的是为构建分布式应用程序提供开放的、社区驱动的可重用组件。
| [仪表板](https://github.com/dapr/dashboard) | Dapr 的通用仪表板
| [Go-sdk](https://github.com/dapr/go-sdk) | Go 的 Dapr SDK
| [Java-sdk](https://github.com/dapr/java-sdk) | Java 的 Dapr SDK
| [JS-sdk](https://github.com/dapr/js-sdk) | JavaScript 的 Dapr SDK
| [Python-sdk](https://github.com/dapr/python-sdk) | Python 的 Dapr SDK
| [Dotnet-sdk](https://github.com/dapr/dotnet-sdk) | .NET 的 Dapr SDK
| [Rust-sdk](https://github.com/dapr/rust-sdk) | Rust 的 Dapr SDK
| [Cpp-sdk](https://github.com/dapr/cpp-sdk) | C++ 的 Dapr SDK
| [PHP-sdk](https://github.com/dapr/php-sdk) | PHP 的 Dapr SDK

## 行为准则

请参考我们的 [Dapr 社区行为准则](https://github.com/dapr/community/blob/master/CODE-OF-CONDUCT.md)