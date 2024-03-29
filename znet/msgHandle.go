package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

/*
消息处理模块的实现
*/
type MsgHandle struct {
	//存放每个MsgID 所对应的处理方法
	Apis map[uint32]ziface.IRouter
	//负责Worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//业务工作Worker池的worker数量
	WorkPoolSize uint32
}

// 初始化/创建MsgHandle方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:         make(map[uint32]ziface.IRouter),
		WorkPoolSize: utils.GlobalObject.WorkerPooleSize, //从全局配置中获取
		TaskQueue:    make([]chan ziface.IRequest, utils.GlobalObject.WorkerPooleSize),
	}
}

func (mh *MsgHandle) DoMsgHandle(request ziface.IRequest) {
	router, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgID = ", request.GetMsgID(), " id NOT FOUNT! Need Register")
		return
	}
	router.PreHandle(request)
	router.Handle(request)
	router.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的Api方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeat api, msgId = " + strconv.Itoa(int(msgId)))
	}
	//2添加msg与API的绑定是关系
	mh.Apis[msgId] = router
	fmt.Println("Add api MsgID = ", msgId, " succ! ")
}

// 启动一个Worker工作池(开启工作池的动作只能发发生一次，一个Zinx框架只能有一个Worker工作池)
func (mh *MsgHandle) StartWorkerPool() {
	//根据WorkerPooleSize分包开启Worker，每个Worker用一个go来承载
	for i := 0; i < int(mh.WorkPoolSize); i++ {
		//一个Worker启动
		//1当前的Worker对应的channel消息队列开辟空间,第0个worker就用低0个channel
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//2启动当前的Worker，阻塞等待消息从channel传递进来
		go mh.StartOneWorker(i)
	}
}

// 启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerId int) {
	fmt.Println("Worker ID =", workerId, " is started")
	//不断的阻塞等待对应消息队列的消息
	for {
		select {
		//如果有消息过来，出列的就是一个客户端的Request,执行当前Request所绑定的业务
		case request := <-mh.TaskQueue[workerId]:
			mh.DoMsgHandle(request)
		}
	}
}

// 将消息交给TaskQueue,由Worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//1将消息分配给不同的Worker
	//根据客户端建议的ConnID来进行分配
	workerID := request.GetConnection().GetConnID() % mh.WorkPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(),
		" request MsgID = ", request.GetMsgID(),
		" to WorkerID = ", workerID)
	//2将消息发送给对应的Worker的TaskQueue即可
	mh.TaskQueue[workerID] <- request
}
