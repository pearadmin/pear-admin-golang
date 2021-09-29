package remote

import (
	"fmt"
	"github.com/cilidm/toolbox/OS"
	"github.com/cilidm/toolbox/logging"
	"github.com/pkg/sftp"
	"os"
	"path"
	"pear-admin-go/app/global"
	"strings"
)

type Remote struct {
	sourcePath   string       // 源路径
	dstPath      string       // 目标路径
	sourceClient *sftp.Client // 源服务器连接
	dstClient    *sftp.Client // 目标服务器连接
}

func NewRemote(sourcePath string, dstPath string, sourceClient *sftp.Client, dstClient *sftp.Client) *Remote {
	return &Remote{sourcePath: sourcePath, dstPath: dstPath, sourceClient: sourceClient, dstClient: dstClient}
}

// 校验目标地址权限
func (this *Remote) CheckAccess() {

}

func (this *Remote) LocalToRemote(fname string, fsize int64) error {
	if OS.IsWindows() {
		this.dstPath = strings.ReplaceAll(this.dstPath, "\\", "/")
	}
	rf := path.Join(this.dstPath, fname) // 文件在服务器的路径及名称
	has, err := this.dstClient.Stat(rf)
	if err == nil && (has.Size() == fsize) {
		global.Log.Debug(fmt.Sprintf("文件%s已存在", rf))
		return nil
	}
	this.dstClient.MkdirAll(this.dstPath)
	err = this.dstClient.Chmod(this.dstPath, os.ModePerm)
	if err != nil {
		return err
	}
	srcFile, err := os.Open(path.Join(this.sourcePath, fname))
	if err != nil {
		logging.Error("源文件无法读取", err.Error())
		return err
	}
	defer srcFile.Close()
	dstFile, err := this.dstClient.Create(rf) // 如果文件存在，create会清空原文件 openfile会追加
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 10000)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[:n]) // 读多少 写多少
	}
	return nil
}

func (this *Remote) RemoteToRemote() {

}

func (this *Remote) RemoteToLocal() {

}