package e7s

import (
	"sync"
)

// ClientManager 连接管理
type clientManager struct {
	clients     map[*client]bool       // 全部的连接
	clientsLock sync.RWMutex           // 读写锁
	users       map[string]*client     // 登录的用户
	userLock    sync.RWMutex           // 读写锁
	register    chan *client           // 连接连接处理
	login       chan *Login            // 用户登录处理
	loginOut    chan *client           // 用户退出处理
	unregister  chan *client           // 断开连接处理程序
	uidBan      chan int               // 断开UID连接处理程序
	broadcast   chan *broadcastMessage // 广播 向全部成员发送数据
}

func newClientManager() *clientManager {
	return &clientManager{
		clients:    make(map[*client]bool),
		users:      make(map[string]*client),
		register:   make(chan *client, 1000),
		login:      make(chan *Login, 1000),
		loginOut:   make(chan *client, 1000),
		unregister: make(chan *client, 1000),
		uidBan:     make(chan int, 1000),
		broadcast:  make(chan *broadcastMessage, 1000),
	}
}

type Login struct {
	uid string
	c   *client
}

type broadcastMessage struct {
	From    string
	Message []byte
}

/**************************  manager  ***************************************/

func (manager *clientManager) inClient(client *client) (ok bool) {
	manager.clientsLock.RLock()
	defer manager.clientsLock.RUnlock()
	_, ok = manager.clients[client]
	if !ok {
		manager.clients[client] = true
	}
	return
}

// GetClients
func (manager *clientManager) getClients() (clients map[*client]bool) {

	clients = make(map[*client]bool)

	manager.clientsRange(func(client *client, value bool) (result bool) {
		clients[client] = value

		return true
	})

	return
}

// 遍历
func (manager *clientManager) clientsRange(f func(client *client, value bool) (result bool)) {

	manager.clientsLock.RLock()
	defer manager.clientsLock.RUnlock()

	for key, value := range manager.clients {
		result := f(key, value)
		if result == false {
			return
		}
	}

	return
}

// GetClientsLen
func (manager *clientManager) getClientsLen() (clientsLen int) {

	clientsLen = len(manager.clients)

	return
}

// AddClients 添加客户端
func (manager *clientManager) addClients(client *client) {
	manager.clientsLock.Lock()
	defer manager.clientsLock.Unlock()

	_, ok := manager.clients[client]
	if !ok {
		manager.clients[client] = true
	}
	return
}

// DelClients 删除客户端
func (manager *clientManager) delClients(client *client) {
	manager.clientsLock.Lock()
	defer manager.clientsLock.Unlock()
	if _, ok := manager.clients[client]; ok {
		delete(manager.clients, client)
	}
}

// GetUserClient 获取用户的连接
func (manager *clientManager) getUserClient(uid string) (client *client) {

	manager.userLock.RLock()
	defer manager.userLock.RUnlock()

	if value, ok := manager.users[uid]; ok {
		client = value
	} else {
		client = nil
	}
	return
}

// GetUsersLen GetClientsLen
func (manager *clientManager) getUOnlineLen() (userLen int) {
	userLen = len(manager.users)

	return
}

// AddUsers 添加用户
func (manager *clientManager) addUsers(uid string, client *client) {
	manager.userLock.Lock()
	defer manager.userLock.Unlock()
	if clients, ok := manager.users[uid]; ok {
		clients.loginTime = 0
		clients.userId = 0
		manager.delUsers(uid)
		manager.users[uid] = client
	} else {
		manager.users[uid] = client
	}
}

// DelUsers 删除用户
func (manager *clientManager) delUsers(uid string) {
	manager.userLock.Lock()
	defer manager.userLock.Unlock()
	if _, ok := manager.users[uid]; ok {
		delete(manager.users, uid)
	}
}

// GetUserClients 获取uid 连接
func (manager *clientManager) getUserClients() (clients []*client) {

	clients = make([]*client, 0)
	manager.userLock.RLock()
	defer manager.userLock.RUnlock()
	for _, v := range manager.users {
		clients = append(clients, v)
	}
	return
}

// EventRegister 用户建立连接事件
func (manager *clientManager) eventRegister(client *client) {
	manager.addClients(client)
}

// EventUnregister 用户断开连接
func (manager *clientManager) eventUnregister(client *client) {
	manager.delClients(client)
	if client.userId != 0 {
		manager.delUsers(client.userId)
	}
}

// EventULoginOut LoginOut 退出
func (manager *clientManager) eventULoginOut(client *client) {
	manager.delUsers(client.userId)
	client.loginTime = 0
	client.userId = 0
}

// EventUidBan  封号
func (manager *clientManager) eventUidBan(uid int) {
	uidClient := manager.getUserClient(uid)
	if uidClient != nil {
		manager.loginOut <- uidClient
	}
	UidClient := manager.getUserClient(uid)
	if UidClient != nil {
		manager.unregister <- UidClient
		UidClient.socket.Close()
	}

}

// Start 管道处理程序
func (manager *clientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			// 建立连接事件
			manager.eventRegister(conn)
		case conn := <-manager.unregister:
			// 断开连接事件
			manager.eventUnregister(conn)
		case user := <-manager.login:
			//登陆事件
			manager.addUsers(user.uid, user.c)
		case uidClient := <-manager.loginOut:
			//退出事件
			manager.eventULoginOut(uidClient)
		case uid := <-manager.uidBan:
			//退出事件
			manager.eventUidBan(uid)
		case message := <-manager.broadcast:
			// 广播事件
			clients := manager.getClients()
			for conn := range clients {
				if message.From != "" && message.From == conn.addr {
					continue
				}
				select {
				case conn.send <- message.Message:
				default:
					close(conn.send)
				}
			}
		}
	}
}
