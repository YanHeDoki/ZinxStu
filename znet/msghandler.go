package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter //路由模块
	WorkerPoolSize uint32                    //业务工作Worker池的数量
	TaskQueue      []chan ziface.IRequest    //Worker负责取任务的消息队列
}

//构造方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalConfig.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalConfig.WorkerPoolSize), //注意一个消息队列对应一个worker池子
	}
}

func (m *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handle, ok := m.Apis[request.GetMsgId()]
	if !ok {
		fmt.Println("not find Router In Apis")
		return
	}
	handle.PreHandle(request)
	handle.Handle(request)
	handle.AfterHandle(request)

}

func (m *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	//1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := m.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	//2 添加msg与api的绑定关系
	m.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

func (m *MsgHandle) StartWorkerPool() {

	//根据配置的workerpool的size来分别开启worker 每个worker用一个go承载
	for i := uint32(0); i < m.WorkerPoolSize; i++ {
		//一个worker被启动
		//1.当前的worker对应的channel消息队列 开辟对应的空间 0号worker对应0号channel
		//用MaxWorkerTaskLen限制一个管道最多接受多少条消息
		m.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalConfig.MaxWorkerTaskLen)
		go m.startOneWorker(i)
	}

}

func (m *MsgHandle) startOneWorker(workerId uint32) {
	fmt.Println("start worker for ", workerId)
	//不断的阻塞去等代消息

	for {
		select {
		//根据id去结构体中取到对应的消息队列来消费，如果管道中有消息的话
		case req := <-m.TaskQueue[workerId]:
			m.DoMsgHandler(req)
		}
	}

}

func (m *MsgHandle) SendMsgToTaskQueue(req ziface.IRequest) {
	//将消息平均的分配给woroker
	//根据客户端建立的连接id来判断
	workerId := req.GetConnection().GetConnID() % m.WorkerPoolSize
	fmt.Println("add connIP=", req.GetConnection().GetConnID(), "req msgid=", req.GetMsgId(),
		"to worker", workerId)
	//将消息发送给消息队列
	m.TaskQueue[workerId] <- req
}
