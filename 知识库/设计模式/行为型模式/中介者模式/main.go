// 中介者模式可运行示范（Go）—— 聊天室
//
// 核心一句话：让对象之间不直接认识，把「多对多」的相互引用收拢到
// 一个中介者身上——成员只跟中介者说话，中介者负责转发给其他人。
//
// 本目录运行：
//
//	go run .
package main

import "fmt"

// =============================================================================
// 一、角色 1：Mediator（中介者）——「消息如何流转」的统一约定
// =============================================================================

// Mediator 中介者接口：注册成员、群发、私聊。
type Mediator interface {
	Register(u *User)
	Broadcast(from *User, msg string)       // 群发（除发送者外）
	Private(from *User, toName, msg string) // 私聊
}

// =============================================================================
// 二、角色 2：Colleague（同事）——「只认识中介者」的成员
// =============================================================================
//
// User 不保存任何其他 User 引用；发消息一律交给 mediator 转发。

// User 聊天室成员：只持有自己的名字和中介者。
type User struct {
	name     string
	mediator Mediator
}

// NewUser 创建成员并注册到中介者。
func NewUser(name string, m Mediator) *User {
	u := &User{name: name, mediator: m}
	m.Register(u)
	return u
}

// Send 群发消息：转发给中介者。
func (u *User) Send(msg string) { u.mediator.Broadcast(u, msg) }

// SendTo 私聊：指定接收者的名字（成员之间依然不直接认识）。
func (u *User) SendTo(to, msg string) { u.mediator.Private(u, to, msg) }

// Receive 收到消息后的行为（由中介者调用）。
func (u *User) Receive(from, msg string) {
	fmt.Printf("  [%s] 收到来自【%s】的消息：%s\n", u.name, from, msg)
}

// =============================================================================
// 三、角色 3：ConcreteMediator（具体中介者）——「聊天室本体」
// =============================================================================
//
// ChatRoom 持有全部成员；真正的派发逻辑集中在它这里。

// ChatRoom 聊天室：维护成员表，负责消息转发。
type ChatRoom struct {
	users map[string]*User
}

// NewChatRoom 创建聊天室。
func NewChatRoom() *ChatRoom {
	return &ChatRoom{users: make(map[string]*User)}
}

// Register 注册成员（重名则后者覆盖）。
func (c *ChatRoom) Register(u *User) {
	c.users[u.name] = u
	fmt.Printf("[聊天室] %s 加入\n", u.name)
}

// Broadcast 群发：转发给除发送者外的所有成员。
func (c *ChatRoom) Broadcast(from *User, msg string) {
	for _, u := range c.users {
		if u != from {
			u.Receive(from.name, msg)
		}
	}
}

// Private 私聊：只发给指定成员。
func (c *ChatRoom) Private(from *User, toName, msg string) {
	if to, ok := c.users[toName]; ok {
		to.Receive(from.name, msg)
	} else {
		fmt.Printf("  [聊天室] 找不到成员 %q，消息未送达\n", toName)
	}
}

// =============================================================================
// 四、main：先群聊、再私聊，成员彼此不直接引用
// =============================================================================

func main() {
	room := NewChatRoom()

	alice := NewUser("Alice", room)
	bob := NewUser("Bob", room)
	NewUser("Carol", room) // Carol 加入但不必持有引用；注册后群发/私聊都能收到

	fmt.Println("--- 群发：Alice 广播 ---")
	alice.Send("大家好，我是 Alice！")

	fmt.Println("--- 私聊：Bob → Carol ---")
	bob.SendTo("Carol", "下班一起吃饭？")

	fmt.Println("--- Bob 群发 ---")
	bob.Send("收到请回复")

	fmt.Println()
	fmt.Println("========== 读懂输出后你会发现 ==========")
	fmt.Println("1. Alice / Bob / Carol 彼此不认识，只认识 room 这个中介者")
	fmt.Println("2. 消息怎么转发（群发/私聊）集中在 ChatRoom，成员代码零改动")
	fmt.Println("3. 新增成员 = 新建一个 User 并注册，老成员不用改")
	fmt.Println("4. 多对多网状引用 → 变成以中介者为中心的一对多")
}
