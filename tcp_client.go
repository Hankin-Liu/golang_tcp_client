package main

import (
	"fmt"
//	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"time"
    "runtime/debug"
	"bytes"
	"encoding/binary"
	"bufio"
)

type protocol struct {
	pkg_len uint16
	pkg_type uint8
}

func handle_signal() {
	chan_signal := make(chan os.Signal)
	signal.Notify(chan_signal)
	signal := <-chan_signal
	fmt.Println("Get signal:", signal)
	os.Exit(-1)
}

func send_debug_pkg(conn net.Conn) {
	content := []byte("debug") 
	hb := protocol {
		pkg_len:9,
		pkg_type:'+',
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, hb.pkg_len)
	if (err != nil) {
		fmt.Println("debug:write to buffer failed");
		os.Exit(-1)
	}
	err = binary.Write(buf, binary.BigEndian, hb.pkg_type)
	if (err != nil) {
		fmt.Println("debug:write to buffer failed");
		os.Exit(-1)
	}
	err = binary.Write(buf, binary.BigEndian, content)
	if (err != nil) {
		fmt.Println("debug:write to buffer failed");
		os.Exit(-1)
	}

	_, err = conn.Write(buf.Bytes())
	if (err != nil) {
		fmt.Println("debug:write to buffer failed");
		os.Exit(-1)
	}
}

func send_unsequenced_pkg(conn net.Conn) {
	content := []byte("unseq") 
	hb := protocol {
		pkg_len:9,
		pkg_type:'U',
	}
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, hb.pkg_len)
	if (err != nil) {
		fmt.Println("unseq:write to buffer failed");
		os.Exit(-1)
	}
	err = binary.Write(buf, binary.BigEndian, hb.pkg_type)
	if (err != nil) {
		fmt.Println("unseq:write to buffer failed");
		os.Exit(-1)
	}
	err = binary.Write(buf, binary.BigEndian, content)
	if (err != nil) {
		fmt.Println("unseq:write to buffer failed");
		os.Exit(-1)
	}

	_, err = conn.Write(buf.Bytes())
	if (err != nil) {
		fmt.Println("unseq:write to buffer failed");
		os.Exit(-1)
	}
}

func send_client_heartbeat(conn net.Conn) {
	ticker1 := time.NewTicker(100 * time.Millisecond)
	go func(t *time.Ticker) {
		hb := protocol {
			pkg_len:3,
			pkg_type:'R',
		}
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, hb.pkg_len)
		if (err != nil) {
			fmt.Println("write to buffer failed");
			os.Exit(-1)
		}
		err = binary.Write(buf, binary.BigEndian, hb.pkg_type)
		if (err != nil) {
			fmt.Println("write to buffer failed");
			os.Exit(-1)
		}
		// time is 0ms, send a client heartbeat
		_, err = conn.Write(buf.Bytes())
		if (err != nil) {
			fmt.Println("write to buffer failed");
			os.Exit(-1)
		}
		duration := 0
		send_hb_countdown := 500
		for {
			<-t.C
			duration += 100
			send_hb_countdown -= 100
			if (send_hb_countdown == 0) {
				// keepalive client heartbeat
				_, err = conn.Write(buf.Bytes())
				if (err != nil) {
					fmt.Println("write to buffer failed");
					os.Exit(-1)
				}
				send_hb_countdown = 500;
			}
			if (duration == 1300) {
				fmt.Println("reach 1300ms, send debug package");
				send_debug_pkg(conn)
			} else if (duration == 4600) {
				fmt.Println("reach 4000ms, send unsequenced package");
				send_unsequenced_pkg(conn)
			} else if (duration == 5400) {
				fmt.Println("reach 5400ms, send unsequenced package");
				send_unsequenced_pkg(conn)
			} else if (duration == 10000) {
				fmt.Println("reach 10000ms, normal exit");
				conn.Close()
				os.Exit(0)
			}
		}
	}(ticker1)

	reader := bufio.NewReaderSize(conn, 10240)
	for {
		byteSize := reader.Buffered()
		data := make([]byte, byteSize)
		_, err := reader.Read(data)
		if (err != nil) {
			os.Exit(-1)
		}

		//data, err := reader.Peek(3)
		//if (err != nil) {
		//	os.Exit(-1)
		//}
		//pkg_type := String(buf[2])
		//if () {
		//}
		//var ptestStruct *TestStructTobytes = *(**TestStructTobytes)(unsafe.Pointer(&data))
	}
}

func process_tcp_session(conn net.Conn) {
	send_client_heartbeat(conn)
}

func main() {
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("panic recover! p: %v", p)
			debug.PrintStack()
		}
	}()
	arg_cnt := len(os.Args);
	if (arg_cnt != 3) {
		fmt.Println("Usage: ./tcp_client ip port");
		os.Exit(-1)
	}

	ip_port := os.Args[1] + ":" + os.Args[2]
	tcpaddr, err := net.ResolveTCPAddr("", ip_port)
	if err != nil {
		fmt.Println("net ResolveTCPAddr error! ", err.Error())
		os.Exit(-1)
	}

	conn, err := net.DialTCP("tcp4", nil, tcpaddr)
	if err != nil {
		fmt.Println("net DialTcp Error!", err.Error())
		os.Exit(-1)
	}
	defer func() {
		if (conn != nil) {
			conn.Close()
		}
	}()

	go process_tcp_session(conn)

	handle_signal()

	//blen, err := conn.Write([]byte("HEAD / HTTP/1.0 \r\n\r\n"))
	//if err != nil {
	//	fmt.Println("err = ", err.Error())
	//}
	//fmt.Println("blen = ", blen)

	//result, err := ioutil.ReadAll(conn)
	//if err != nil {
	//	fmt.Println("ReadAll error: ", err.Error())
	//}
	//fmt.Println("result = ", string(result))

	//fmt.Println(conn.LocalAddr())
	//fmt.Println(conn.RemoteAddr())
}
