[![License](https://img.shields.io/:license-apache-blue.svg)](https://opensource.org/licenses/Apache-2.0) [![Go.Dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/lack/vine?tab=doc) [![Travis CI](https://api.travis-ci.org/lack-io/vine.svg?branch=master)](https://travis-ci.org/lack-io/vine) [![Go Report Card](https://goreportcard.com/badge/lack-io/vine)](https://goreportcard.com/report/github.com/vine-io/vine)

Vine (*/vaɪn/*) 一套简单、高效、可插拔的分布式 RPC 框架。

## 简介

**Vine** 基于 [go-micro](https://github.com/asim/go-micro) 框架，因此继承了它的优点。并在此基础上增加性能和可用性。

## 架构图

![vine架构图](https://vine-io.github.io/vine/docs/quickstart/overview/vine%E6%9E%B6%E6%9E%84%E5%9B%BE.png)

## 功能

**Vine** 抽象化分布式系统的内部细节. 以下是主要功能.

- **动态配置 (Dynamic Config)** - 从任意地方加载和热加载动态配置。配置接口提供了一种方法，可以从任何源（如环境变量，文件，etcd）加载应用级别配置。同时支持合并源、甚至定义回退。

- **数据存储 (Dao)** - 结合 gorm，对数据进行持久化操作。

- **缓存接口 (Cache)** - 支持内部数据缓存，保存临时数据。

- **服务发现 (Service discovery)** - 自动服务注册和名称解析。服务发现是微服务开发的核心功能。当服务A与服务B通讯时，需要知道服务B的IP地址等信息。**Vine** 内置(mdns)作为服务发现组件，它是零配置系统。

- **负载均衡 (Load Balancing)** - 基于服务发现的客户端负载均衡。当内部存在多个服务实例的地址时，我们需要一种方式确定路由到哪个节点，默认使用随机指定其中一个地址。同时在请求错误时重试不同的节点。

- **消息编码 (Message Encoding)** - 基于`Content-Type`的动态消息编码。Codec 为客户端和服务端提供 Go 类型的编码和解码，支持各种不同的类型。默认使用 protobuf 和 json。

- **gRPC 传输 (gRPC transport)** - 基于 gRPC 的请求响应，同时支持双向流。**Vine** 为同步通讯提供一个抽象，使发向服务的请求被自动解析、负载均衡、拨号和流化。

- **异步消息 (Async Messaging)** - 内置订阅发布模型，作为异步通讯和事件驱动架构。事件通知属于微服务开发中的核心模式。默认的消息系统为 HTTP 事件代理。

- **同步 (Synchronization)** - 分布式系统通常以最终一致性的方式构建。内置的 **Sync** 接口实现分布式锁和领导选举。

- **可插拔接口 (Pluggable Interfaces)** - 得益于 Go 语言的抽象特性。 **Vine** 为每个模块提供抽象接口，正因为如此，这些接口都是可插拔的。可以在 [github.com/vine-io/plugins](https://github.com/vine-io/plugins) 查询你需要的插件。

## 许可

Vine 遵守 Apache 2.0 开源许可.

## 文档
关于 `vine` 的文档可以参考 [vine 文档](https://vine-io.github.io/vine/)

## 支持
感谢 JetBrain 提供的支持

<img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" alt="https://www.jetbrains.com" style="width:30%;height:30%;" />