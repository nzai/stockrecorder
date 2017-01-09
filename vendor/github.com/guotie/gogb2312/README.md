gogb2312
========

convert gb2312(cp936) to utf-8

Usage: 

# 转换gb2312编码为utf8编码

func ConvertGB2312(input []byte) (output []byte, err error, ic int, oc int)

参数：
   input 待转换的gb2312编码的byte数组

返回：
   output 转换后的utf8编码的byte数组
   err    错误码，如果成功为nil
   ic     成功转换的input的长度，如果input转换部分成功时，返回成功部分的长度
   oc     output的长度

output, err, ic, oc := ConvertGB2312(input)

func ConvertGB2312String(input string) (soutput string, err error, ic int, oc int)

与ConvertGB2312类似，不过参数和返回值由[]byte改为string