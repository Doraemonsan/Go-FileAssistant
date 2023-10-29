package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var logger *log.Logger
var authToken string

func uploadFile(w http.ResponseWriter, r *http.Request, dirPath string) {
	// 输出请求方法和其他信息到控制台
	fmt.Printf("请求方法：%s\n", r.Method)
	fmt.Printf("请求URL：%s\n", r.URL.String())
	fmt.Printf("请求客户端IP：%s\n", r.RemoteAddr)

	// 获取口令参数
	auth := r.FormValue("auth")
	logger.Printf("口令参数：%s\n", auth)

	if auth != authToken {
		// 口令认证失败，返回 401 错误代码
		logger.Println("口令认证失败")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
			Message: "非法上传请求",
			Time:    currentTime,
		}

		// 将JSON数据编码并返回给客户端
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		// 输出程序运行日志
		logger.Println(response)
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	// 获取当前所在目录参数
	dir := r.FormValue("dir")
	fmt.Println("dir:", dir)

	// 解码目录参数
	decodedDir, err := url.QueryUnescape(dir)
	if err != nil {
		// 处理解码错误
		fmt.Println("URL decoding error:", err)
		return
	}

	// 构建目标路径
	dstPath := dirPath + decodedDir + "/" + handler.Filename
	fmt.Println("dstPath:", dstPath)

	// 创建保存文件的目标路径
	dst, err := os.Create(dstPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dst.Close()

	// 将上传的文件拷贝到目标路径
	if _, err := io.Copy(dst, file); err != nil {
		fmt.Println("Copy error:", err)
		return
	}

	// 返回上传成功的消息
	fmt.Fprintf(w, "文件上传成功")

	// 输出程序运行日志
	logger.Printf("文件上传成功，目标路径：%s\n", dstPath)
}

func deleteFile(w http.ResponseWriter, r *http.Request, dirPath string) {
	// 输出请求方法和其他信息到控制台
	fmt.Printf("请求方法：%s\n", r.Method)
	fmt.Printf("请求URL：%s\n", r.URL.String())
	fmt.Printf("请求客户端IP：%s\n", r.RemoteAddr)

	// 获取口令参数
	auth := r.FormValue("auth")
	logger.Printf("口令参数：%s\n", auth)

	if auth != authToken {
		// 口令认证失败，返回 401 错误代码
		logger.Println("口令认证失败")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
			Message: "非法删除请求",
			Time:    currentTime,
		}

		// 将JSON数据编码并返回给客户端
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		// 输出程序运行日志
		logger.Println(response)
		return
	}

	// 获取要删除的文件路径参数
	filePath := r.FormValue("file")
	fmt.Println("filePath:", filePath)

	// 解码文件路径参数
	decodedFilePath, err := url.QueryUnescape(filePath)
	if err != nil {
		// 处理解码错误
		fmt.Println("URL decoding error:", err)
		return
	}

	// 构建完整的文件路径
	fullPath := dirPath + decodedFilePath
	fmt.Println("fullPath:", fullPath)

	// 删除文件
	err = os.Remove(fullPath)
	if err != nil {
		// 处理删除错误
		fmt.Println("File deletion error:", err)
		return
	}

	// 返回删除成功的消息
	fmt.Fprintf(w, "文件删除成功")

	// 输出程序运行日志
	logger.Printf("文件删除成功，文件路径：%s\n", fullPath)
}

func main() {
	// 解析命令行参数
	// 通过命令行参数设置端口号，默认为8080
	port := flag.String("p", "8080", "设置服务器监听的端口号")
	logLevel := flag.String("log", "info", "设置日志输出级别")
	auth := flag.String("a", "", "设置口令认证字符串")
	dir := flag.String("dir", ".", "设置网站根目录")
	flag.Parse()

	// 根据命令行参数初始化日志输出级别
	switch *logLevel {
	case "debug":
		logger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
	case "info":
		logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	case "warn":
		logger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime)
	case "error":
		logger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	default:
		logger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	}

	// 设置日志输出到控制台
	logger.SetOutput(os.Stdout)

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
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		deleteFile(w, r, *dir)
	})

	// 启动服务器，监听指定的端口
	fmt.Printf("服务器已启动，监听端口：%s\n", *port)
	http.ListenAndServe(":"+*port, nil)

}
