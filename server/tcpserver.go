package main

import (
   // "bufio"
   "fmt"
   "net"
   "os"
   "strings"
   "time"
)

func main() {
   arguments := os.Args
   if len(arguments) == 1 {
      fmt.Println("Please provide port number")
      return
   }

   PORT := ":" + arguments[1]
   l, err := net.Listen("tcp", PORT)
   if err != nil {
      fmt.Println(err)
      return
   }
   defer l.Close()
   fmt.Print("TCP server start and listening.\n")

   for {
      conn, err := l.Accept()
      if err != nil {
         fmt.Println(err)
      }
      go handleConnection(conn)
   }
}

func handleConnection(conn net.Conn) {
   remoteAddr := conn.RemoteAddr().String()
   fmt.Println("Client connected from: " + remoteAddr)
   buf := make([]byte, 1024)

   for {
      reqLen, err := conn.Read(buf)
      if err != nil {
         if err.Error() == "EOF" {
            fmt.Println("Disconned from: ", remoteAddr)
            break
         } else {
            fmt.Println(err.Error())
            break
         }
      }
      message := strings.TrimSpace(string(buf[:reqLen]))
      if message == "STOP" {
         fmt.Printf("%s exiting TCP server!\n", remoteAddr)
         break
      }
      conn.Write([]byte("https://www.sinica.edu.tw\n"))
      t := time.Now()
      myTime := t.Format(time.RFC3339)
      fmt.Printf("%v :%s\n", myTime, message)
   }
   conn.Close()
}
