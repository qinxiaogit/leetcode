package t66y

import (
	"context"
	//"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//Image 存图片信息
type Image struct {
	URL    string `yaml:"url"`
	Size   int64  `yaml:"size"`
	Status string `yaml:"status"`
}

//Post 一篇帖子相关信息
type Post struct {
	ID      string    `yaml:"id"`
	Name    string    `yaml:"name"`
	URL     string    `yaml:"url"`
	Images  []*Image  `yaml:"images"`
	Created time.Time `yaml:"created"`
}

type Progress struct {
	Dir     string
	Message string
}

// Data 管理帖子的图片信息
type Data struct {
	first bool
	Post  Post
	_dir  string
	_date string
	sync  chan bool
}

var reDir = regexp.MustCompile(`^https://t66y\.com/htm_data/((\d+)/(\d+)/(\d+))\.html$`)

func dir(url string) string {
	numbers := reDir.FindStringSubmatch(url)
	if numbers == nil {
		panic("no id")
	}
	return numbers[1]
}

func NewData(name string, url string, images []string) *Data {
	d := &Data{
		Post: Post{
			ID:      dir(url),
			Name:    name,
			URL:     url,
			Created: time.Now(),
		},
		sync: make(chan bool),
	}
	d.setImages(images)
	d.init()
	return d
}

//NewDataFromFile 指定数据文件加载数据
func NewDataFromFile(data string) *Data {
	d := &Data{
		sync: make(chan bool),
	}
	if err := d.load(data); err != nil {
		panic(err)
	}
	return d
}

func (d *Data) setImages(images []string) {
	d.Post.Images = d.Post.Images[:0]
	for _, image := range images {
		d.Post.Images = append(d.Post.Images, &Image{
			URL: image,
		})
	}
}

func (d *Data) dir() string {
	if d._dir == "" {
		path := filepath.Join("data", d.Post.ID)
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(err)
		}
		d._dir = path
	}
	return d._dir
}

//返回当天对应的符号链接目录名
// 返回对应的当天的符号连接目录名
func (d *Data) date(day time.Time) string {
	if d._date == "" {
		name := d.Post.Name

		// 目录分隔符会出错
		// TODO 没考虑 Windows 特殊字符
		name = strings.ReplaceAll(name, "/", "")

		// 太长会出错
		if len(name) >= 200 {
			runes := []rune(name)
			if len(runes) >= 200 {
				name = string(runes[:200])
			}
		}

		// 追加ID
		name += fmt.Sprintf("[%s]", strings.ReplaceAll(d.Post.ID, "/", "-"))

		// 当天的目录名
		now := day.Format(`2006/01/02`)

		dir := fmt.Sprintf("date/%s", now)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}

		d._date = fmt.Sprintf("%s/%s", dir, name)
	}
	return d._date
}

// Save ...
func (d *Data) Save() {
	fp, err := os.Create(d.data())
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	if err := yaml.NewEncoder(fp).Encode(d.Post); err != nil {
		panic(err)
	}
}
func (d *Data) data() string {
	return filepath.Join(d.dir(), "data.yml")
}

// 从 path 加载数据文件。
// 如果 path 为空，则从 data() 目录加载。
func (d *Data) load(path string) error {
	if path == "" {
		path = d.data()
	}
	fp, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	return yaml.NewDecoder(fp).Decode(&d.Post)
}

func (d *Data) init() {
	if _, err := os.Stat(d.data()); err == nil {
		if err := d.load(""); err != nil {
			log.Println("数据文件错误", err)
			d.Save()
			d.first = true
		} else {
			d.first = false
		}
	} else {
		d.first = true
	}
}

// Fresh 判断是否是新的下载
func (d *Data) Fresh() bool {
	return d.first || d.Post.Name == ""
}

func (d *Data) close() {
	close(d.sync)
}

// Download 启动下载
func (d *Data) Download(downloader *Downloader, output chan<- Progress) {
	d.link()

	if len(d.Post.Images) == 0 {
		return
	}

	// 统计进度
	i, total := 0, len(d.Post.Images)
	numChanged := 0

	go func() {
		for changed := range d.sync {
			if changed {
				numChanged++
				d.Save()
			}
			if i++; i == total {
				// 仅在有变更的情况下才报告
				if numChanged > 0 {
					output <- Progress{
						Dir:     d.dir(),
						Message: "全部图片处理完成",
					}
				}
				d.close()
			}
		}
	}()

	downloader.Download(context.Background(), d.dir(), d.Post.Images, d.sync)
}

// 当天下载的全部放到当天日期的目录
func (d *Data) link() {
	// 不是今天新创建的
	created := d.Post.Created.Day()
	today := time.Now().Day()
	if created != today {
		return
	}

	if err := os.Symlink("../../../../"+d.dir(), d.date(time.Now())); err != nil {
		//if errors.Is(err, os.ErrExist) {
		//	return
		//}
		panic(err)
	}

	log.Println("创建符号连接：", d.dir(), d.date(time.Now()))
}

// LinkTo 符号连接到指定的日期
func (d *Data) LinkTo(day time.Time) {
	if err := os.Symlink("../../../../"+d.dir(), d.date(day)); err != nil {
		//if errors.Is(err, os.ErrExist) {
		//	return
		//}
		panic(err)
	}
}
