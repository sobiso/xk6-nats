package nats

import (
	"encoding/json"
	"fmt"
	natsio "github.com/nats-io/nats.go"
	"go.k6.io/k6/js/modules"
	"time"
)

func init() {
	modules.Register("k6/x/nats", new(RootModule))
}

// RootModule is the global module object type. It is instantiated once per test
// run and will be used to create k6/x/nats module instances for each VU.
type RootModule struct{}

// ModuleInstance represents an instance of the module for every VU.
type Nats struct {
	//conn    *natsio.Conn
	vu modules.VU
	//exports map[string]interface{}
	sub     *natsio.Subscription
	subSync *natsio.Subscription
}

// Ensure the interfaces are implemented correctly.
var (
	_ modules.Module   = &RootModule{}
	_ modules.Instance = &Nats{}
)

// NewModuleInstance implements the modules.Module interface and returns
// a new instance for each VU.
func (r *RootModule) NewModuleInstance(vu modules.VU) modules.Instance {
	return &Nats{vu: vu}
	//mi := &Nats{
	//	vu: vu,
	//	exports: make(map[string]interface{}),
	//}

	//mi.exports["Nats"] = mi.client

	//return mi
}

// Exports implements the modules.Instance interface and returns the exports
// of the JS module.
func (n *Nats) Exports() modules.Exports {
	return modules.Exports{
		Default: n,
	}
	//return modules.Exports{
	//	Named: mi.exports,
	//}
}

type MessageTopic struct {
	Header map[string]string
	Body   []byte
}

func (n *Nats) Open(host string) (*natsio.Conn, error) {
	natsOptions := natsio.GetDefaultOptions()
	conn, err := natsOptions.Connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

//func (n *Nats) client(c goja.ConstructorCall) *goja.Object {
//	rt := n.vu.Runtime()
//	fmt.Println("New client init")
//
//	var cfg Configuration
//	err := rt.ExportTo(c.Argument(0), &cfg)
//	if err != nil {
//		common.Throw(rt, fmt.Errorf("Nats constructor expect Configuration as it's argument: %w", err))
//	}
//
//	natsOptions := natsio.GetDefaultOptions()
//	natsOptions.Servers = cfg.Servers
//	if cfg.Unsafe {
//		natsOptions.TLSConfig = &tls.Config{
//			InsecureSkipVerify: true,
//		}
//	}
//	if cfg.Token != "" {
//		natsOptions.Token = cfg.Token
//	}
//
//	conn, err := natsOptions.Connect()
//	if err != nil {
//		common.Throw(rt, err)
//	}
//
//	return rt.ToValue(&Nats{
//		vu:   n.vu,
//		conn: conn,
//	}).ToObject(rt)
//}

func (*Nats) Close(conn *natsio.Conn) {
	if conn != nil {
		conn.Close()
	}
}

func (*Nats) Publish(conn *natsio.Conn, topic, message string) error {
	if conn == nil {
		return fmt.Errorf("the connection is not valid")
	}

	return conn.Publish(topic, []byte(message))
}

func (n *Nats) Subscribe(conn *natsio.Conn, topic string, handler MessageHandler) (*natsio.Subscription, error) {
	if conn == nil {
		return nil, fmt.Errorf("the connection is not valid")
	}
	sub, err := conn.Subscribe(topic, func(msg *natsio.Msg) {

		err := handler(msg.Data)
		if err != nil {
			fmt.Printf("handler error, %s", err.Error())
		}
	})
	//sub, err := conn.Subscribe(topic, func(msg *natsio.Msg) {
	//	fmt.Printf("msg: %+v", msg)
	//
	//	//message := Message{
	//	//	Data:      string(msg.Data),
	//	//	DataBytes: msg.Data,
	//	//	Topic:     msg.Subject,
	//	//}
	//	//
	//	err := handler(string(msg.Data))
	//	if err != nil {
	//		return fmt.Errorf("%v", err)
	//	}
	//
	//})
	if err != nil {
		return nil, err
	}
	n.sub = sub
	return sub, nil
}

func (n *Nats) SubscribeSync(conn *natsio.Conn, topic string) ([]byte, error) {
	if conn == nil {
		return nil, fmt.Errorf("the connection is not valid")
	}

	sub, err := conn.SubscribeSync(topic)
	if err != nil {
		return nil, err
	}
	m, err := sub.NextMsg(1 * time.Second)
	if err != nil {
		return nil, nil
	}

	var msg MessageTopic
	err = json.Unmarshal(m.Data, &msg)
	if err != nil {
		return nil, err
	}

	return msg.Body, nil
}

func (n *Nats) Unsubscribe(sub *natsio.Subscription) error {
	fmt.Println("Unsubscribe")
	return sub.Unsubscribe()
}

//	func (n *Nats) Subscribe(topic string, handler MessageHandler) error {
//		if n.conn == nil {
//			return fmt.Errorf("the connection is not valid")
//		}
//
//		sub, err := n.conn.Subscribe(topic, func(msg *natsio.Msg) {
//			message := Message{
//				Data:  string(msg.Data),
//				Topic: msg.Subject,
//			}
//			fmt.Printf("message: %s", message)
//
//			handler(message)
//		})
//
//		n.sub = sub
//		return err
//	}
//
//	func (n *Nats) Unsubscribe() {
//		fmt.Println("Unsubscribe")
//		_ = n.sub.Unsubscribe()
//	}
//
// // Connects to JetStream and creates a new stream or updates it if exists already
//
//	func (n *Nats) JetStreamSetup(streamConfig *natsio.StreamConfig) error {
//		if n.conn == nil {
//			return fmt.Errorf("the connection is not valid")
//		}
//
//		js, err := n.conn.JetStream()
//		if err != nil {
//			return fmt.Errorf("cannot accquire jetstream context %w", err)
//		}
//
//		stream, _ := js.StreamInfo(streamConfig.Name)
//		if stream == nil {
//			_, err = js.AddStream(streamConfig)
//		} else {
//			_, err = js.UpdateStream(streamConfig)
//		}
//
//		return err
//	}
//
//	func (n *Nats) JetStreamDelete(name string) error {
//		if n.conn == nil {
//			return fmt.Errorf("the connection is not valid")
//		}
//
//		js, err := n.conn.JetStream()
//		if err != nil {
//			return fmt.Errorf("cannot accquire jetstream context %w", err)
//		}
//
//		js.DeleteStream(name)
//
//		return err
//	}
//
//	func (n *Nats) JetStreamPublish(topic string, message string) error {
//		if n.conn == nil {
//			return fmt.Errorf("the connection is not valid")
//		}
//
//		js, err := n.conn.JetStream()
//		if err != nil {
//			return fmt.Errorf("cannot accquire jetstream context %w", err)
//		}
//
//		_, err = js.Publish(topic, []byte(message))
//
//		return err
//	}
//
//	func (n *Nats) JetStreamSubscribe(topic string, handler MessageHandler) error {
//		if n.conn == nil {
//			return fmt.Errorf("the connection is not valid")
//		}
//
//		js, err := n.conn.JetStream()
//		if err != nil {
//			return fmt.Errorf("cannot accquire jetstream context %w", err)
//		}
//
//		sub, err := js.Subscribe(topic, func(msg *natsio.Msg) {
//			message := Message{
//				Data:  string(msg.Data),
//				Topic: msg.Subject,
//			}
//			handler(message)
//		})
//
//		defer func() {
//			if err := sub.Unsubscribe(); err != nil {
//				fmt.Errorf("Error unsubscribing")
//			}
//		}()
//
//		return err
//	}
//
//	func (n *Nats) Request(subject, data string) (Message, error) {
//		if n.conn == nil {
//			return Message{}, fmt.Errorf("the connection is not valid")
//		}
//
//		msg, err := n.conn.Request(subject, []byte(data), 5*time.Second)
//		if err != nil {
//			return Message{}, err
//		}
//
//		return Message{
//			Data:  string(msg.Data),
//			Topic: msg.Subject,
//		}, nil
//	}
//
//	type Configuration struct {
//		Servers []string
//		Unsafe  bool
//		Token   string
//	}
type Message struct {
	Data      string
	DataBytes []byte
	Topic     string
}

type MessageHandler func([]byte) error
