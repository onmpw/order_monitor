package config

import (
	"log"
	"os"
)

const commentIdentifier = '#'
const configFileName = "monitor.ini"

type Ini struct {
	Config					map[string]string
	configFilePath			string
}

var Conf *Ini

func Init() error{
	err := NewIni().Load()
	return err
}

// NewIni : 构造函数 初始化配置结构
func NewIni() *Ini {
	Conf = &Ini{
		Config: make(map[string]string),
		configFilePath: "/etc/",
	}

	return Conf
}

// Load :  加载配置文件配置项
func (config *Ini) Load() error{
	var configFileName = config.configFilePath+configFileName

	file,err := os.OpenFile(configFileName,os.O_RDONLY, 0755)
	if err != nil {
		log.Panic(err.Error())
	}

	length, err := file.Seek(0,2)
	if err != nil {
		return err
	}
	var b = make([]byte,length)

	_, err = file.Seek(0,0)
	if err != nil {
		return err
	}

	_,err = file.Read(b)
	if err != nil {
		return err
	}
	// 读取成功之后，加入一个结束符
	b = append(b,byte('\n'))

	var byteStr = make([]byte,0)

	var commentFlag = false

	var lineStart = true

	for _,char := range b {
		// # 作为注释
		if char == byte(commentIdentifier) && lineStart {
			commentFlag = true // 表示此行是注释行
			continue
		}

		if char == byte('\n') {
			lineStart = true
			if commentFlag {
				// 表示刚刚读取的那一行是注释行，下面过程不用处理直接重置 byteStr
				byteStr = byteStr[0:0]
				commentFlag = false  // 重置标识符
				continue
			}

			if len(byteStr) == 0 {  // 空行 直接读取下一行
				continue
			}

			flag := 0
			keyEnd := 0
			valueStart := 0
			for index,t := range byteStr {
				if t == byte(' ') || t == byte('\t') {
					if flag == 0{
						keyEnd = index
					}
					flag = 1
				}
				if flag == 1 && !(t == byte(' ') || t == byte('\t')) {
					valueStart = index
					break
				}
			}
			config.Config[string(byteStr[:keyEnd])] = string(byteStr[valueStart:])
			byteStr = byteStr[0:0]
			continue
		}
		if !commentFlag { // 注释行的字符不作处理
			lineStart = false
			byteStr = append(byteStr, char)
		}
	}
	return nil
}

func (config *Ini) C(key string) string {
	if val , ok := config.Config[key]; ok {
		return val
	}
	return ""
}

