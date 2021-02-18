# pulsar-client-go

[English](README.md) | [中文版](README_cn.md)

## Introduction

Tuya pulsar client SDK for Golang based on [Comcast](https://github.com/Comcast/pulsar-client-go) and [wolfstudy](https://github.com/Comcast/pulsar-client-go). A Go client library for the [Apache Pulsar](https://pulsar.incubator.apache.org/) project.

## What's new

1. Fixed the problem of messy code when go-client receives messages that are pushed by java-client through batch.

2. Fixed the problem of repeated consumption of some messages when a new consumer joins in failover mode.

3. Fixed the problem of repeated consumption of some messages during topic restart and broker migration, when pulsar-broker executes load balancing.

4. Optimized the problem of high memory usage during initialization.



## Support

You can get support from Tuya with the following methods:

- [Tuya Smart Help Center](https://support.tuya.com/en/help)
- [Technical Ticket](https://iot.tuya.com/council)

