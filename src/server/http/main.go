package main

import (
	"fmt"
	"storage"
	"weibo"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	db, err := sqlx.Open("mysql", "root:@tcp(127.0.0.1:3306)/weibo")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	userRepo := storage.NewUserRepository(db)
	service := weibo.NewService(userRepo)

	err = service.Register(&weibo.User{
		Account:  "lhc1",
		Avatar:   "xxx.jpg",
		Password: "123123",
	})

	if err == nil {
		fmt.Println("用户注册成功")
	} else {
		fmt.Println(err)
	}
}
