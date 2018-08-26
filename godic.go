package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const ACCESS_TOKEN = "YOUR_ACCESS_TOKEN"
const JE_URL = "https://api.codic.jp/v1/engine/translate.json"
const EJ_URL = "https://api.codic.jp/v1/ced/lookup.json"

type TransResp []struct {
	Successful     bool   `json:"successful"`
	Text           string `json:"text"`
	TranslatedText string `json:"translated_text"`
	Words          []struct {
		Successful     bool       `json:"successful"`
		Text           string     `json:"text"`
		TranslatedText string     `json:"translated_text"`
		Candidates     Candidates `json:"candidates"`
	} `json:"words"`
}

type Candidates []struct {
	Text string `json:"text"`
}

type Item struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Digest string `json:"digest"`
}

func getValues(source string) url.Values {
	values := url.Values{}
	values.Add("text", source)
	return values
}

func jaToEn(source string) (TransResp, error) {
	var ts TransResp

	values := url.Values{}
	values.Add("text", source)
	url := JE_URL + "?" + values.Encode()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ts, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ACCESS_TOKEN))

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return ts, err
	}
	defer res.Body.Close()

	byteArray, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return ts, err
	}

	err = json.Unmarshal(byteArray, &ts)
	if err != nil {
		fmt.Println(err)
	}

	return ts, nil
}

func enToja(source string) ([]Item, error) {
	var items []Item

	values := url.Values{}
	values.Add("query", source)
	values.Add("count", "1")
	url := EJ_URL + "?" + values.Encode()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return items, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ACCESS_TOKEN))

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		return items, err
	}
	defer res.Body.Close()

	byteArray, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return items, err
	}

	err = json.Unmarshal(byteArray, &items)
	if err != nil {
		return items, err
	}

	return items, nil
}

func fetchSynonym(synonymWords Candidates) ([]*Item, int, error) {
	ret := make([]*Item, len(synonymWords))
	longestLen := 0
	for i, s := range synonymWords {
		items, err := enToja(s.Text)
		if err != nil {
			return nil, longestLen, err
		}
		ret[i] = &items[0]
		if len(items[0].Title) > longestLen {
			longestLen = len(items[0].Title)
		}
	}
	return ret, longestLen, nil
}

func translate(source string) error {
	ts, err := jaToEn(source)
	if err != nil {
		return err
	}
	if len(ts) == 0 {
		return fmt.Errorf("Translation Failed")
	}

	if opt == "n" {
		fmt.Printf("%s", ts[0].TranslatedText)

	} else if opt == "s" {
		synos, length, err := fetchSynonym(ts[0].Words[0].Candidates)
		if err != nil {
			return err
		}
		for _, s := range synos {
			fmt.Printf("%-*s: %s\n", length, s.Title, s.Digest)
		}
	}
	return nil
}

const usage = `godic is cli of codic

Command:
	n	: Translate to variable name or function name
	s	: List synonym

Example:
	godic n 存在するか
	godic s 取得する

Learn more
	https://codic.jp/engine
`

var opt string

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		fmt.Println(usage)
		return
	}

	opt = flag.Args()[0]
	if opt != "n" && opt != "s" {
		fmt.Println(usage)
		return
	}
	source := flag.Args()[1]

	err := translate(source)
	if err != nil {
		fmt.Println(err)
	}
}
