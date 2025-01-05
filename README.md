# copyright 

https://github.com/gulyasm/jsonui

# JSONUI
[![](https://travis-ci.org/gulyasm/jsonui.svg?branch=master)](https://travis-ci.org/gulyasm/jsonui) [![](https://goreportcard.com/badge/github.com/anthony-dong/jsonui)](https://goreportcard.com/report/github.com/anthony-dong/jsonui)

`jsonui` is an interactive JSON explorer in your command line. You can pipe any JSON into `jsonui` and explore it, copy the path for each element.

![](img/jsonui.gif)

## Install
`go install -v github.com/anthony-dong/jsonui@master`

## Binary Releases
[Binary releases are availabe](https://github.com/anthony-dong/jsonui/releases)

## Usage
Just use the standard output:
```
cat example.json | jsonui

jsonui -r example.json

jsonui < example.json
```

### Keys

```shell
ctrl+e/ArrowDown = Move a line down   
ctrl+y/ArrowUp   = Move a line up     
ctrl+d/PageDown  = Move 15 line down  
ctrl+u/PageUp    = Move 15 line up    
c                = Copy node value    
f                = Format node data   
q/ctrl+c         = Exit               
h/?              = Toggle help message
```


## Acknowledgments
Special thanks for [asciimoo](https://github.com/asciimoo) and the [wuzz](https://github.com/asciimoo/wuzz) project for all the help and suggestions.  

