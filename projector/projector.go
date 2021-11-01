package main

import (
   // "bufio"
   "fmt"
   "net"
   "os"
   "time"
   "strings"
   "os/exec"
   "runtime"
   "github.com/joho/godotenv"
   "github.com/asccclass/staticfileserver/libs/ip" 
   "github.com/asccclass/staticfileserver/libs/line" 
)

// open browser
func openBrowser(url string) bool {
   var args []string
   switch runtime.GOOS {
   case "darwin":
      args = []string{"open"}
   case "windows":
      args = []string{"cmd", "/c", "start"}
   default:
      args = []string{"xdg-open"}
   }
   cmd := exec.Command(args[0], append(args[1:], url)...)
   return cmd.Start() == nil
}

func handleConnection(conn net.Conn) {
   remoteAddr := conn.RemoteAddr().String()
   fmt.Println("  Client connected from: " + remoteAddr)
   buf := make([]byte, 1024)

   for {
      reqLen, err := conn.Read(buf)
      if err != nil {
         if err.Error() == "EOF" {
            fmt.Println("  Disconned from: ", remoteAddr)
            break
         } else {
            fmt.Println(err.Error())
            break
         }
      }
      message := strings.TrimSpace(string(buf[:reqLen]))
      if message == "STOP" {
         fmt.Printf("  %s exiting TCP server!\n", remoteAddr)
         break
      } else if message[:3] == "URL" {
         openBrowser(message[4:])
      }
      // conn.Write([]byte("https://www.sinica.edu.tw\n"))
      t := time.Now()
      myTime := t.Format(time.RFC3339)
      fmt.Printf("  %v :%s\n", myTime, message)
   }
   conn.Close()
}

func main() {
   systemName := os.Getenv("SystemName")
   currentDir, _ := os.Getwd()
   if err := godotenv.Load(currentDir + "/envfile"); err != nil {
      fmt.Println("envfile is not exist.")
      return
   }
   arguments := os.Args
   if len(arguments) == 1 {
      fmt.Println("Please provide port number")
      return
   }

   // Initial Line notify
   line, err := SherryLineBot.NewLineBot()
   if err != nil {
      fmt.Println(err.Error())
      return 
   }

   PORT := ":" + arguments[1]
   l, err := net.Listen("tcp", PORT)
   if err != nil {
      fmt.Println(err)
      return
   }
   defer l.Close()


   // Get IP
   device := os.Getenv("InternetDevice")
   if device == "" {
      device = "wlan0"
   }
   ip, _ := IPService.NewIP(device)
   fmt.Printf(systemName + " start and listening at %s:%v......\n", ip.LocalIP, PORT)

   // Send IP to Line
   line.Notify.SendLINENotify(os.Getenv("LineNotifyToken"), systemName + " is online now, ip:" + ip.LocalIP)

   for {
      conn, err := l.Accept()
      if err != nil {
         fmt.Println(err)
      }
      go handleConnection(conn)
   }
}

