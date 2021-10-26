package main

import (
   "os"
   "fmt"
   "net"
   "bufio"
   "os/exec"
   "runtime"
   "strings"
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

func main() {
   arguments := os.Args
   if len(arguments) == 1 {
      fmt.Println("Please provide host:port.")
      return
   }

   CONNECT := arguments[1]
   c, err := net.Dial("tcp", CONNECT)
   if err != nil {
      fmt.Println(err)
      return
   }

   for {
      reader := bufio.NewReader(os.Stdin)
      fmt.Print(">> ")
      text, _ := reader.ReadString('\n')  // send data
      fmt.Fprintf(c, text + "\n")  // to server

      if strings.TrimSpace(string(text)) == "STOP" {
         fmt.Println("TCP client exiting...")
         return
      }
      message, _ := bufio.NewReader(c).ReadString('\n')
      if len(message) > 0 {
         message = message[:len(message) - 1]
         fmt.Printf("open %s now...\n", message)
         openBrowser(message)
      }
   }
}
