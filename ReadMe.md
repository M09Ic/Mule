# Mule

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


Flags:
   * -C, --Cookie string         Request's Cookie
   * -H, --Headers stringArray   Request's Headers
   * -m, --mod string            default to brute dict, set '-m host' to brute host
   * -t, --Thread int            the size of thread pool (default 30)
   * --blacklist ints        the black list of statuscode
   * -b, --block int             the number of auto stop brute (default 10)
   * -d, --dic string            dictionary to brute
   * -f, --flag string           use default dictionary in /Data
   * -h, --help                  help for Brute
   * -o, --output string         默认会根据你本次扫描命令行输出在log文件夹下
   * --timeout int           request's timeout (default 5)
   * -u, --url string            brute target(currently only single url)
   * -U, --urls string           targets from file
   * --noconsole                 结果不输出在console，只有基本信息
   * --noupdate                  不生成和更新内部字典
   * --nolog                     不生成日志
   * --Poolsize
   * -r, --range                 设置爆破字典的范围，如2k的字典可以爆破前100即 -r 0-100
   * -P                          调整poolsize 即同时爆破的url数量
   * -follow                     是否跟随跳转
   * -a                          自动随机字典
   > 内置字典为A 大写字母，a小写字母，n数字，§用于分隔内置和自己添加，例如A§1为使用内置大写26个字母加数字1随机生成
   * -c                          随机字典的位数


* 默认线程是30,实际测试出来100也完全没事,怕路由器不行,就还是30吧
* 默认日志是放在当前目录下的res.log中,同时会在console中输出
* 这里有一部分默认字典
* flag参数的意思是,在默认的Data/DefDict文件夹下的字典可以通过调用其文件名的方式调用不用带后缀
* 如果使用自己的字典,即-d参数则会自动在Data/DefDict文件夹下转换为json格式,之后只需要通过文件名调用
* 每次扫描结束后会更新一次字典,自我优化迭代

example:
Mule Brute -u http://baidu.com -f inter -U /root/aaa.txt -t 100
inter 为曾华字典
inter2 为王一帆字典

## 部分设计细节
1. 会没100次请求后,检测一下是否ip被block了,如果被block则停止,block次数默认为3,可以自行根据网络状况更改,网络差的话也容易触发
2. 使用ctrl+C后会跳过当前目标进入下一个
3. 在Data目录下有个SpecialList的目录,里面的文件exwildcard是为了部分中间件或者防火墙等对敏感文件后缀做了特殊处理,而容易产生大量误报,如果你发现在目录爆破中出现一个后缀有大量误报,请扩展他.
    1. 扩展格式为/$$.ext,$$为占位符用于替换测试







