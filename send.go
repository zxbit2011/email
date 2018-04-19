package main

import (
	"email/common"

	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func send(c echo.Context) error {
	user := c.FormValue("account")
	password := c.FormValue("pwd")
	subject := c.FormValue("title")
	body := c.FormValue("content")
	to := c.FormValue("email")
	file, err := c.FormFile("file")
	filePath := ""
	if err != nil {
		println("Send mail File! " + err.Error())
	} else {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusOK, "Send mail File error!"+err.Error())
		}
		defer src.Close()
		dir := "files"
		path := file.Filename
		println("file：" + path + "         ")
		flog, srcDir, errDir := checkDir(dir)
		if !flog {
			return c.JSON(http.StatusOK, "Send mail File error!"+errDir.Error())
		}
		filePath = srcDir + getSeparator() + file.Filename

		dst, err := os.Create(filePath)
		if err != nil {
			return c.JSON(http.StatusOK, "Send mail File error!"+err.Error())
		}
		defer dst.Close()

		// Copy
		if _, err = io.Copy(dst, src); err != nil {
			return c.JSON(http.StatusOK, "Send mail File error!"+err.Error())
		}
	}

	host := getSMTP(user) //"smtp.163.com:25"
	println(user, password, host, to, subject, body, "html", filePath)
	err = sendEmail(user, password, host, to, subject, body, filePath)
	if err != nil {
		return c.JSON(http.StatusOK, "Send mail error!"+err.Error())
	} else {
		return c.JSON(http.StatusOK, "Send mail success!")
	}
	return c.JSON(http.StatusOK, "")
}

func main() {
	e := echo.New()
	e.Static("/page", "page")
	e.File("/", "index.html")
	e.POST("/send", send)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	Open("http://localhost:8081/")
	e.Start(":8081")
}

func sendEmail(user, password, host, to, subject, body, filePath string) error {
	e := common.NewEmail()
	e.From = user
	// 去除空格
	to = strings.Replace(to, " ", "", -1)
	// 去除换行符
	to = strings.Replace(to, "\n", "", -1)
	e.To = strings.Split(to, ";")
	e.Subject = subject
	e.HTML = []byte(body)
	if filePath != "" {
		e.AttachFile(filePath)
	}
	hp := strings.Split(host, ":")
	err := e.Send(host, smtp.PlainAuth("", user, password, hp[0]))
	return err
}

var commands = map[string]string{
	"windows": "cmd /c start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

func Open(uri string) error {
	run, ok := commands[runtime.GOOS]
	if !ok {
		return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	}
	println("操作系统：", runtime.GOOS)
	cmd := exec.Command(run, uri)
	return cmd.Start()
}

func checkDir(pathDir string) (bool, string, error) {
	path := getSeparator()
	dir, _ := os.Getwd()
	src := dir + path + pathDir
	_, err := os.Stat(pathDir)
	if err == nil {
		return true, src, nil
	}
	//不存在创建目录
	if os.IsNotExist(err) { //当前的目录
		err = os.Mkdir(src, os.ModePerm) //在当前目录下生成md目录
		if err == nil {
			return true, src, nil
		}
		return false, src, err
	}
	return false, src, err

}

//系统分隔符
func getSeparator() string {
	var path string
	if os.IsPathSeparator('\\') { //前边的判断是否是系统的分隔符
		path = "\\"
	} else {
		path = "/"
	}
	return path
}

func getSMTP(account string) string {
	if account == "" {
		return ""
	}
	m := strings.Split(account, "@")
	if len(m) <= 1 {
		return ""
	}
	switch m[1] {
	case "163.com":
		return "smtp.163.com:25"
	case "126.com":
		return "smtp.126.com:25"
	case "139.com":
		return "smtp.139.com:25"
	case "qq.com":
		return "smtp.qq.com:25"
	case "sohu.com":
		return "smtp.sohu.com:25"
	case "vip.sina.com":
		return "smtp.vip.sina.com:25"
	case "sina.com":
		return "smtp.sina.com:25"
	}
	return ""
}
