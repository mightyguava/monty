# gomon

`gomon` is a tool for restarting a command when files change, kind of like `nodemon` and `chokidar`. It's ridiculously simple.

All it does is use [fsnotify](https://github.com/fsnotify/fsnotify) to watch the current directory for changes, and restart the command if anything changed. It doesn't care about what language your code is written in, just that you give it a command that can be executed on *nix.

This could just be a bash script around `inotifywait`, but a bash script just isn't as cool as a Go tool 🔥 🔥 🔥

## Usage

### Basic

```
gomon echo hello world
```

Now `gomon` will watch for any file changes within your current directory, and keep saying "hello world". It doesn't do subdirectories yet, that's a TODO.

### Web Servers

It can do web servers too!

```
gomon python -m SimpleHTTPServer
```

### Fancy

You can run a whole bash script. It's ok that your script exits. `gomon` will restart it whenever you change something.

```
gomon sh -c 'echo hello; echo world; echo i can chain forever'
```

## Installation

You should download it! It's free.

```
go get -u github.com/mightyguava/gomon
```

## Similar tools

- https://github.com/Unknwon/bra - Does similar things, with more flexibility but has configuration
- https://github.com/johannesboyne/gomon - Does similar things, but is a node script installed via NPM, for Go (why????)