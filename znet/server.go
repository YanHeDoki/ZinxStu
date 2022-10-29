package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

type Server struct {
	//服务器名称
	Name string
	//IP版本 IPv4 or other
	IPVersion string
	//服务器ip
	IP string
	//服务器监听的端口
	Port int

	//当前的对象添加一个router server注册的链接对应的业务
	//当前Server的消息管理模块，用来绑定MsgId和对应的router
	MsgHandler ziface.IMsgHandle
}

func (s *Server) Start() {

	//日志，以后应该用日志来处理
	fmt.Printf("[start]Server Listenner at IP %s,Port %d ,is staring", s.IP, s.Port)

	//开启工作线程池
	s.MsgHandler.StartWorkerPool()
	//由server方法来阻塞所以异步处理

	go func() {
		//获取一个Tcp的Addr地址
		resolveIPAddr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("Start ServerErr err:", err)
			panic(err)
		}
		//监听服务器的地址
		listen, err := net.ListenTCP(s.IPVersion, resolveIPAddr)
		if err != nil {
			fmt.Println("ListenIPErr err:", err)
			panic(err)
		}

		fmt.Println("Start Zinx Server success", s.Name, "success Listening...")
		var cid uint32
		cid = 0
		//阻塞的等待客户端的连接 处理客户端的链接业务（读写）
		for {
			conn, err := listen.AcceptTCP()
			if err != nil {
				fmt.Println("AcceptTCP err:", err)
				continue
			}
			fmt.Println("test@", cid)
			//使用新的connection模块
			newConnection := NewConnection(conn, cid, s.MsgHandler)
			cid++
			go newConnection.Start()
		}
	}()

}

func (s *Server) Stop() {

	//断开服务器，将一些服务器的资源链接释放

}

func (s *Server) Server() {

	//启动服务器
	s.Start()

	//TODO 留空位可以给以后操作空间
	//阻塞 否则主Go退出， listenner的go将会退出
	select {}
}

func (s *Server) AddRouter(msgid uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgid, router)
}

//初始化server服务器方法
func NewServer() ziface.IServer {
	return &Server{ //报错不能返回这个类型
		Name:       utils.GlobalConfig.Name,
		IPVersion:  "tcp4",
		Port:       utils.GlobalConfig.TcpPort,
		MsgHandler: NewMsgHandle(),
		IP:         utils.GlobalConfig.Host,
	}
}
