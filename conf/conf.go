package conf

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
)

const userFile = "users.json"

var Users = make(map[string]string)

func init() {
	initUsers()
}
func initUsers() {
	file, err := os.OpenFile(userFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Panic("打开用户信息:", err)
	}
	defer file.Close()

	bs, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic("读取用户信息:", err)
	}
	if bs != nil && len(bs) > 0 {
		if err = json.Unmarshal(bs, &Users); err != nil {
			log.Panic("解析用户信息:", err)
		}
	}
}

func UpdateUser(username, password string) error {
	p, ok := Users[username]
	if ok && p == password {
		return nil
	}
	Users[username] = password
	file, err := os.OpenFile(userFile, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Panic("打开用户信息:", err)
	}
	defer file.Close()
	data, _ := json.Marshal(Users)
	if _, err = file.Write(data); err != nil {
		return errors.WithMessage(err, "修改用户信息")
	}
	return nil
}
