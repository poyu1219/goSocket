package main

import (
	"net"

	"./utils"

	//	"strconv"
	"strconv"
)

func main() {
	//啟動服務器, config.yaml 是設定檔
	startServer("./conf/config.yaml")
}

// 啟動服務器
func startServer(configpath string) {
	//	setup a socket and listen the port
	//yaml like xml 但比較簡單
	configmap := utils.GetYamlConfig(configpath)
	//host like as "localhost:1024"
	host := utils.GetElement("host", configmap)
	//心跳包間隔時間
	timeinterval, err := strconv.Atoi(utils.GetElement("beatinginterval", configmap))
	utils.CheckError(err)

	//開始監聽  tcp
	netListen, err := net.Listen("tcp", host)
	utils.CheckError(err)
	//最後需進行  close
	defer netListen.Close()
	utils.Log("Waiting for clients")

	for {
		//while 回圈，監聽連線資料
		conn, err := netListen.Accept()
		if err != nil {
			continue //發生錯誤，跳開再監聽
		}

		utils.Log(conn.RemoteAddr().String(), " tcp connect success")
		// 用 goroutine 處理連入資料，並設定心跳包
		go handleConnection(conn, timeinterval)
	}
}

//handle the connection
func handleConnection(conn net.Conn, timeout int) {

	//宣告一個解封包的 byte array
	tmpBuffer := make([]byte, 0)
	//宣告一個接封包資料的 byte array
	buffer := make([]byte, 1024)
	//心跳包用的channel資料，判別有無資料進入
	messnager := make(chan byte)
	for {
		//接收資料
		n, err := conn.Read(buffer)
		if err != nil {
			utils.Log(conn.RemoteAddr().String(), " connection error: ", err)
			return
		}

		//解封包, 當前 tmpBuffer 未清，則  append 資料再一并處理
		tmpBuffer = utils.Depack(append(tmpBuffer, buffer[:n]...))
		utils.Log("receive data string:", string(tmpBuffer))
		utils.TaskDeliver(tmpBuffer, conn)
		//start heartbeating
		go utils.HeartBeating(conn, messnager, timeout)
		//check if get message from client
		go utils.GravelChannel(tmpBuffer, messnager)

	}
	defer conn.Close()

}
