package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/osteele/liquid"
)

type Card struct {
	Prompt string `json:"prompt"`
}

func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			/*
				_, _, cssContentSize, err := page.GetLayoutMetrics().Do(ctx)
				if err != nil {
					return err
				}
			*/

			//width, height := int64(math.Ceil(cssContentSize.Width)), int64(math.Ceil(cssContentSize.Height))

			// force viewport emulation
			err := emulation.SetDeviceMetricsOverride(597, 1122, 0, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  597,
					Height: 1122,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible),
	}
}

func screenshot(fname string) {
	// create context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// capture screenshot of an element
	var buf []byte

	if err := chromedp.Run(ctx, fullScreenshot(`file:///home/csthompson/workspace/go/src/github.com/csthompson/cardbuilder/render.html`, 90, &buf)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(fname, buf, 0o644); err != nil {
		log.Fatal(err)
	}
}

func main() {

	//Read the cards file
	jsonFile, err := os.Open("cards.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var cards []Card

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	json.Unmarshal(byteValue, &cards)

	//Read the template file
	b, err := ioutil.ReadFile("card.html") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	tmpl := string(b) // convert content to a 'string'

	engine := liquid.NewEngine()

	fmt.Println(cards)

	for i, c := range cards {

		bindings := map[string]interface{}{
			"card": c,
		}
		out, err := engine.ParseAndRenderString(tmpl, bindings)
		if err != nil {
			log.Fatalln(err)
		}
		d1 := []byte(out)
		fname := "./output/render_" + strconv.Itoa(i) + ".png"
		fmt.Println(fname)
		ioutil.WriteFile("./render.html", d1, 0644)
		screenshot(fname)
	}
}
