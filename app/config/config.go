package config

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zhangpanyi/basebot/logger"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

// 服务配置
type Serve struct {
	Host              string  `yaml:"host"`                 // 主机地址
	Port              int     `yaml:"port"`                 // HTTP端口
	APIAccess         string  `yaml:"api_access"`           // API接入点
	SupportStaff      *int64  `yaml:"support_staff"`        // 电报客服ID
	SecretKey         string  `yaml:"secret_key"`           // 验证码密钥
	Token             string  `yaml:"token"`                // 机器人token
	Name              string  `yaml:"name"`                 // 资产名称
	Symbol            string  `yaml:"symbol"`               // 资产符号
	Precision         int     `yaml:"precision"`            // 资产精度
	WithdrawFee       float64 `yaml:"withdraw_fee"`         // 提现手续费
	BolTDBPath        string  `yaml:"boltdb_path"`          // BoltDB路径
	Languages         string  `yaml:"languages"`            // 语言配置路径
	Expire            uint32  `yaml:"expire"`               // 红包过期时间
	MaxMessageLen     int     `yaml:"max_message_len"`      // 最大留言长度
	MaxHistoryTextLen int     `yaml:"max_history_text_len"` // 历史文本长度
	ThumbURL          string  `yaml:"thumb_url"`            // 红包缩略图URL
}

// 配置解析器
type parser interface {
	parse([]byte) error
}

// 配置管理器
type Manager struct {
	serve      *Serve
	languges   *Languges
	watcher    *fsnotify.Watcher
	fileparser map[string]parser
}

// 获取服务配置
func GetServe() Serve {
	return *globalManager.serve
}

// 获取语言配置
func GetLanguge() *Languges {
	return globalManager.languges
}

// 加载配置文件
func LoadConfig(path string) {
	once.Do(func() {
		// 创建观察器
		fileparser := make(map[string]parser)
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}

		// 加载主配置
		data, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		serve := Serve{}
		err = yaml.Unmarshal(data, &serve)
		if err != nil {
			panic(err)
		}

		// 加载语言包配置
		languages, files := readLanguages(serve.Languages)
		for _, filename := range files {
			watcher.Add(filename)
			fileparser[filename] = languages
		}

		// 初始化全局配置
		globalManager = &Manager{
			serve:      &serve,
			languges:   languages,
			fileparser: fileparser,
			watcher:    watcher,
		}
		go globalManager.watch()
	})
}

// 全局配置管理器
var once sync.Once
var globalManager *Manager

// 观察文件变更
func (m *Manager) watch() {
	for {
		select {
		case evt := <-m.watcher.Events:
			if evt.Op&fsnotify.Write == fsnotify.Write {
				handler, ok := m.fileparser[evt.Name]
				if ok {
					data, err := ioutil.ReadFile(evt.Name)
					if err == nil {
						err = handler.parse(data)
						if err != nil {
							logger.Warnf("File notify: parse failed, %v, %v", evt.Name, err)
						} else {
							logger.Infof("File notify: realod file finished, %v", evt.Name)
						}
					} else {
						logger.Warnf("File notify: handle failed, %v, %v", evt.Name, err)
					}
				} else {
					logger.Warnf("File notify: ignore write event, %v", evt.Name)
				}
			}
		case err := <-m.watcher.Errors:
			logger.Warnf("File notify: recv error event, %v", err)
		}
	}
}

// 读取语言包配置
func readLanguages(dir string) (*Languges, []string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	languges := NewLanguges()
	paths := make([]string, 0)
	for _, v := range files {
		if !v.IsDir() {
			ext := strings.ToLower(filepath.Ext(v.Name()))
			if ext == ".lang" {
				fullname := dir + string(filepath.Separator) + v.Name()
				data, err := ioutil.ReadFile(fullname)
				if err != nil {
					panic(err)
				}
				err = languges.parse(data)
				if err != nil {
					panic(err)
				}
				paths = append(paths, fullname)
			}
		}
	}
	return languges, paths
}
