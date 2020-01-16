package t66y

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
)

// Anchor 帖子信息
type Anchor struct {
	Href string
	Text string
}

// NewContext 创建新的浏览器实例
func NewContext() (context.Context, context.CancelFunc) {
	chromedp.DefaultExecAllocatorOptions = [25]chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		//chromedp.Headless,
		//chromedp.NoFirstRun,

		// After Puppeteer's default behavior.
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-backgrounding-occluded-windows", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("password-store", "basic"),
		chromedp.Flag("use-mock-keychain", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	chromedp.Run(ctx, network.Enable())
	chromedp.ListenTarget(ctx, onTargetEvent)
	return ctx, cancel
}

// ParsePage 分析具体的某一分页上的所有帖子链接
// 游客只能访问前 100 页
func ParsePage(ctx context.Context, page int) []Anchor {
	var res []byte

	if err := chromedp.Run(ctx,
		chromedp.Navigate(`https://www.v2ex.com/api/topics/hot.json?page=`+fmt.Sprint(page)),
		chromedp.WaitVisible("#ajaxtable"),
		chromedp.EvaluateAsDevTools(`
			var list = [];
			var anchors = document.querySelectorAll('#ajaxtable tr.t_one h3 a');
			anchors.forEach(a=>{
				list.push(a.href+'|'+encodeURIComponent(a.innerText));
			});
			list;
			`, &res,
		),
	); err != nil {
		panic(err)
	}

	var ss []string
	if err := json.Unmarshal(res, &ss); err != nil {
		panic(err)
	}

	var anchors []Anchor

	for _, s := range ss {
		parts := strings.Split(s, "|")
		anchors = append(anchors, Anchor{
			Href: parts[0],
			Text: queryUnescape(parts[1]),
		})
	}

	return anchors
}

// IsValidLink 报告是否是合法的帖子链接
func IsValidLink(link string) bool {
	// 部分置顶的帖子不是用户帖，但是CSS无法区分出来
	if strings.Contains(link, "fid") {
		return false
	}
	// 目录格式要一致
	if !reDir.MatchString(link) {
		return false
	}
	return true
}

// ParsePost 分析某篇帖子的图片等信息
func ParsePost(ctx context.Context, data *Data, anchor Anchor) {
	var result []byte
	var title string

	var document *network.EventResponseReceived
	var documentErr error
	OnDocument = func(err error, sent *network.EventRequestWillBeSent, recv *network.EventResponseReceived) {
		if sent.DocumentURL == anchor.Href && sent.Initiator.Type == "other" {
			if err != nil {
				documentErr = err
				return
			}
			document = recv
		}
	}
	mustRun(ctx, chromedp.Navigate(anchor.Href))
	for {
		if document != nil || documentErr != nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	if documentErr != nil {
		panic(documentErr)
	}
	if document.Response.Status != 200 {
		return
	}

	mustRun(ctx, chromedp.WaitReady(`div.tpc_content`, chromedp.ByQuery))
	mustRun(ctx, chromedp.EvaluateAsDevTools(`
				var images = document.querySelectorAll('.tpc_content img[data-src],.tpc_content input[type=image]');
				var list = [];
				images.forEach(i => {
					list.push(i.getAttribute('data-src'));
				});
				list;`, &result))
	mustRun(ctx, chromedp.EvaluateAsDevTools(`document.querySelector('h4').innerText`, &title))

	var images []string
	if err := json.Unmarshal(result, &images); err != nil {
		panic(err)
	}

	data.Post.Name = title
	data.setImages(images)
}

func queryUnescape(s string) string {
	t, _ := url.QueryUnescape(s)
	return t
}

func mustRun(ctx context.Context, actions ...chromedp.Action) {
	err := chromedp.Run(ctx, actions...)
	if err != nil {
		panic(err)
	}
}

var OnDocument func(err error, sent *network.EventRequestWillBeSent, recv *network.EventResponseReceived)

var documentMap = make(map[network.RequestID]*network.EventRequestWillBeSent)

func onTargetEvent(event interface{}) {
	if _, ok := event.(*cdproto.Message); ok {
		return
	}
	switch typed := event.(type) {
	case *network.EventRequestWillBeSent:
		if typed.Type == network.ResourceTypeDocument {
			documentMap[typed.RequestID] = typed
		}
	case *network.EventResponseReceived:
		if typed.Type == network.ResourceTypeDocument {
			if _, ok := documentMap[typed.RequestID]; ok {
				if OnDocument != nil {
					OnDocument(nil, documentMap[typed.RequestID], typed)
					delete(documentMap, typed.RequestID)
				}
			}
		}
	case *network.EventLoadingFailed:
		if typed.Type == network.ResourceTypeDocument {
			if _, ok := documentMap[typed.RequestID]; ok {
				OnDocument(errors.New(typed.ErrorText), documentMap[typed.RequestID], nil)
			}
		}
	}
}
