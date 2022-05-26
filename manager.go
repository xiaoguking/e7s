package e7s

import (
	"fmt"
	"sync"
)

// 连接管理
type ClientManager struct {
	Clients     map[*Client]bool     // 全部的连接
	ClientsLock sync.RWMutex         // 读写锁
	Users       map[string][]*Client // 登录的用户
	UserLock    sync.RWMutex         // 读写锁
	Register    chan *Client         // 连接连接处理
	Login       chan *Login          // 用户登录处理
	LoginOut    chan string          // 用户退出处理
	Unregister  chan *Client         // 断开连接处理程序
	Broadcast   chan []byte          // 广播 向全部成员发送数据
}

func NewClientManager() (clientManager *ClientManager) {
	clientManager = &ClientManager{
		Clients:    make(map[*Client]bool),
		Users:      make(map[string][]*Client),
		Register:   make(chan *Client, 1000),
		Login:      make(chan *Login, 1000),
		LoginOut:   make(chan string, 1000),
		Unregister: make(chan *Client, 1000),
		Broadcast:  make(chan []byte, 1000),
	}

	return
}

type Login struct {
	Uid string
	C   *Client
}

/**************************  manager  ***************************************/

func (manager *ClientManager) InClient(client *Client) (ok bool) {
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	// 连接存在，在添加
	_, ok = manager.Clients[client]

	return
}

// GetClients
func (manager *ClientManager) GetClients() (clients map[*Client]bool) {

	clients = make(map[*Client]bool)

	manager.ClientsRange(func(client *Client, value bool) (result bool) {
		clients[client] = value

		return true
	})

	return
}

// 遍历
func (manager *ClientManager) ClientsRange(f func(client *Client, value bool) (result bool)) {

	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	for key, value := range manager.Clients {
		result := f(key, value)
		if result == false {
			return
		}
	}

	return
}

// GetClientsLen
func (manager *ClientManager) GetClientsLen() (clientsLen int) {

	clientsLen = len(manager.Clients)

	return
}

// 添加客户端
func (manager *ClientManager) AddClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	manager.Clients[client] = true
}

// 删除客户端
func (manager *ClientManager) DelClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()
	if _, ok := manager.Clients[client]; ok {
		delete(manager.Clients, client)
	}
}

// 获取用户的连接
func (manager *ClientManager) GetUserClient(userId string) (client []*Client) {

	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()

	if value, ok := manager.Users[userId]; ok {
		client = value
	}

	return
}

// GetClientsLen
func (manager *ClientManager) GetUsersLen() (userLen int) {
	userLen = len(manager.Users)

	return
}

// 添加用户
func (manager *ClientManager) AddUsers(key string, client *Client) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()
	if clients, ok := manager.Users[key]; ok {
		manager.Users[key] = append(clients, client)
	} else {
		value := make([]*Client, 0)
		manager.Users[key] = append(value, client)
	}
}

// 删除用户
func (manager *ClientManager) DelUsers(key string) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()

	fmt.Println("DelUsers 4")
	if _, ok := manager.Users[key]; ok {
		delete(manager.Users, key)
	}
}

// 获取用户的key
func (manager *ClientManager) GetUserKeys() (userKeys []string) {

	userKeys = make([]string, 0)
	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()
	for key := range manager.Users {
		userKeys = append(userKeys, key)
	}
	return
}

// 获取用户的key
func (manager *ClientManager) GetUserClients() (clients []*Client) {

	clients = make([]*Client, 0)
	manager.UserLock.RLock()
	defer manager.UserLock.RUnlock()
	for _, v := range manager.Users {
		for _, vs := range v {
			clients = append(clients, vs)
		}
	}
	return
}

// 用户建立连接事件
func (manager *ClientManager) EventRegister(client *Client) {
	manager.AddClients(client)
}

// 用户断开连接
func (manager *ClientManager) EventUnregister(client *Client) {
	manager.DelClients(client)
	if client.UserId != "" {
		manager.UserLock.RLock()
		defer manager.UserLock.RUnlock()
		if userClient, ok := manager.Users[client.UserId]; ok {
			userClientLen := len(userClient)
			if userClientLen <= 0 {
				delete(manager.Users, client.UserId)
			}
			if userClientLen == 1 && userClient[0] == client {
				delete(manager.Users, client.UserId)
			} else {
				var newUserClinet []*Client
				for i := range userClient {
					if userClient[i] == client {
						newUserClinet = append(userClient[:i], userClient[i+1:]...)
						manager.Users[client.UserId] = newUserClinet
					}
				}
			}
		}
	}
}

//LoginOut
func (manager *ClientManager) EventULoginOut(uid string) {
	manager.UserLock.Lock()
	defer manager.UserLock.Unlock()
	if v, ok := manager.Users[uid]; ok {
		for _, cs := range v {
			cs.LoginTime = 0
			cs.UserId = ""
		}
		delete(manager.Users, uid)
	}
}

//sendUid
func (manager *ClientManager) SendToUid(uid string, msg []byte) {
	client := manager.GetUserClient(uid)
	for _, v := range client {
		v.Send <- msg
	}
}

//uids
func (manager *ClientManager) SendToUids(uid []string, msg []byte) {
	for _, v := range uid {
		client := manager.GetUserClient(v)
		for _, vs := range client {
			vs.Send <- msg
		}
	}
}

// 向全部成员(除了自己)发送数据
func (manager *ClientManager) SendOther(message []byte, ignore *Client) {

	clients := manager.GetUserClients()
	for _, conn := range clients {
		if conn != ignore {
			conn.Send <- message
		}
	}
}

//发送广播
func (manager *ClientManager) SendAll(message []byte) {
	manager.Broadcast <- message
}

// 管道处理程序
func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.Register:
			// 建立连接事件
			manager.EventRegister(conn)
		case conn := <-manager.Unregister:
			// 断开连接事件
			manager.EventUnregister(conn)
		case user := <-manager.Login:
			//登陆事件
			manager.AddUsers(user.Uid, user.C)
		case uid := <-manager.LoginOut:
			//退出事件
			manager.EventULoginOut(uid)
		case message := <-manager.Broadcast:
			// 广播事件
			clients := manager.GetClients()
			for conn := range clients {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
				}
			}
		}
	}
}
