package t66y

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type _ImageData struct {
	image *Image
	dir   string
	sync  chan<- bool
}
type _HostData struct {
	//当前线程数
	count int
	//当前的任务队列
	queue chan _ImageData
}

//downloader 下载器
type Downloader struct {
	hosts map[string]*_HostData
	//保护hosts
	lock sync.Mutex
	//单个域名最大线程数
	maxRoutines int
	//
	wg sync.WaitGroup
	//http 客户端
	client *http.Client
	//日志输出
	output chan<- Progress
	//剩余任务数
	taskCount int32
}

//新下载器
func NewDownloader(output chan<- Progress) *Downloader {
	d := &Downloader{hosts: map[string]*_HostData{},
		maxRoutines: 30,
		output:      output,
		client: &http.Client{
			Transport: &http.Transport{
				Dial:                (&net.Dialer{Timeout: time.Second * 5}).Dial,
				TLSHandshakeTimeout: time.Second * 5,
			},
			Timeout: time.Second * 300,
		},
	}
	return d
}

//wait 等待所有任务退出
func (d *Downloader) Wait() {
	d.wg.Wait()
}

//TaskCount 返回当前剩余任务数
func (d *Downloader) TaskCount() int32 {
	return atomic.LoadInt32(&d.taskCount)
}

//download 启动新的下载
func (d *Downloader) Download(ctx context.Context, dir string, images []*Image, sync chan<- bool) {
	for _, image := range images {
		queue, err := d.schedule(ctx, image.URL)
		//调度失败：链接无效
		if err != nil {
			d.output <- Progress{
				Dir:     dir,
				Message: fmt.Sprintf("cannot schedule: %s", image.URL),
			}
			sync <- false
			continue
		}
		queue <- _ImageData{
			image: image,
			dir:   dir,
			sync:  sync,
		}
	}
}

//动态调整某个域名的线程数
func (d *Downloader) schedule(ctx context.Context, link string) (chan<- _ImageData, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	host := strings.ToLower(u.Hostname())
	var data *_HostData
	if da, ok := d.hosts[host]; ok {
		data = da
	} else {
		data = &_HostData{
			queue: make(chan _ImageData, d.maxRoutines*10),
		}
		d.hosts[host] = data
	}

	if data.count < d.maxRoutines {
		data.count++
		d.wg.Add(1)
		go d.worker(ctx, host, data.queue)
	}
	atomic.AddInt32(&d.taskCount, 1)
	return data.queue, nil
}

//工作线程
func (d *Downloader) worker(ctx context.Context, host string, queue <-chan _ImageData) {
	defer func() {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.hosts[host].count--
		d.wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case id := <-queue:
			changed := d.do(ctx, id.dir, id.image)
			id.sync <- changed
			atomic.AddInt32(&d.taskCount, -1)
		case <-time.After(time.Second * 3):
			return
		}
	}
}

// 返回值表示是否有状态改变，不表示下载成功
func (d *Downloader) do(ctx context.Context, dir string, image *Image) bool {
	switch {
	case image.Status == "done":
		return false
	case image.Status == "code:404":
		return false
	case image.Status == "code:503":
		return false
	}
	req, err := http.NewRequest("GET", image.URL, nil)
	if err != nil {
		panic(err)
	}
	req = req.WithContext(ctx)

	resp, err := d.client.Do(req)
	if err != nil {
		d.output <- Progress{
			Dir:     dir,
			Message: err.Error(),
		}
		image.Status = err.Error()
		return true
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		d.output <- Progress{
			Dir:     dir,
			Message: resp.Status,
		}
		image.Status = fmt.Sprintf("code:%d", resp.StatusCode)
		return true
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" {
		if !strings.HasPrefix(strings.ToLower(contentType),"image/"){
			d.output<- Progress{
				Dir:     dir,
				Message: "not an image file",
			}
			image.Status = fmt.Sprintf("type:%s",contentType)
			return true
		}
	}
	name := d.guessFilename(image.URL,resp,contentType)
	path := filepath.Join(dir,name)

	image.Size = resp.ContentLength
	if image.Size>0 {
		info,err := os.Stat(path)
		if err == nil && info.Size() == image.Size{
			d.output <- Progress{
				Dir:     dir,
				Message: fmt.Sprintf("File %s already downloaded", path),
			}
			image.Status = "done"
			return true
		}
	}
	fp,err := os.Create(path)
	if err!= nil{
		panic(err)
	}
	defer fp.Close()

	n,err := io.Copy(fp,resp.Body)
	if err != nil{
		d.output <- Progress{
			Dir:     dir,
			Message: err.Error(),
		}
		image.Status = err.Error()
		return false
	}
	if image.Size >0 && n != image.Size{
		d.output <- Progress{
			Dir:     dir,
			Message: "truncated",
		}
		image.Status = "truncated"
		return true
	}else if image.Size == 0 {
		image.Size = n
	}
	d.output <- Progress{
		Dir:     dir,
		Message: fmt.Sprintf("file %s downloaded.",path),
	}
	image.Status = "done"
	return true
}
// guessFilename 猜测图片的文件名
// 应该优先使用响应头部的名字，这里从简。
func (d *Downloader) guessFilename(imageURL string, resp *http.Response, contentType string) string {
	u, err := url.Parse(imageURL)
	if err != nil {
		// 理应不会到达这里，因为前面以前发起过请求了
		panic(err)
	}

	name := filepath.Base(u.Path)
	ext := filepath.Ext(name)
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".gif", ".png":
		return name
	}

	// TODO 从响应头部获取文件名
	_ = resp

	name = base64.RawURLEncoding.EncodeToString([]byte(u.RequestURI()))

	// 从 ContentType 拿
	if exts, err := mime.ExtensionsByType(contentType); err == nil && len(exts) > 0 {
		return name + exts[0]
	}
	return name
}
