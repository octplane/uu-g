package uu

import (
	"fmt"
	"html/template"
	"strconv"
	"time"
)

type TimeSpan struct {
	Name     string
	duration int
	selected bool
}

func (t TimeSpan) SelectedAttribute() template.HTML {
	if t.selected {
		return template.HTML("selected")
	} else {
		return template.HTML("")
	}

}

var expiries = [...]TimeSpan{
	TimeSpan{"30 min", 1800, false},
	TimeSpan{"1 day", 86400, false},
	TimeSpan{"1 hour", 3600, false},
	TimeSpan{"1 week", 86400 * 7, true},
	TimeSpan{"1 year", 86400 * 365, false}}

func makeExpiryFromPost(expiry_key string, never bool) string {

	if never {
		return "-1"
	}
	for _, exp := range expiries {
		if expiry_key == exp.Name {
			return strconv.FormatInt(time.Now().Add(time.Duration(exp.duration)*time.Second).Unix(), 10)
		}
	}
	panic(fmt.Sprintf("Unknown duration \"%s\"", expiry_key))
}

func expiryStringFromTime(when int64) string {
	if when == -1 {
		return "never"
	}
	expire := time.Unix(when, 0)
	rest := int64(expire.Sub(time.Now()) / time.Second)
	if rest > 86400*2 {
		return fmt.Sprintf("in %d days", rest/86400)
	}
	if rest > 3600*2 {
		return fmt.Sprintf("in %d hours", rest/3600)
	}
	return fmt.Sprintf("in %d minutes", rest/60)
}
