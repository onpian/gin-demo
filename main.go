package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io/ioutil"
	"os"

	//_ "github.com/spf13/viper"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	gorm.Model
	UserName string `gorm:"column:username;type:varchar(20);not null" json:"username" form:"username"`
	PassWord string `gorm:"column:password;type:varchar(16);not null" json:"password" form:"password"`
	TelPhone string `gorm:"column:telphone;type:varchar(11);not null;unique" json:"telphone" form:"telphone"`
}

func main() {
	Db := InitDB()
	defer Db.Close()

	r := gin.Default()
	r.Handle("POST", "/api/auth/register", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		telphone := ctx.PostForm("telphone")
		//数据验证
		rand.Seed(time.Now().Unix())
		if len(telphone) != 11 {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "手机号不正确!"})
			return
		}
		if len(password) < 6 {
			ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "密码不能少于6位"})
			return
		}
		if len(username) == 0 {
			username = RandomString(6)
		}
		log.Println(username, password, telphone)

		//查询手机号
		if isTelPhoneExist(Db, telphone) == true {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"code": 422, "msg": "用户已注册!"})
			return
		}
		newUser := User{
			UserName: username,
			PassWord: password,
			TelPhone: telphone,
		}
		Db.Create(&newUser)

		ctx.JSON(200, gin.H{"code": 200, "msg": "注册成功!"})
	})

	r.Run(":80")
}

func isTelPhoneExist(db *gorm.DB, phone string) bool {
	var user User
	db.Where("telphone=?", phone).First(&user)
	fmt.Println("看看啥意思", &user)
	if user.ID != 0 {
		return true
	}
	return false
}

func RandomString(n int) string {
	var res string
	rand.Seed(time.Now().Unix())
	for i := 0; i < n-2; i++ {
		res += strconv.Itoa(rand.Intn(10))
	}
	return "CC" + string(res)
}
func InitDB() *gorm.DB {
	type LinkStr struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		DataBase string `json:"data_base"`
		RootName string `json:"root_name"`
		PassWord string `json:"pass_word"`
		Charset  string `json:"charset"`
	}

	f, err := os.Open("conf/db.json")
	if err != nil {
		fmt.Println("打开文件失败!", err)
	}
	defer f.Close()
	jsonByte, err2 := ioutil.ReadAll(f)
	fmt.Println(string(jsonByte))

	if err2 != nil {
		fmt.Println(err2)
		panic("JSON配置文件读取失败!")
	}
	var linkStr LinkStr
	err3 := json.Unmarshal(jsonByte, &linkStr)
	if err3 != nil {
		fmt.Println("看看ERR3", err3)
		panic("JSON配置文件读取失败!")
	}

	args := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", linkStr.RootName, linkStr.PassWord,
		linkStr.Host, linkStr.Port, linkStr.DataBase, linkStr.Charset)
	db, err4 := gorm.Open("mysql", args)
	if err4 != nil {
		log.Fatal("	db, err4 := gorm.Open err: ", err4)
	}
	if !db.HasTable("user") {
		db.AutoMigrate(&User{})
	}
	return db
}
