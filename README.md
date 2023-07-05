# stevedore

![alt text](stevedore.png "stevedore")

[![Maintenance](https://img.shields.io/badge/Maintained%3F-yes-green.svg)](https://GitHub.com/jameswoolfenden/stevedore/graphs/commit-activity)
[![Build Status](https://github.com/JamesWoolfenden/stevedore/workflows/CI/badge.svg?branch=master)](https://github.com/JamesWoolfenden/stevedore)
[![Latest Release](https://img.shields.io/github/release/JamesWoolfenden/stevedore.svg)](https://github.com/JamesWoolfenden/stevedore/releases/latest)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/JamesWoolfenden/stevedore.svg?label=latest)](https://github.com/JamesWoolfenden/stevedore/releases/latest)
![Terraform Version](https://img.shields.io/badge/tf-%3E%3D0.14.0-blue.svg)
[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)
[![checkov](https://img.shields.io/badge/checkov-verified-brightgreen)](https://www.checkov.io/)
[![Github All Releases](https://img.shields.io/github/downloads/jameswoolfenden/stevedore/total.svg)](https://github.com/JamesWoolfenden/stevedore/releases)

Stevedore manages labels in Dockerfiles and their layers

## Table of Contents

<!--toc:start-->
- [stevedore](#stevedore)
  - [Table of Contents](#table-of-contents)
  - [Install](#install)
    - [MacOS](#macos)
    - [Windows](#windows)
    - [Docker](#docker)
  - [Usage](#usage)

<!--toc:end-->

## Install

Download the latest binary here:

<https://github.com/JamesWoolfenden/stevedore/releases>

Install from code:

- Clone repo
- Run `go install`

Install remotely:

```shell
go install  github.com/jameswoolfenden/stevedore@latest
```

### MacOS

```shell
brew tap jameswoolfenden/homebrew-tap
brew install jameswoolfenden/tap/stevedore
```

### Windows

I'm now using Scoop to distribute releases, 
it's much quicker to update and easier to manage than previous methods,
you can install scoop from <https://scoop.sh/>.

Add my scoop bucket:

```shell
scoop bucket add iac https://github.com/JamesWoolfenden/scoop.git
```

Then you can install a tool:

```bash
scoop install stevedore
```

### Docker

```shell
docker pull jameswoolfenden/stevedore
docker run --tty --volume /local/path/to/tf:/tf jameswoolfenden/stevedore scan -d /tf
```

<https://hub.docker.com/repository/docker/jameswoolfenden/stevedore>

## Usage

### Directory scan

This will look for the .github/workflow folder and update all the files it finds there, and display a diff of the changes made to each file:

```bash
$stevedore label -d .
```

### Individual file scan

```bash
$stevedore label -f Dockerfile
```

## Help

```bash
NAME:
   stevedore - Update Dockerfile labels

USAGE:
   stevedore [global options] command [command options] [arguments...]

VERSION:
   9.9.9

AUTHOR:
   James Woolfenden <jim.wolf@duck.com>

COMMANDS:
   label, l    Updates Dockerfiles labels
   version, v  Outputs the application version
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Building

```go
go build
```

or

```Make
Make build
```

## Extending

Log an issue, a pr or send an email to jim.wolf @ duck.com.
