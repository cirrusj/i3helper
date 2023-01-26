# i3helper
After switching to [i3](https://i3wm.org/docs/userguide.html), I always forgot the selected layout of my current window. As such I kept changing layouts and/or spawning new windows a few times, before I got the intended result.

i3helper will use [i3-ipc](https://pkg.go.dev/go.i3wm.org/i3/v4) and [dbus](https://pkg.go.dev/github.com/godbus/dbus/v5), in order to dynamically update and show the current layout in an [i3status-rs](https://github.com/greshake/i3status-rust) block.

![screenshot1](img/i3helper_1.png)
![screenshot2](img/i3helper_2.png)

You can install with:
```
go install github.com/cirrusj/i3helper@latest
```

## i3status-rs configuration
Add this in your i3status-rs configuration (`config.toml`):

```
[[block]]
block = "custom_dbus"
name = "LayoutMode"
initial_text = "layout"
```

You can verify the block is defined and working with something like:
```
gdbus introspect --session --dest i3.status.rs --object-path /LayoutMode
```

## i3 configuration
Don't forget to `exec` from your i3 config (or run manually).

```
exec --no-startup-id i3helper
```

## Switch to previously focused window
`i3helper` can also be used to bind a shortcut to switch focus to the previously focused window.
You need to bind `nop last_focus` to the shortcut you want. For example:

```
bindsym Mod1+Tab nop last_focus
```

## Command line parameters
You can use `-debug` to print messages on what `i3helper` is doing.