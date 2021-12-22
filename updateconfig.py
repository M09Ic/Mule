import json, yaml
import sys, io, os, zlib
from base64 import b64encode

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf8')


def compress(s):
    flatedict = bytes(', ":'.encode())
    com = zlib.compressobj(level=9, zdict=flatedict)
    return b64encode(zlib.compress(s.encode())[2:-4]).decode()





def fingerload(filename):
    tcpfinger = open("fingers/%s" % filename, "r", encoding="utf-8")
    tcpjsonstr = tcpfinger.read()
    tcpjsonstr = tcpjsonstr.replace("\\0", "\\u0000").replace("\\x", "\\u00")
    j = json.loads(tcpjsonstr)
    j = sorted(j, key=lambda x: x["level"])
    return j


if __name__ == "__main__":
    httpfingers = fingerload("httpfingers.json")
    md5fingers = json.loads(open("fingers/md5fingers.json", "r", encoding="utf-8").read())
    mmh3fingers = json.loads(open("fingers/mmh3fingers.json", "r", encoding="utf-8").read())
    port = json.loads(open("fingers/port.json", "r", encoding="utf-8").read())
    f = open("utils/finger.go", "w", encoding="utf-8")
    base = '''package utils

func LoadConfig(typ string)[]byte  {
	if typ=="http"{
		return `%s`
	}else if typ =="md5"{
     		return `%s`
    }else if typ =="mmh3"{
         	return `%s`
    }
	return []byte{}
}
	'''

    f.write(base % (
                    json.dumps(httpfingers),
                   json.dumps(md5fingers),
                    json.dumps(mmh3fingers),

                    ))
#     print(compress(json.dumps(tcpfingers)))
    print("fingerprint update success")

