# Mule

> 一个沙雕的毕业设计
>
> 一个垃圾的目录扫描工具
> 
> 名字叫:可变异的Web目录模糊测试工具



## 基本用法

Usage:

Available Commands:

    Brute       a weak Directory Blasting
    help        Help about any command

Mule Brute [flags]

Flags:

    -C, --Cookie string         Request's Cookie
    -H, --Headers stringArray   Request's Headers
    -t, --Thread int            the size of thread pool (default 30)
    -d, --dic string            dictionary to brute
    -f, --flag string           use default dictionary in /Data
    -h, --help                  help for Brute
    -o, --output string         output res default in ./res.log (default "./res.log")
    --timeout int           request's timeout (default 2)
    -u, --url string            brute target(currently only single url)

* 默认线程是30,实际测试出来100也完全没事,怕路由器不行,就还是30吧
* 默认日志是放在当前目录下的res.log中,同时会在console中输出
* 暂时还未支持批量爆破,现在是单个url进行爆破


## TODO

1. 加上进度条
2. 加上waf识别

