package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-rod/rod"
	"github.com/gocolly/colly"
)

type Produto struct {
	Nome  string
	Preco string
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.bramilemcasa.com.br"),
	)

	c.UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"

	var produtos []Produto

	c.OnHTML(".vtex-search-result-3-x-galleryItem", func(e *colly.HTMLElement) {
		nomeRaw := e.ChildText(".vip-card-produto-descricao")
		precoRaw := e.ChildText("[data-cy='preco']")

		precoLimpo := strings.NewReplacer("R$", "", "\u00a0", "").Replace(precoRaw)
		precoLimpo = strings.TrimSpace(precoLimpo)

		if nomeRaw != "" {
			item := Produto{
				Nome:  strings.TrimSpace(nomeRaw),
				Preco: precoLimpo,
			}
			produtos = append(produtos, item)
			fmt.Printf("> Colly: %s | %s\n", item.Nome, item.Preco)
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("! Colly error:", err)
	})

	_ = c.Visit("https://www.bramilemcasa.com.br/mercearia")

	fmt.Printf("\n--- Colly done: %d items ---\n\n", len(produtos))

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("https://www.bramilemcasa.com.br/mercearia")
	page.MustElement(".vtex-search-result-3-x-galleryItem").MustWaitVisible()

	items := page.MustElements(".vtex-search-result-3-x-galleryItem")

	fmt.Println("--- Rod Render ---")
	for _, item := range items {
		nome, _ := item.Element(".vip-card-produto-descricao")
		preco, _ := item.Element("[data-cy='preco']")

		if nome != nil && preco != nil {
			fmt.Printf("> Rod: %s | %s\n", nome.MustText(), preco.MustText())
		}
	}
}
