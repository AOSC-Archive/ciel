# Contributing to Ciel

Welcome!

You may use the following shell commands to get the lastest code.
```
git clone https://github.com/AOSC-Dev/ciel
git submodule update --init
```

If you want to change anything in `/src/ciel-driver`, it is at repository https://github.com/AOSC-Dev/ciel-driver.

## Files & Directories

Plugins are placed at `/plugin`, this directory will be copied to `/usr/libexec/ciel-plugin` when you do `make install`.

You MUST keep the plugin executable, `chmod a+x ciel-myplugin`.

A plugin could be a shell script, a python script, a binary...

Use `./format.sh` to format your code BEFORE committing, I'll suggest you to make a symbol link at `/.git/hook/pre_commit` to `format.sh`.


## Coding Style

Keep it simple and readable. Come on, you can do it right.

But, again, **don't forget to use `./format.sh` after coding things.** :)

## Commit Message

```
about_something: do what
or
about_something: do what; fix #123
```

You may check the `git log --oneline` out, there will be many examples for you.
