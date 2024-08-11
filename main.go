package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type SearchResp struct {
	Places        []Place `json:"places"`
	NextPageToken string  `json:"nextPageToken"`
}

type Text struct {
	Text string `json:"text"`
}

type Place struct {
	Name    string  `json:"name"`
	Address string  `json:"formattedAddress"`
	Rating  float64 `json:"rating"`

	GoogleMapsURI  string `json:"googleMapsUri"`
	WebsiteURI     string `json:"websiteUri"`
	BusinessStatus string `json:"businessStatus"`
	PrimaryType    string `json:"primaryType"`
	DisplayName    Text   `json:"displayName"`
	Summary        Text   `json:"editorialSummary"`

	Takeout         bool `json:"takeout"`
	Delivery        bool `json:"delivery"`
	DineIn          bool `json:"dineIn"`
	Reservable      bool `json:"reservable"`
	ServesBreakfast bool `json:"servesBreakfast"`
	ServesLunch     bool `json:"servesLunch"`
	ServesDinner    bool `json:"servesDinner"`
	ServesBrunch    bool `json:"servesBrunch"`
	AllowsDogs      bool `json:"allowsDogs"`
}

func main() {
	ctx := context.Background()
	client := &http.Client{}
	first := true
	count := 1

	//sheetsService, err := sheets.NewService(ctx)
	//if err != nil {
	//	panic(err)
	//}
	f, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(f)

	sr := SearchResp{}
	for sr.NextPageToken != "" || first {
		first = false
		target := "https://places.googleapis.com/v1/places:searchText"
		body := getReqBody(sr.NextPageToken)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, body)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("X-Goog-FieldMask", "*")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Goog-Api-Key", os.Args[1])

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		//os.WriteFile("resp.json", b, 0644)
		/*b, err := os.ReadFile("resp.json")
		if err != nil {
			panic(err)
		}*/
		sr = SearchResp{}
		json.Unmarshal(b, &sr)
		fmt.Println("okay")

		for _, p := range sr.Places {
			fmt.Printf("%d %#v\n", count, p)
			// =HYPERLINK("https://maps.google.com/?cid=17772397080046998457", "test")
			record := []string{
				linkIf(p.WebsiteURI, p.DisplayName.Text),
				linkIf(p.GoogleMapsURI, p.Address),
				fmt.Sprintf("%f", p.Rating),
				serves(p),
				p.Summary.Text,
			}
			w.Write(record)
			count++
		}
		fmt.Println("next:", sr.NextPageToken)
		defer resp.Body.Close()
	}
	w.Flush()
	f.Close()
}

func serves(p Place) string {
	breakfast := "🥞"
	lunch := "🍔"
	dinner := "🍽️"
	brunch := "🍊"
	dogs := "🦮"

	serves := ""
	if p.ServesBreakfast {
		serves += breakfast
	}
	if p.ServesBrunch {
		serves += brunch
	}
	if p.ServesLunch {
		serves += lunch
	}
	if p.ServesDinner {
		serves += dinner
	}
	if p.AllowsDogs {
		serves += dogs
	}
	return serves
}

func linkIf(url, text string) string {
	if url != "" {
		text = "=HYPERLINK(\"" + url + "\", \"" + text + "\")"
	}
	return text
}

func getReqBody(page string) *bytes.Buffer {
	if page != "" {
		return bytes.NewBuffer([]byte(`{
  "textQuery" : "restaurants in Tarrytown, NY",
  "pageToken": "` + page + `"
}`))

	} else {
		return bytes.NewBuffer([]byte(`{
  "textQuery" : "restaurants in Tarrytown, NY"
}`))
	}
}
