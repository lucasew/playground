package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
	"github.com/chromedp/chromedp/kb"
	"github.com/davecgh/go-spew/spew"
)

func MustSucess(err error) {
	if err != nil {
		panic(err)
	}
}

func EnterEnd(str string) string {
	return fmt.Sprintf("%s%s", str, kb.Enter)
}

func TypeSlowly(str string, delay time.Duration) chromedp.Action {
	return chromedp.ActionFunc(
		func(ctx context.Context) error {
			for i := range str {
				err := chromedp.Run(ctx,
					chromedp.KeyEvent(string(str[i])),
					chromedp.Sleep(delay),
				)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
}

func main() {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.DisableGPU,
		chromedp.CombinedOutput(os.Stderr),
		chromedp.UserDataDir("/tmp/chromedp"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
		func(a *chromedp.ExecAllocator) {
			spew.Dump(a)
		},
	)
	ctx, _ = chromedp.NewContext(ctx, chromedp.WithLogf(log.Printf))
	defer cancel()
	var buf string
	MustSucess(chromedp.Run(ctx,
		chromedp.Emulate(device.KindleFireHDX),
		chromedp.Navigate("https://docs.new"),
		chromedp.Sleep(1*time.Second),
		chromedp.Focus(".docs-title-input", chromedp.ByQuery),
		TypeSlowly(EnterEnd(" Boa tarde, estou funcionando?"), 10*time.Millisecond), // O espaço antes apaga o texto antigo
		TypeSlowly("Boa tarde, estou funcionando?", 10*time.Millisecond),            // O espaço antes apaga o texto antigo
		chromedp.Sleep(10*time.Minute),
	))
	fmt.Println(buf)
}
