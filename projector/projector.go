package main

import (
   "os"
   "fmt"
   "net"
   "time"
   "bytes"
   "strings"
   "syscall"
   "context"
   "os/exec"
   "runtime"
   "net/http"
   "io/ioutil"
   "os/signal"
   "github.com/joho/godotenv"
   "github.com/asccclass/staticfileserver/libs/ip" 
   "github.com/asccclass/staticfileserver/libs/line" 
)

var (
   systemName   string
   IPAddr       string
   PORT         string
)

// off line
func OffLine(ctx context.Context) {
   stop := make(chan bool)
   go func() {
      url := os.Getenv("ProjectorServer")
      if url != "" {
         url += "/iamoffline/" + IPAddr 
         _, err := http.Get(url)
         if err != nil {
            fmt.Println(err.Error())
         }
      } else {
         fmt.Println("Can not find ProjectorServer in envfile.")
      }
      stop <- true
   }()
   select {
      case <- ctx.Done():
         fmt.Println("Time is out.")
         break
      case <-stop:
         fmt.Println("\n" + systemName + " off line...")
         break
   }
   return
}

// send ip information to projector server
func RegisterInfo(ip, port string)(error) {
   url := os.Getenv("ProjectorServer")
   if url == "" {
      return fmt.Errorf("Can not find ProjectorServer in envfile.")
   }
   url += "/iamonline"
   info := fmt.Sprintf("{\"name\":\"%s\", \"ip\":\"%s\", \"port\":\"%s\"}", os.Getenv("SystemName"), ip, port)

   req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(info)))
   if err != nil {
      return err
   }
   req.Header.Set("Content-Type", "application/json; charset=UTF-8")
   client := &http.Client{}
   res, err := client.Do(req)
   if err != nil {
      return err
   }
   defer res.Body.Close()
   body, err := ioutil.ReadAll(res.Body)
   if err != nil {
      return err
   }
   fmt.Println(string(body)) 
   return nil
}

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

// process connection's action
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
      fmt.Printf("  %v %s\n", myTime, message)
   }
   conn.Close()
}

func main() {
   systemName = os.Getenv("SystemName")
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

   PORT = ":" + arguments[1]
   l, err := net.Listen("tcp", PORT)
   if err != nil {
      fmt.Println(err.Error())
      return
   }
   defer l.Close()

   // Get IP
   device := os.Getenv("InternetDevice")
   if device == "" {
      device = "wlan0"
   }

   ip, _ := IPService.NewIP(device)
   IPAddr = ip.LocalIP
   fmt.Printf("%s start and listening at %s%s......\n", systemName, IPAddr, PORT)
   if err := RegisterInfo(IPAddr, PORT); err != nil {
      fmt.Println(err)
      return
   } 

   // Send IP to Line
   line.Notify.SendLINENotify(os.Getenv("LineNotifyToken"), systemName + " is online now, ip:" + IPAddr + PORT)

   interrupt := make(chan os.Signal)
   signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
   go func() {
      <-interrupt
      ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
      defer cancel()
      OffLine(ctx)
      os.Exit(0)
   }()

   for {
      conn, err := l.Accept()
      if err != nil {
         fmt.Println(err)
      }
      go handleConnection(conn)
   }
}

