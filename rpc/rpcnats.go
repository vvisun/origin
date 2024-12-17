package rpc

import "sync/atomic"

type RpcNats struct {
	NatsServer
	NatsClient
}

func (rn *RpcNats) Start() error{
	err := rn.NatsServer.Start()
	if err != nil {
		return err
	}

	return rn.NatsClient.Start(rn.NatsServer.natsConn)
}

func (rn *RpcNats) Init(natsUrl string, noRandomize bool, nodeId string,compressBytesLen int,rpcHandleFinder RpcHandleFinder,notifyEventFun NotifyEventFun){
	rn.NatsClient.localNodeId = nodeId
	rn.NatsServer.initServer(natsUrl,noRandomize, nodeId,compressBytesLen,rpcHandleFinder,notifyEventFun)
	rn.NatsServer.iServer = rn
}

func  (rn *RpcNats) NewNatsClient(targetNodeId string,localNodeId string,callSet *CallSet,notifyEventFun NotifyEventFun) *Client{
	var client Client

	client.clientId = atomic.AddUint32(&clientSeq, 1)
	client.targetNodeId = targetNodeId
	natsClient := &rn.NatsClient
	natsClient.localNodeId = localNodeId
	natsClient.client = &client
	natsClient.notifyEventFun = notifyEventFun

	client.IRealClient = natsClient
	client.CallSet = callSet

	return &client
}