package xtemplate

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"path/filepath"
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
			return w.Write([]byte(strings.TrimSpace(retv)))
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

// MixAsset reads a laravel-mix mix-manifest.json file
// and returns the hashed filename.
// assumes that the file will be in ./static
func MixAsset(publicPath string) func(val string) string {
	return func(val string) string {

		manifest := filepath.Join(publicPath, "mix-manifest.json")
		content, err := ioutil.ReadFile(manifest)
		if err != nil {
			return fmt.Sprintf("err-cant-read-mix-manifest")
		}

		data := map[string]string{}
		if err := json.Unmarshal(content, &data); err != nil {
			return fmt.Sprintf("err-cant-unmarshal-mix-manifest")
		}

		retv, found := data[val]
		if !found {
			return val
		}

		return retv
	}
}

func NoCache(file string) string {
	return fmt.Sprint(file, "?t=", time.Now().UnixNano())
}

func IsEmpty(val interface{}) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	if v.Type().Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}

		v = reflect.Indirect(v)
	}

	if v.IsZero() {
		return true
	}

	return false
}

// IfEmpty if val is empty return the first non-empty item in values
func IfEmpty(val interface{}, values ...interface{}) interface{} {

	if IsEmpty(val) {
		for _, v := range values {
			if !IsEmpty(v) {
				return v
			}

			return values[0]
		}
	}

	return val
}
