// 初始化
package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitConfig() {
	viper.SetConfigName("app")      // name of config file (without extension)
	viper.SetConfigType("yaml")     // or whatever config type (e.g., "json")
	viper.AddConfigPath("./config") // path to look for the config file in

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	// fmt.Println("Config app:", viper.Get("app"))
	// fmt.Println("Config mysql:", viper.GetString("mysql.dns"))
	fmt.Println("Init app")
}

func InitMySQL() {
	//自定义日志模板 打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			//Sets the threshold duration for considering a SQL query as "slow". In this case, any query taking longer than one second will be marked as slow. This is useful for identifying potentially inefficient database queries.
			SlowThreshold: time.Second, //慢SQL阈值
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	fmt.Println("Init MySQL")
}
