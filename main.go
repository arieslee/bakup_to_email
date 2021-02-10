package main

import (
	"archive/zip"
	"fmt"
	"github.com/jordan-wright/email"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"
	"time"
)

const (
	dbHost = "localhost" // 数据库地址: localhost
	dbPort = "3306" // 数据库端口: 3306
	dbUser = "root" // 数据库用户名: root
	dbPwd = "12345678" // 数据库密码: 12345678
	dbName = "qhdcp_db" // 需要备份的数据库名: test
	tableName = "" // 需要备份的表名: test
	sqlPath = "./" // 备份SQL存储路径: /tmp/
	emailFrom = "abc@email.com"
	emailTo= "i@liming.me"
	smtpUser = "abc@aaa.com"
	smtpPwd = "*******"
	smtpHost = "smtp.aaa.com"
	smtpAddr = "smtp.aaa.com:25"
)

func main() {
	_, _ = BackupMySqlDb()
}

func BackupMySqlDb() (error,string)  {
	var cmd *exec.Cmd

	if tableName == "" {
		cmd = exec.Command("mysqldump", "--opt", "-h"+dbHost, "-P"+dbPort, "-u"+dbUser, "-p"+dbPwd, dbName)
	} else {
		cmd = exec.Command("mysqldump", "--opt", "-h"+dbHost, "-P"+dbPort, "-u"+dbUser, "-p"+dbPwd, dbName, tableName)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return err,""
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return err,""
	}

	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal(err)
		return err,""
	}
	now := time.Now().Format("20060102150405")
	var backupPath string
	if tableName == "" {
		backupPath = sqlPath+dbName+"_"+now+".sql"
	} else {
		backupPath = sqlPath+dbName+"_"+tableName+"_"+now+".sql"
	}
	err = ioutil.WriteFile(backupPath, bytes, 0644)

	if err != nil {
		panic(err)
		return err,""
	}
	// 压缩文件
	zipFileName := backupPath+".zip"
	err = compress(backupPath, zipFileName)
	if err != nil {
		panic(err)
		return err,""
	}
	// 发送邮件
	mailTo(zipFileName)
	return nil,zipFileName
}
func compress(file string, dest string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer d.Close()

	wr := zip.NewWriter(d)
	defer wr.Close()

	w, err := wr.Create(file)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

func mailTo(attachment string)  {
	e := email.NewEmail()
	e.From = "GolangBackupToEmail<"+emailFrom+">"
	e.To = []string{emailTo}
	e.Subject = "数据库备份"
	nowDate := time.Now().Format("2006-01-02 15:04:05")
	e.Text = []byte(nowDate+"的数据库备份文件")
	_, _ = e.AttachFile(attachment)
	err := e.Send(smtpAddr, smtp.PlainAuth("", smtpUser, smtpPwd, smtpHost))
	if err != nil{
		fmt.Println(err.Error())
	}
}