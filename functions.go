package xtemplate

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/valyala/fasttemplate"
)

// marshalJSON marshals val to json
func marshalJSON(val interface{}) template.JS {
	retv := []byte{}
	retv, _ = json.Marshal(val)

	return template.JS(retv)
}

func capitalize(val string) string {
	return strings.ToUpper(val[:1]) + val[1:]
}

func lower(val string) string {
	return strings.ToLower(val)
}

func upper(val string) string {
	return strings.ToUpper(val)
}

func getDefault(val, defa interface{}) interface{} {
	v := reflect.ValueOf(val)
	if v.IsZero() || v.IsNil() {
		return defa
	}

	return val
}

func args(val ...interface{}) interface{} {
	return val
}

func kwargs(arg ...interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	for i := 0; i < len(arg); i += 2 {
		m[fmt.Sprintf("%v", arg[i])] = arg[i+1]
	}

	return m
}

// tags a default/reference implementation which supports the <tag></tag> feature
func tags(typ string, attr map[string]interface{}, content string) template.HTML {

	tpl := "<[tag] [attr]>[content]</[tag]>"
	t := fasttemplate.New(tpl, "[", "]")
	s := t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		switch tag {
		case "tag":
			return w.Write([]byte(typ))
		case "attr":
			retv := ""
			for k, v := range attr {
				retv = fmt.Sprintf("%s %s=\"%v\"", retv, k, v)
			}
			return w.Write([]byte(retv))
		default:
			return w.Write([]byte(content))
		}
	})

	return template.HTML(s)
}

// formatDate formate time.Time using pkg time layout strings
func formatDate(dt time.Time, layout string) string {
	return dt.Format(layout)
}

// formatCDate format a time.Time value
// adapted from https://github.com/tyler-sommer/stick/blob/a6b3e7c8738498d203a59d5f5b99c6019e212a4b/twig/filter/filter.go#L127
func formatCDate(dt time.Time, format string) string {

	// build a golang date string
	table := map[string]string{
		"d": "02",
		"D": "Mon",
		"j": "2",
		"l": "Monday",
		"N": "", // TODO: ISO-8601 numeric representation of the day of the week (added in PHP 5.1.0)
		"S": "", // TODO: English ordinal suffix for the day of the month, 2 characters
		"w": "", // TODO: Numeric representation of the day of the week
		"z": "", // TODO: The day of the year (starting from 0)
		"W": "", // TODO: ISO-8601 week number of year, weeks starting on Monday (added in PHP 4.1.0)
		"F": "January",
		"m": "01",
		"M": "Jan",
		"n": "1",
		"t": "", // TODO: Number of days in the given month
		"L": "", // TODO: Whether it's a leap year
		"o": "", // TODO: ISO-8601 year number. This has the same value as Y, except that if the ISO week number (W) belongs to the previous or next year, that year is used instead. (added in PHP 5.1.0)
		"Y": "2006",
		"y": "06",
		"a": "pm",
		"A": "PM",
		"B": "", // TODO: Swatch Internet time (is this even still a thing?!)
		"g": "3",
		"G": "15",
		"h": "03",
		"H": "15",
		"i": "04",
		"s": "05",
		"u": "000000",
		"e": "", // TODO: Timezone identifier (added in PHP 5.1.0)
		"I": "", // TODO: Whether or not the date is in daylight saving time
		"O": "-0700",
		"P": "-07:00",
		"T": "MST",
		"c": "2006-01-02T15:04:05-07:00",
		"r": "Mon, 02 Jan 2006 15:04:05 -0700",
		"U": "", // TODO: Seconds since the Unix Epoch (January 1 1970 00:00:00 GMT)
	}
	var layout string

	maxLen := len(format)
	for i := 0; i < maxLen; i++ {
		char := string(format[i])
		if t, ok := table[char]; ok {
			layout += t
			continue
		}
		if "\\" == char && i < maxLen-1 {
			layout += string(format[i+1])
			continue
		}
		layout += char
	}

	return dt.Format(layout)
}

func randString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
