package config

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/zbysir/blog/internal/pkg/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

// App is Config struct
var App struct {
	// 业务Debug
	Debug bool `yaml:"debug"`
	// OrmDebug开启后会打印sql语句
	OrmDebug bool `yaml:"orm_debug"`
	// LogDebug开启后会使用颜色
	LogDebug bool `yaml:"log_debug"`
	//Mysql    struct {
	//	Duckment string `yaml:"duckment"`
	//} `yaml:"mysql"`
	//Sqlite struct {
	//	Dbfile string `yaml:"dbfile"`
	//} `yaml:"sqlite"`
	Db struct {
		Source string `yaml:"source"`
	} `yaml:"db"`

	// 在测试环境和线上环境中, http服务应使用80端口, 微服务之间的Grpc应使用8080端口.
	HttpAddr   string `yaml:"http_addr"`
	Redis      string `yaml:"redis"`
	DnsService string `yaml:"dns_service"`

	// 只有当host = ui_host时, 才会访问到ui, 否则进入渲染
	UiHost string `yaml:"ui_host"`
}

// 初始化
// - config
// - log
// - orm调试模式
func init() {
	defConfigDir := defaultConfigDir()

	var confFile string

	pflag.StringVar(&confFile, "config", filepath.Join(defConfigDir, "config.yaml"), "set config file")
	pflag.String("db.source", "", "指定要使用的数据源, 可以是rpc")

	pflag.Parse()

	// Establishing defaults
	dbSource := "sqlite3://" + filepath.Join(defConfigDir, ".sqlite3")
	viper.SetDefault("db.source", dbSource)
	viper.SetDefault("debug", false)
	viper.SetDefault("log_debug", false)
	viper.SetDefault("orm_debug", false)
	viper.SetDefault("http_addr", ":8080")
	viper.SetDefault("redis", "localhost:3306")
	viper.SetDefault("ui_host", "duckment.bysir.top")

	log.Infof("loading config from file: `%s`", confFile)

	viper.SetConfigFile(confFile)

	// 读环境变量
	viper.AutomaticEnv()

	// 读flag`
	_ = viper.BindPFlag("db.source", pflag.Lookup("db.source"))

	// 读文件
	if err := viper.ReadInConfig(); err != nil {
		// 不存在创建, 全部使用默认值
		ioutil.WriteFile(confFile, nil, os.ModePerm)
	}

	err := viper.Unmarshal(&App, func(decoderConfig *mapstructure.DecoderConfig) {
		decoderConfig.TagName = "yaml"
	})
	if err != nil {
		panic(err)
	}
}

func defaultConfigDir() string {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	dir := filepath.Join(baseDir, "bblog")

	// 如果文件夹不存在, 则新建文件夹
	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModePerm)
		} else {
			panic(err)
		}
	}

	return dir
}
