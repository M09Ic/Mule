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
* flag参数的意思是,在默认的Data/DefDict文件夹下的字典可以通过调用其文件名的方式调用不用带后缀
* 如果使用自己的字典,即-d参数则会自动在Data/DefDict文件夹下转换为json格式,之后只需要通过文件名调用
* 每次扫描结束后会更新一次字典,自我优化迭代

example:
Mule Brute -u http://baidu.com -f php -u /root/aaa.txt -t 100 -o ./res3.log


## 部分设计细节
1. 会没100次请求后,检测一下是否ip被block了,如果被block则停止
2. 使用ctrl+C后会跳过当前目标进入下一个
3. 在Data目录下有个SpecialList的目录,里面的文件exwildcard是为了部分中间件或者防火墙等对敏感文件后缀做了特殊处理,而容易产生大量误报,如果你发现在目录爆破中出现一个后缀有大量误报,请扩展他.
    1. 扩展格式为/$$.ext,$$为占位符用于替换测试




## TODO
1. waf 识别? 思考了一下有waf也就不爆破,不如乘着没被封多爆破几下

