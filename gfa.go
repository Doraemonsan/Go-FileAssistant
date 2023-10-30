package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	authToken   string
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	warnLogger  = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// 处理上传文件请求逻辑
func uploadFile(w http.ResponseWriter, r *http.Request, dirPath string) {

	// 输出请求方法和其他调试信息到控制台
	debugLogger.Printf("请求客户端IP：%s\n", r.RemoteAddr)
	debugLogger.Printf("请求方法：%s\n", r.Method)
	debugLogger.Printf("请求URL：%s\n", r.URL.String())

	// 获取口令参数
	auth := r.FormValue("auth")

	// 口令验证逻辑
	if auth != authToken {
		// 口令认证失败，返回 401 错误代码
		warnLogger.Print("口令认证失败")
		warnLogger.Printf("客户端IP：%s\n", r.RemoteAddr)
		warnLogger.Printf("请求URL：%s\n", r.URL.String())
		warnLogger.Print("口令参数：", auth)
		http.Error(w, "401 Unauthorized, 口令验证失败，未经授权的访问", http.StatusUnauthorized)
		return
	}

	// 判断请求类型是否为POST
	if r.Method != http.MethodPost {
		// 获取当前系统的日期与时间
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		// 构建返回的JSON数据
		response := struct {
			IP      string `json:"ip"`
			Message string `json:"message"`
			Time    string `json:"time"`
		}{
			IP:      r.RemoteAddr,
			Message: "请求方法错误，非法上传请求",
			Time:    currentTime,
		}

		// 将JSON数据编码并返回给客户端
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		// 输出请求日志
		warnLogger.Print(response)
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		errorLogger.Print("获取文件失败：", err)
		return
	}
	defer file.Close()

	// 获取当前所在目录参数
	dir := r.FormValue("dir")
	debugLogger.Printf("原始路径： %s", dir)

	// 解码目录参数
	decodedDir, err := url.QueryUnescape(dir)
	if err != nil {
		// 处理解码错误
		errorLogger.Print("获取文件目录出错：", err)
		return
	}

	// 构建目标路径
	dstPath := dirPath + decodedDir + "/" + handler.Filename
	debugLogger.Printf("本地路径：%s", dstPath)

	// 创建保存文件的目标路径
	dst, err := os.Create(dstPath)
	if err != nil {
		errorLogger.Print("目录创建失败：", err)
		return
	}
	defer dst.Close()

	// 将上传的文件拷贝到目标路径
	if _, err := io.Copy(dst, file); err != nil {
		errorLogger.Print("文件复制出错：", err)
		return
	}

	// 输出文件保存成功信息和路径
	infoLogger.Print("文件上传成功，目标路径：", dstPath)
}

// 处理删除请求逻辑
func deleteFile(w http.ResponseWriter, r *http.Request, dirPath string) {

	// 输出请求方法和其他调试信息到控制台
	debugLogger.Printf("请求客户端IP：%s\n", r.RemoteAddr)
	debugLogger.Printf("请求方法：%s\n", r.Method)
	debugLogger.Printf("请求URL：%s\n", r.URL.String())

	// 获取口令参数
	auth := r.FormValue("auth")

	// 口令验证逻辑
	if auth != authToken {
		// 口令认证失败，返回 401 错误代码
		warnLogger.Print("口令认证失败")
		warnLogger.Printf("客户端IP：%s\n", r.RemoteAddr)
		warnLogger.Printf("请求URL：%s\n", r.URL.String())
		warnLogger.Print("口令参数：", auth)
		http.Error(w, "401 Unauthorized, 口令验证失败，经授权的访问", http.StatusUnauthorized)
		return
	}

	// 判断请求类型是否为DELETE
	if r.Method != http.MethodDelete {
		// 获取当前系统的日期与时间
		currentTime := time.Now().Format("2006-01-02 15:04:05")

		// 构建返回的JSON数据
		response := struct {
			IP      string `json:"ip"`
			Message string `json:"message"`
			Time    string `json:"time"`
		}{
			IP:      r.RemoteAddr,
			Message: "请求方法错误，非法删除请求",
			Time:    currentTime,
		}

		// 将JSON数据编码并返回给客户端
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		// 输出请求日志
		warnLogger.Print(response)
		return
	}

	// 获取要删除的文件路径参数
	filePath := r.FormValue("file")
	debugLogger.Printf("原始路径： %s", filePath)

	// 解码文件路径参数
	decodedFilePath, err := url.QueryUnescape(filePath)
	if err != nil {
		// 处理解码错误
		errorLogger.Print("获取文件目录出错：", err)
		return
	}

	// 构建完整的文件路径
	fullPath := dirPath + decodedFilePath
	debugLogger.Printf("本地路径：%s", fullPath)

	// 删除文件
	err = os.Remove(fullPath)
	if err != nil {
		// 处理删除错误
		errorLogger.Print("文件删除失败：", err)
		return
	}

	// 输出删除成功信息和路径
	infoLogger.Print("文件删除成功，文件路径：", fullPath)
}

func main() {

	// 解析命令行参数
	port := flag.String("p", "8080", "设置服务器监听的端口号")
	logLevel := flag.String("log", "info", "设置日志输出级别，支持：debug，info，warn，error")
	auth := flag.String("a", "", "设置口令认证字符串（4-16位字符串，仅支持数字，大小写字母和英文符号）")
	dir := flag.String("dir", ".", "设置网站根目录")
	flag.Parse()

	// 根据命令行参数选择要使用的日志记录器
	var selectedLogger *log.Logger
	switch *logLevel {
	case "debug":
		selectedLogger = debugLogger
	case "info":
		selectedLogger = infoLogger
	case "warn":
		selectedLogger = warnLogger
	case "error":
		selectedLogger = errorLogger
	default:
		selectedLogger = infoLogger
	}

	// 非debug级别日志时不输出调试信息
	if *logLevel != "debug" {
		debugLogger.SetOutput(io.Discard)
	}

	// 设置日志输出到控制台
	selectedLogger.SetOutput(os.Stdout)

	// 用户输入日志级别参数校验逻辑
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if *logLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		fmt.Println("日志级别参数无效，请输入debug、info、warn或error")
		flag.PrintDefaults()
		return
	}

	// 口令验证变量以及处理逻辑
	if *auth != "" {
		match, _ := regexp.MatchString(`^[0-9a-zA-Z!@#$%^&*()_+\-=[\]{};':"|,.<>/?]{4,16}$`, *auth)
		if !match {
			fmt.Println("请使用4-16位密码，仅支持数字，大小写字母和英文符号")
			return
		}
		authToken = *auth
	}

	// 设置/upload路由的处理函数为uploadFile函数
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadFile(w, r, *dir)
	})

	// 设置/delete路由的处理函数为deleteFile函数
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		deleteFile(w, r, *dir)
	})

	// 端口可用性验证逻辑
	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		fmt.Printf("服务启动失败：%s\n", err)
		return
	}
	listener.Close()

	// 在启动前打印服务监听信息，避免执行http.ListenAndServe语句后造成阻塞无法打印
	fmt.Printf("服务已启动，监听端口：%s\n", *port)

	// 启动服务器，监听指定的端口
	http.ListenAndServe(":"+*port, nil)

}
