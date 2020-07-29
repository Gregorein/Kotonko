package main

import (
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
	)

// fast check for substring occurence
func has(k string, l ...string) bool {
	for _, i := range l {
		if strings.Contains(k, i) {
			return true
			}
		}
	return false
	}

// fast check for string similiarity
func is(k string, l ...string) bool {
	for _, i := range l {
		if i == k {
			return true
			}
		}
	return false
	}

// turn message into BOT friendly format
func parseMessage(m string, u string) string {
	msg := strings.ToLower(strings.Replace(m, u, "username", 1))
	return msg
	}

// pad string to a fixed length with a string
func padL(s string, p string, l int) string {
	for i := 0; i < (l - len(s)); i++ {
		s = p + s
		}
		return s
}
func padR(s string, p string, l int) string {
	for i := 0; i < (l - len(s)); i++ {
		s = s + p
		}
	return s
	}

// join arrays
func join(a []string, b ...string) []string {
	for _, _b := range b {
		if !is(_b, a...) {
			a = append(a, _b)
			}
		}
		return a
	}

// check for month
func monthCheck(month string) int {
	months := []string{"sty","lut","mar","kwie","maj","czer","lipi","sier","wrze","paźd","list","grud"}
	for i, m := range months {
		if strings.Contains(month, m) {
			return i+1
			}
		}
	return 0
	}

// web
func crawl(url string) string {
	resp, _ := http.Get(url)
	bytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	s := string(bytes)

	rList := regexp.MustCompile("(?i)(list_news_type2[\\W\\w]*?href=\")([a-zA-Z/0-9]+)")
	list := rList.FindStringSubmatch(s)

	return list[2]
	}