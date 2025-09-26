package main

import (
	"HTTP_FROM_TCP/internal/request"
	"fmt"
	"log"
	"net"
)

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	var out = make(chan string)

// 	go func() {
// 		defer close(out)
// 		defer f.Close()
// 		current_line := ""
// 		for {
// 			data := make([]byte, 8)
// 			count, err := f.Read(data)
// 			if err != nil {
// 				break
// 			}
// 			data = data[:count]
// 			for {
// 				i := bytes.IndexByte(data, '\n')
// 				if i == -1 {
// 					break
// 				}

// 				current_line += string(data[:i])
// 				out <- current_line
// 				data = data[i+1:]
// 				current_line = ""

// 			}
// 			current_line += string(data)
// 		}

// 		if len(current_line) != 0 {
// 			out <- current_line
// 		}
// 	}()

// 	return out
// }

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	fmt.Println("Listening on :42069")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		r, err := request.RequestFromReader(conn)

		if err != nil {
			log.Fatal("error", "error", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		r.Headers.ForEach(func(k, v string) {
			fmt.Printf("%s: %s\n", k, v)
		})

		fmt.Println("Body:")
		fmt.Printf("%s\n", string(r.Body))
	}
}
