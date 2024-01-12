# lsapps

This is a simple utility to list all the desktop applications that you have
installed on your system by reading all the desktop files in
`$XDG_DATA_DIRS/applications` and `$HOME/.local/share/applications` and 
outputting them to `stdout`.

This is meant to be used with [dmenu_alias](https://github.com/JLi69/dmenu_alias)
as a way to generate output for [dmenu](https://tools.suckless.org/dmenu/) 
to use as input and allow the user to simply input an application on their system 
that they wish to start.

By default, `dmenu` lists all the binaries in a user's `$PATH` but this might
not be desirable as many of these applications are actually terminal applications
that lack a gui or terminal interface which makes them useless to run with 
`dmenu` and therefore this program functions as a way to filter those applications
down to those that have a terminal or gui interface and have a corresponding
desktop file.

## Build/Set up

Build and install:
```
go build .
sudo ./install
```

Set up `dmenu_path`:
```
#!/bin/sh

cachedir="${XDG_CACHE_HOME:-"$HOME/.cache"}"
cache="$cachedir/dmenu_run"

[ ! -e "$cachedir" ] && mkdir -p "$cachedir"

IFS=:
if stest -dqr -n "$cache" $PATH; then
	lsapps -g > $HOME/.config/dmenu_alias_list
	lsapps -n | sort -u | tee "$cache"
else
	cat "$cache"
fi
```
This script generates a `dmenu_alias_list` in `$HOME/.config` and then outputs
the applications that are available for user to run.

Set up `dmenu_run` (make sure to have 
[dmenu_alias](https://github.com/JLi69/dmenu_alias) installed):
```
#!/bin/sh
dmenu_path | dmenu "$@" | dmenu_alias | ${SHELL:-"/bin/sh"} &
```
This pipes the output of `dmenu_path` into `dmenu` and the output from
`dmenu` is piped into `dmenu_alias` where it is finally piped into the
user's shell to be run.

if you wish to uninstall the program, simply run `sudo ./uninstall`

## Usage
```
lsapps [-n|-e|-a|-g]
```

`-n` or `--names` lists the names of the applications, this can be used as
input to dmenu

`-e` or `--exec` lists the commands to execute each of the applicatons

`-a` or `--all` lists all the application names and executable names in the
format `name=exec`

`-g` or `--gen-alias` generates an output that can be used as a 
`dmenu_alias_list` for `dmenu_alias`. It outputs in the format `name=exec` and
only outputs those applications that have name that is different from their
exec string.

To generate a `dmenu_alias_list`, simply do the following:
```
lsapps -g > $HOME/.config/dmenu_alias_list
```

## License
MIT
