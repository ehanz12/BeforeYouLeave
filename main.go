package main

import (
	"fmt"

	"github.com/ehanz12/BeforeYouLeave/configs"
	"github.com/ehanz12/BeforeYouLeave/databases"
)

func main() {
  //init database
  configs.LoadEnv()
  databases.ConnectDB()
 fmt.Println("berhasil")
}
