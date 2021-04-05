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
    --timeout int               request's timeout (default 2)
    -u, --url string            brute target
    -U, --urls string           targets from file

* 默认线程是30,实际测试出来100也完全没事,怕路由器不行,就还是30吧
* 默认日志是放在当前目录下的res.log中,同时会在console中输出
* 这里有一部分默认字典




## TODO
1. waf 识别? 思考了一下有waf也就不爆破了不如爆破几下
