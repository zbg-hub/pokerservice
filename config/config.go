package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AccountAddr string `yaml:"AccountAddr"`
	ServicePort string `yaml:"ServicePort"`
}

var conf *Config

func initConfig() {
	conf = &Config{}
	ex, _ := os.Executable()
	pwd := filepath.Dir(ex)
	confPath := getParentDirectory(pwd) + "/conf" + "/base.yaml"
	fmt.Printf("confPath is %s\n", confPath)
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		fmt.Printf("read conf err is %s\n", err)
	}
	if err := yaml.Unmarshal(buf, conf); err != nil {
		fmt.Printf("prase conf err: %s\n", err)
	}
	return
}

func GetConfig() *Config {
	if conf == nil {
		initConfig()
	}
	return conf
}

func substr(s string, pos, length int) string {
	runes := []rune(s)
	l := pos + length
	if l > len(runes) {
		l = len(runes)
	}
	return string(runes[pos:l])
}

func getParentDirectory(dirctory string) string {
	return substr(dirctory, 0, strings.LastIndex(dirctory, "/"))
}
