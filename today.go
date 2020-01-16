package main

import (
	"fmt"
	t66y "go-es/src"
)

func today(output chan<- t66y.Progress) {
	downloader := t66y.NewDownloader(output)
	ctx, cancel := t66y.NewContext()
	defer cancel()

	startPage := 3
	for p := startPage; p > 0; p-- {
		fmt.Println("爬虫开始")
		anchors := t66y.ParsePage(ctx, p)
		for i, anchor := range anchors {
			output <- t66y.Progress{Message: fmt.Sprintf(
				"第 %d/%d 页，第 %d/%d 帖：%s %s",
				p, startPage,
				i+1, len(anchors),
				anchor.Href, anchor.Text,
				),
			}
			data := t66y.NewData(anchor.Text,anchor.Href,nil)
			if data.Fresh(){
				t66y.ParsePost(ctx,data,anchor)
				data.Save()
			}
			data.Download(downloader,output)
		}
	}
}
