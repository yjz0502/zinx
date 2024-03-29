package znet

import (
	"fmt"
	"net"
	"zinx/utils"
	"zinx/ziface"
)

// IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器的名称
	Name string
	//服务器绑定的ip版本
	IPVersion string
	//服务器监听的IP
	IP string
	//服务器监听的端口
	Port int
	//当前的Server的消息管理模块，用来绑定MsgID和对应的业务处理的API关系
	MsgHandle ziface.IMsgHandle
	//当前的Server的链接管理模块，用来绑定ConnID和对应的Conn
	ConnManager ziface.IConnManager
	//该Server创建链接之后自动调用Hook函数-OnConnStart
	OnConnStart func(conn ziface.IConnection)
	//该Server创建链接之后自动调用Hook函数-OnConnStop
	OnConnStop func(conn ziface.IConnection)
}

func (s *Server) Start() {
	fmt.Printf("[Zinx] Server Name : %s, listenner at IP : %s, Port : %d is starting\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] Version : %s, MaxConn : %d, MaxPackeetSize : %d\n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPackageSize)

	go func() {
		//开启一个工作池
		s.MsgHandle.StartWorkerPool()

		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addt error ：", err)
			return
		}
		//2 监听服务器的地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}
		fmt.Println("start Zinx server succ, ", s.Name, " succ, Linstenning...")
		var cid uint32 = 0

		//3 阻塞的等待客服端链接，处理客户端连接业务(读写)
		for {
			//如果有客户端连接过来，阻塞会返回
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			//设置最大链接个数的判断，如果超过最大链接，那么则关闭此新的链接
			if s.ConnManager.Len() >= utils.GlobalObject.MaxConn {
				//TODO 给客户端相应一个超出最大链接的错误包
				fmt.Println("Too Many Connection MaxConn = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}
			//将处理新连接的业务方法和conn进行绑定，得到我们的链接模块
			delConn := NewConnection(s, conn, cid, s.MsgHandle)
			cid++
			//启动当前的链接业务方法
			go delConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	//TODO 将一些服务器的资源、状态或者一下已经开辟的链接信息 进行停止或者回收
	fmt.Println("[STOP] Zinx server name ", s.Name)
	s.ConnManager.ClearConn()
}

func (s *Server) Serve() {
	//启动server的服务功能
	s.Start()

	//TODO 做一些启动服务器之后的额外业务

	//阻塞状态
	select {}
}

func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.MsgHandle.AddRouter(msgId, router)
	println("Add Router Succ!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnManager
}

// 初始化Server模块的方法
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:        utils.GlobalObject.Name,
		IPVersion:   "tcp4",
		IP:          utils.GlobalObject.Host,
		Port:        utils.GlobalObject.TcpPort,
		MsgHandle:   NewMsgHandle(),
		ConnManager: NewConnManager(),
	}
	return s
}

// 注册OnConnStart钩子函数的方法
func (s *Server) SetOnConnStart(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 调用OnConnStart钩子函数的方法
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("————>Call OnConnStart()...")
		s.OnConnStart(conn)
	}
}

// 注册OnConnStop钩子函数的方法
func (s *Server) SetOnConnStop(hookFunc func(connection ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStop钩子函数的方法
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("————>Call OnConnStop()...")
		s.OnConnStop(conn)
	}
}
