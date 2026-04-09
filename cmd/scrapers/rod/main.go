package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"github.com/gocolly/colly/v2"
)

type Product struct {
	ProductName string `json:"productName"`
	Items       []struct {
		Sellers []struct {
			CommertialOffer struct {
				Price float64 `json:"Price"`
			} `json:"commertialOffer"`
		} `json:"sellers"`
	} `json:"items"`
}

func main() {
	path, _ := launcher.LookPath()
	l := launcher.New().Bin(path).Headless(false).Set("disable-blink-features", "AutomationControlled")
	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	defer browser.MustClose()

	page := stealth.MustPage(browser)

	fmt.Println("> Target: Bramil Areal/Posse")
	page.MustNavigate("https://www.bramilemcasa.com.br/")

	_ = rod.Try(func() {
		page.Timeout(4 * time.Second).MustElement("#onetrust-accept-btn-handler").MustClick()
	})

	inputCEP := page.MustElement("input[formcontrolname='cep']")
	inputCEP.MustSelectAllText().MustInput("25770460")
	page.MustElementR("button", "/verificar/i").MustClick()

	time.Sleep(3 * time.Second)
	err := rod.Try(func() {
		addr := page.MustElementR("div, p, span", "/25770460/i")
		addr.MustWaitVisible().MustClick()

		time.Sleep(1 * time.Second)
		page.MustElementR("button", "/receber em casa/i").MustClick()
	})

	if err != nil {
		fmt.Println("! Modal flow failed, bypass via URL")
	}

	_ = rod.Try(func() { page.Keyboard.MustType(input.Escape) })

	wait := page.MustWaitNavigation()
	page.MustNavigate("https://www.bramilemcasa.com.br/cerveja")
	wait()
	time.Sleep(2 * time.Second)

	cookies := make([]*http.Cookie, 0)
	for _, c := range page.MustCookies() {
		cookies = append(cookies, &http.Cookie{
			Name: c.Name, Value: c.Value, Domain: c.Domain, Path: c.Path,
		})
	}

	fmt.Printf("✔ Session: %d cookies captured\n", len(cookies))

	collector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"),
	)

	u, _ := url.Parse("https://www.bramilemcasa.com.br/")
	collector.SetCookies(u.String(), cookies)

	apiUrl := "https://www.bramilemcasa.com.br/api/catalog_system/pub/products/search?ft=cerveja&_from=0&_to=15"

	collector.OnResponse(func(r *colly.Response) {
		var data []Product
		if err := json.Unmarshal(r.Body, &data); err != nil {
			fmt.Println("! JSON Parse error")
			return
		}

		fmt.Println("\n--- MARKET TRACKER ---")
		for _, p := range data {
			if len(p.Items) > 0 && len(p.Items[0].Sellers) > 0 {
				price := p.Items[0].Sellers[0].CommertialOffer.Price
				fmt.Printf("R$ %06.2f | %s\n", price, p.ProductName)
			}
		}
	})

	collector.Visit(apiUrl)
}
