package main

import (
	"ginchat/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:Hl011028@tcp(127.0.0.1:3306)/golang?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.UserBasic{})

	// // Create
	// user := &models.UserBasic{}
	// user.Name = "testUser"

	// db.Create(user)

	// Read

	// fmt.Println(db.First(user, 1)) // find product with integer primary key

	// Update - update product's price to 200
	// db.Model(user).Update("Password", "1234")
	// Update - update multiple fields
	//db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	//db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	//db.Delete(&product, 1)
}
