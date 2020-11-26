# pulsar-client-go

[English](README.md) | [中文版](README_cn.md)

## 介绍

A Go client library for the [Apache Pulsar](https://pulsar.incubator.apache.org/) project.

基于[Comcast](https://github.com/Comcast/pulsar-client-go) 和 [wolfstudy](https://github.com/Comcast/pulsar-client-go)两个项目开发，完全用go实现的pulsar-client

## 主要优化以下几个点

1. 修复java-client默认使用batch方式推送消息，go-client接收消息时会出现乱码

2. 修复failover模式下，在新的consumer加入时，会出现部分消息重复消费的问题

3. 修复pulsar-broker 进行负载均衡时，会出现topic重启以及broker迁移，此时会导致部分消息重复消费

4. 优化初始化时内存占用过高的问题


## 技术支持

你可以通过以下方式获得Tua开发者技术支持：

- 涂鸦帮助中心: [https://support.tuya.com/zh/help](https://support.tuya.com/zh/help)
- 涂鸦技术工单平台: [https://iot.tuya.com/council](https://iot.tuya.com/council)
