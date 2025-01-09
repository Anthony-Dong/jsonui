# JSONUI

1. 支持大JSON解析，优化了数据量大卡顿的问题 (10w+key 无压力)
2. 支持数据 Format 
3. 支持数据 拷贝

![](img/jsonui.gif)

## 安装
`go install -v github.com/anthony-dong/jsonui@master`

## 使用方式
```
cat example.json | jsonui

jsonui -r example.json

jsonui < example.json
```

### 快捷键

```shell
ctrl+e/ArrowDown = Move a line down   
ctrl+y/ArrowUp   = Move a line up     
ctrl+d/PageDown  = Move 15 line down  
ctrl+u/PageUp    = Move 15 line up    
c                = Copy node value    
f                = Format node data   
q/ctrl+c         = Exit               
h/?              = Toggle help message
tab              = Switch View
```

# copyright

https://github.com/gulyasm/jsonui