// 初始化
package utils

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	fmt.Println("Config app:", viper.Get("app"))
	fmt.Println("Config mysql:", viper.GetString("mysql.dns"))
}

func InitMySQL() {
	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{})
}
