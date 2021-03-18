## Introduction

A Go client library for the [Apache Pulsar](https://pulsar.incubator.apache.org/) project.

Developed based on two projects [Comcast](https://github.com/Comcast/pulsar-client-go) and [wolfstudy](https://github.com/Comcast/pulsar-client-go), completely using go Implemented pulsar-client.

## Mainly optimize the following points

* Fixed the problem of messy code when go-client receives messages that are pushed by java-client through batch.

* Fixed the problem of repeated consumption of some messages when a new consumer joins in failover mode.

* Fixed the problem of repeated consumption of some messages during topic restart and broker migration, when pulsar-broker executes load balancing.

* Optimized the problem of high memory usage during initialization.

## Prepare
Go 1.11+.

## Example
For examples of producers and consumers, see [cli](https://github.com/tuya/tuya-pulsar-client-go/blob/main/cmd/cli/main.go).

## Technical Support

You can get Tuya developer technical support in the following ways:

* [Tuya Help Center](https://support.tuya.com/zh/help)
* [Tuya technical ticket platform](https://iot.tuya.com/council)
