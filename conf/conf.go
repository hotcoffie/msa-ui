package conf

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const configPath = "setting"
const userFile = "users.json"

var Users = make(map[string]string)

func init() {
	initUsers()
}

func initUsers() {
	fileName := filepath.Join(configPath, userFile)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
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

type ConfData struct {
	Active    string
	Username  string
	Password  string
	Points    string
	ThreadNum int
	Info      string
}

const confFormat = `# 系统状态
# prod - 正式抢点
# dev - 试访问网站检查参数是否正确
active: %s

# 账号密码
username: %s
password: %s

#可用时间点，时间点空格分隔
points: %s

# 同一个点用几个线程去抢
threadNum: %d`

func WritConfForRun(cd *ConfData) error {
	confFileName := filepath.Join(configPath, "conf.yml")
	confFile, err := os.OpenFile(confFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0766)
	if err != nil {
		return errors.WithMessage(err, "打开conf.yml")
	}
	defer confFile.Close()
	confStr := fmt.Sprintf(confFormat, cd.Active, cd.Username, cd.Password, cd.Points, cd.ThreadNum)
	if _, err = confFile.Write([]byte(confStr)); err != nil {
		return errors.WithMessage(err, "修改conf.yml")
	}

	infoFileName := filepath.Join(configPath, "info.yml")
	infoFile, err := os.OpenFile(infoFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0766)
	if err != nil {
		return errors.WithMessage(err, "打开info.yml")
	}
	defer infoFile.Close()
	if _, err = infoFile.Write([]byte(cd.Info)); err != nil {
		return errors.WithMessage(err, "修改info.yml")
	}
	return nil
}
