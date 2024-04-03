# Spring Initializer Go

A [Spring Initializer](https://github.com/spring-io/start.spring.io) client
written in Go.

![preview](./assets/preview-1.png)

## Motivation

As a software developer, I frequently found myself writing extensive code,
leading to discomfort and wrist pain. Seeking solutions, I explored ways to
reduce mouse usage. Transitioning many aspects of my workflow to the terminal
proved both comfortable and efficient.

However, during the initiation of a new Java project, I encountered a roadblock:
the absence of a terminal-based user interface (TUI) version of **Spring
Initializr**. Determined to bridge this gap, I embarked on creating a solution
tailored to my needs.

## Installation

### Manual

Clone the repository and run: `make compile-current` or `go build ./cmd/spring-initializer/`
in the root of the repository.

### Go install (Recommended)

#### Prerequisites

- [Golang](https://go.dev/doc/install)

```bash
go install github.com/eslam-allam/spring-initializer-go/cmd/spring-initializer@latest
```

### Pre Compiled Binary

You can also grab one of the pre-compiled binaries from the
[Releases Section](https://github.com/eslam-allam/spring-initializer-go/releases)
and place it in a folder currently in PATH.

> Compiled binaries may not always be up to date so if you want the latest
> features, the other methods are recommended.

## Usage

This app has a similar interface to official [Web Spring Initializer](https://start.spring.io/).
Just run the app using `spring-initializer` and you will be able to see a list of
available key maps at the bottom of the screen.

You may also pass the target directory as a positional command line argument as
follows:

```bash
spring-initializer 'some-directory/some-other-directory' # Relative directory
```

```bash
spring-initializer '/home/eslamallam/personal_projects/java/something' # Absolute directory
```

```bash
spring-initializer '~/projects/spring' # ~ will be expanded to $HOME
```

The directory will be created if it doesn't exist.

## Todo

- [x] Add ability to pick project folder.
- [ ] Add description to dependency entries.
- [ ] Make the UI more intuitive.
- [ ] Refactor this unsightly code.
- [ ] Add confirmation message when creating a new project.

## Issues

This project was mainly created to make my workflow more convenient but tickets
are more than welcome. If you encounter any crashes or if there's something you
wish was done differently, don't hesitate to open an issue or PR if you think
you can handle it yourself.
