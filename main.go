package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/osteele/liquid"
)

type Card struct {
	Prompt string `json:"prompt"`
}

// Take a full screenshot of the page and override the viewport to match the required card dimensions
func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
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

// Take a screenshot of the inputFile and save to the outputFile
func screenshot(inputFile, outputFile string) {
	// create context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
		// chromedp.WithDebugf(log.Printf),
	)
	defer cancel()

	// capture screenshot of an element
	var buf []byte

	if err := chromedp.Run(ctx, fullScreenshot(inputFile, 0, &buf)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(outputFile, buf, 0o644); err != nil {
		log.Fatal(err)
	}
}

func main() {

	// Open the cards file
	inputFile, err := os.Open("cards.json")
	if err != nil {
		fmt.Println(err)
	}
	defer inputFile.Close()

	// Read the cards file into bytes
	byteValue, _ := ioutil.ReadAll(inputFile)

	// Initialize the slice of cards that will be printed
	var cards []Card

	// Unmarshal the input file into a slice of Card objects
	json.Unmarshal(byteValue, &cards)

	// Read the card template file
	b, err := ioutil.ReadFile("card.html")
	if err != nil {
		fmt.Print(err)
	}

	// Convert the template to a string
	tmpl := string(b)

	// Initialize a new liquid templating engine
	engine := liquid.NewEngine()

	// Get the current working directory
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	// The name of the output file to screenshot in Chrome
	renderFile := fmt.Sprintf(`file://%s/render.html`, path)

	// Iterate over the slice of cards
	for i, c := range cards {

		// Create a binding for the card object
		bindings := map[string]interface{}{
			"card": c,
		}

		// Inject the bindings to the liquid engine and use the template from the card.html file
		out, err := engine.ParseAndRenderString(tmpl, bindings)
		if err != nil {
			log.Fatalln(err)
		}

		// Convert the result from the template to bytes so that it can be written to a file
		outBytes := []byte(out)

		// Create a filename for the screenshot
		fname := fmt.Sprintf("./output/render_%d.png", i)

		// Write the output from the template to a file
		ioutil.WriteFile("./render.html", outBytes, 0644)

		// Use headless Chrome to open the HTML template, take a screenshot, and save to the render file
		screenshot(renderFile, fname)
	}
}
