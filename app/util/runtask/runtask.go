package runtask

import (
	"fmt"
	"github.com/cilidm/toolbox/OS"
	"github.com/cilidm/toolbox/file"
	"github.com/cilidm/toolbox/logging"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"pear-admin-go/app/core/scli"
	"pear-admin-go/app/dao"
	"pear-admin-go/app/global"
	"pear-admin-go/app/global/e"
	"pear-admin-go/app/model"
	"pear-admin-go/app/util/check"
	"pear-admin-go/app/util/pool"
	"strings"
	"sync/atomic"
	"time"
)

type RunTask struct {
	task         model.Task
	sourceClient *sftp.Client // 源服务器连接
	dstClient    *sftp.Client // 目标服务器连接
	fp           *pool.Pool
	counter      uint64
}

func (this *RunTask) SetSourceClient() *RunTask {
	if this.task.SourceType == e.Local {
		return this
	}
	c, err := this.getClient(this.task.SourceServer)
	if err != nil {
		global.Log.Error("SetSourceClient.getClient", zap.Error(err))
		return this
	}
	this.sourceClient = c
	return this
}

func (this *RunTask) SetDstClient() *RunTask {
	if this.task.DstType == e.Local {
		return this
	}
	c, err := this.getClient(this.task.DstServer)
	if err != nil {
		global.Log.Error("SetSourceClient.getClient", zap.Error(err))
		return this
	}
	this.dstClient = c
	return this
}

func (this *RunTask) getClient(sid int) (*sftp.Client, error) {
	server, err := dao.NewTaskServerDaoImpl().FindOne(sid)
	if err != nil {
		return nil, err
	}
	c, err := scli.Instance(*server)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (this *RunTask) Run() {
	if this.task.SourceType == e.Remote && this.task.DstType == e.Remote {
		this.RunR2R()
	} else if this.task.SourceType == e.Remote && this.task.DstType == e.Local {
		this.RunR2L()
	} else if this.task.SourceType == e.Local && this.task.DstType == e.Remote {
		this.RunL2R()
	}
}

func (this *RunTask) RunR2R() {

}

func (this *RunTask) RunR2L() {
	this.WalkRemotePath(this.task.SourcePath)
}

func (this *RunTask) WalkRemotePath(dirPath string) {
	globPath := pathJoin(dirPath)
	files, err := this.sourceClient.Glob(globPath)
	if err != nil {
		logging.Error(err)
		return
	}
	for _, v := range files {
		stat, err := this.sourceClient.Stat(v)
		if err != nil {
			global.Log.Error("WalkRemotePath.this.sourceClient.Stat", zap.Error(err))
			continue
		}
		if stat.IsDir() {
			this.WalkRemotePath(v)
		} else {
			this.fp.Add(1)
			atomic.AddUint64(&this.counter, 1)
			go func(v string, size int64) {
				defer this.fp.Done()
				err = this.RemoteSend(v, size)
				if err != nil {
					global.Log.Error("WalkRemotePath.RemoteToLocal", zap.Error(err))
				}
			}(v, stat.Size())
		}
	}
}

// 远端->本地 使用 sourceClient
func (this *RunTask) RemoteSend(fname string, fsize int64) error { // 本地文件夹
	dstFile := path.Join(this.task.DstPath, strings.ReplaceAll(fname, this.task.SourcePath, "")) // 需要保存的本地文件地址
	has, err := check.CheckFile(dstFile)                                                         // 是否已存在
	if err != nil {
		global.Log.Error("RemoteToLocal.CheckFile", zap.Error(err))
		return err
	}
	if has != nil && has.Size() == fsize {
		global.Log.Debug(fmt.Sprintf("文件%s已存在", dstFile))
		return nil
	}
	dir, _ := path.Split(dstFile)
	err = file.IsNotExistMkDir(dir)
	if err != nil {
		global.Log.Error("RemoteToLocal.IsNotExistMkDir", zap.Error(err))
		return err
	}

	srcFile, err := this.sourceClient.Open(fname)
	if err != nil {
		global.Log.Error("RemoteToLocal.sourceClient.Open", zap.Error(err))
		return err
	}
	defer srcFile.Close()
	lf, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer lf.Close()

	if _, err = srcFile.WriteTo(lf); err != nil {
		return err
	}
	global.Log.Info(fmt.Sprintf("copy %s finished!", srcFile.Name()))
	return nil
}

func pathJoin(p string) (np string) {
	if strings.HasSuffix(p, "/") == false {
		p = p + "/"
	}
	if OS.IsWindows() {
		np = strings.ReplaceAll(path.Join(p, "*"), "\\", "/")
	} else {
		np = path.Join(p, "*")
	}
	return
}

// 本地->远端
func (this *RunTask) RunL2R() {
	_ = filepath.Walk(this.task.SourcePath, func(v string, info fs.FileInfo, err error) error {
		if err != nil {
			global.Log.Error("RunL2R.Walk.err", zap.Error(err))
			return nil
		}
		if info == nil {
			global.Log.Error("RunL2R.Walk.info Is Nil")
			return nil
		}
		stat, err := os.Stat(v)
		if err != nil {
			global.Log.Error("RunL2R.os.Stat", zap.Error(err))
			return nil
		}
		if stat.IsDir() && v != this.task.SourcePath {
			dname := string([]rune(strings.ReplaceAll(v, this.task.SourcePath, ""))[1:])
			err = this.MkRemotedir(dname)
			if err != nil {
				global.Log.Error("RunL2R.rm.Mkdir", zap.Error(err))
				return nil
			}
		} else {
			atomic.AddUint64(&this.counter, 1)
			this.fp.Add(1)
			go func(v string, size int64) {
				defer this.fp.Done()
				err = this.LocalSend(v, size)
				if err != nil {
					global.Log.Error("WalkPath.LocalToRemote", zap.Error(err))
				}
			}(v, stat.Size())
		}
		return nil
	})
}

func (this *RunTask) LocalSend(fname string, fsize int64) error {
	if OS.IsWindows() {
		this.task.SourcePath = strings.ReplaceAll(this.task.SourcePath, "\\", "/")
		this.task.DstPath = strings.ReplaceAll(this.task.DstPath, "\\", "/")
		fname = strings.ReplaceAll(fname, "\\", "/")
		fname = strings.ReplaceAll(fname, this.task.SourcePath, "")
	}
	rf := path.Join(this.task.DstPath, fname) // 文件在服务器的路径及名称
	has, err := this.dstClient.Stat(rf)
	if err == nil && (has.Size() == fsize) {
		global.Log.Debug(fmt.Sprintf("文件%s已存在", rf))
		return nil
	}
	err = this.dstClient.MkdirAll(this.task.DstPath)
	if err != nil {
		return err
	}
	err = this.dstClient.Chmod(this.task.DstPath, os.ModePerm)
	if err != nil {
		return err
	}
	srcFile, err := os.Open(path.Join(this.task.SourcePath, fname))
	if err != nil {
		global.Log.Error("源文件无法读取", zap.Error(err))
		return err
	}
	defer srcFile.Close()
	dstFile, err := this.dstClient.Create(rf) // 如果文件存在，create会清空原文件 openfile会追加
	if err != nil {
		global.Log.Error("this.dstClient.Create", zap.Error(err))
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
	err = dao.NewTaskLogDaoImpl().Insert(model.TaskLog{
		TaskId:     this.task.Id,
		ServerId:   this.task.DstServer,
		SourcePath: path.Join(this.task.SourcePath, fname),
		DstPath:    rf,
		Size:       fsize,
		CreateTime: time.Now(),
	})
	if err != nil {
		return err
	}
	global.Log.Info(fmt.Sprintf("【%s】传输完毕", fname))
	return nil
}

func (this *RunTask) MkRemotedir(p string) error {
	p = check.CheckWinPath(p)
	dst := path.Join(this.task.DstPath, p)
	err := this.dstClient.MkdirAll(dst)
	if err != nil {
		return err
	}
	return nil
}

func NewRunTask(task model.Task) *RunTask {
	return &RunTask{task: task, fp: pool.NewPool(e.MaxPool), counter: 0}
}