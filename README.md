# gomon

`gomon` is a tool for doing something when files change. It can do one of 2 things, or both:
- Re-run a command (or restart a web server) kind of like `nodemon`, `chokidar`, and `reflex`
- Launch Chrome and reload the page as needed, kind of like `livereload`, but with no setup required!

`gomon` is ridiculously simple. The whole thing could just be a bash script around `inotifywait`, but a bash script just isn't as cool as a Go tool 🔥 🔥 🔥

## Installation

```
go get -u github.com/mightyguava/gomon
```

## Usage

### Restart a command or web server

```
gomon echo hello world
gomon java Main.java
gomon go run main.go
gomon python -m SimpleHTTPServer
```

Now `gomon` will watch for any file changes within your current directory and subdirectories, and keep saying "hello world" every time something changes.

### Live reload a Chrome tab

```
gomon -url www.google.com
```

**Quirik:** Chrome opens in the background in a new window. You'll have to manually switch to it. I haven't figured out how to make it open in the foreground https://github.com/chromedp/chromedp/issues/171.i

### Restart a web server AND live reload a Chrome tab

```
gomon -url localhost:8000 npm start
```

### Bash script

You can run a whole bash script. It's ok that your script exits. `gomon` will restart it whenever you change something.

```
gomon sh -c 'echo hello; echo world; echo i can chain forever'
gomon sh -c my_super_awesome_script.sh
```

## How it works

`gomon` uses [notify](https://github.com/rjeczalik/notify) to watch the current directory for changes, and restart the command if anything changed. It doesn't care about what language your code is written in, just that you give it a command that can be executed on *nix.

`gomon` uses [chromedp](https://github.com/chromedp/chromedp) to control a Chrome window, reloading it when a file changes.

## Similar tools

- https://github.com/cespare/reflex - Really cool tool for re-running commands
- https://github.com/Unknwon/bra - Does similar things, with more flexibility but has configuration
- https://github.com/johannesboyne/gomon - Does similar things, but is a node script installed via NPM, for Go (why????)
