# Go-FileAssistant
### 一个简单的Golang文件处理接口，用于处理NGINX搭建的简单静态HTTP文件服务器的文件操作

### 使用方法:
1:源码运行
go run gfa.go

2：运行二进制文件
./gfa

当前支持:
* -a 设置口令认证字符串(default NONE)
*  -dir 设置网站根目录 (default ".")
*  -log 设置日志输出级别 (default "info")
*  -p 设置服务器监听的端口号 (default "8080")
***
### 上传接口请求方法：
Request Method:POST
/upload?dir=youdir&auth=youauth
### 删除接口请求办法:
Request Method:DELETE
/delete?file=%2Fyoufile.txt&auth=youauth
[示例页面](https://files.gitlx.com)
