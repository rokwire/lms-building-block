// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	// HoursInDay is the number of hours in one day
	HoursInDay int = 24
	// SecondsInDay is the number of seconds in one 24-hour day
	SecondsInDay int = HoursInDay * SecondsInHour
	// SecondsInHour is the number of seconds in one hour
	SecondsInHour int = 60 * SecondsInMinute
	// SecondsInMinute is the number of seconds in one minute
	SecondsInMinute int = 60
	// MinTZOffset is the minimum allowed timezone offset from UTC in seconds (UTC-12 = -43200)
	MinTZOffset int = -12 * SecondsInHour
	// MaxTZOffset is the maximum allowed timezone offset from UTC in seconds (UTC+14 = 50400)
	MaxTZOffset int = 14 * SecondsInHour
)

// Filter represents find filter for finding entities by the their fields
type Filter struct {
	Items []FilterItem
}

// FilterItem represents find filter pair - field/value
type FilterItem struct {
	Field string
	Value []string
}

// ConstructFilter constructs Filter from the http request params
func ConstructFilter(r *http.Request) *Filter {
	values := r.URL.Query()
	if len(values) == 0 {
		return nil
	}

	var filter Filter
	var items []FilterItem
	for k, v := range values {
		if len(v) > 0 {
			items = append(items, FilterItem{Field: k, Value: v})
		}
	}
	filter.Items = items
	return &filter
}

// ModifyHTMLContent removes all not web href links. It also remove web links which points to pdf document
// For example:
// <a href="mailto:email@abc.abc">email@abc.abc</a> -> email@abc.abc
// <a href="ftp://server/file">Some text</a> -> Some text
// <a href="tel:1234">1234</a> -> 1234
//
// <a href="https://humanresources.illinois.edu/assets/docs/COVID-19-Pay-Continuation-Protocol-Final-3-22-2020.pdf">the university's pay continuation protocol</a> ->
// the university's pay continuation protocol(https://humanresources.illinois.edu/assets/docs/COVID-19-Pay-Continuation-Protocol-Final-3-22-2020.pdf)
func ModifyHTMLContent(input string) string {
	reader := strings.NewReader(input)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Printf("error creating reader from the html string - %s\n", err)
		//there is no what to do so return the input
		return input
	}

	//process
	doc.Find("a").Each(func(_ int, link *goquery.Selection) {
		text := strings.TrimSpace(link.Text())
		href, ok := link.Attr("href")
		if ok && len(href) > 0 {

			splitHref := strings.Split(href, ":")
			if len(splitHref) > 0 {
				protocol := splitHref[0]

				if protocol == "http" || protocol == "https" {
					//it is a web protocol, so we just need to look for .pdf resources
					if strings.HasSuffix(href, ".pdf") {
						log.Printf("modifying.. href - %s\ttext - %s\n", href, text)
						link.ReplaceWithHtml(text + "(" + href + ")")
					}
				} else {
					//it is not а web protocol, so here we need to apply modifications

					log.Printf("modifying.. href - %s\ttext - %s\n", href, text)
					link.ReplaceWithHtml(text)
				}
			}

		}
	})

	body := doc.Find("body")
	if body == nil {
		log.Printf("body is nil for some reasons - %s\n", input)
		//there is no what to do so return the input
		return input
	}
	final, err := body.Html()
	if err != nil {
		log.Printf("error getting html from body - %s\n", err)
		//there is no what to do so return the input
		return input
	}
	return final
}

// LogRequest logs the request as hide some header fields because of security reasons
func LogRequest(req *http.Request) {
	if req == nil {
		return
	}

	method := req.Method
	path := req.URL.Path

	val, ok := req.Header["User-Agent"]
	if ok && len(val) != 0 && val[0] == "ELB-HealthChecker/2.0" {
		return
	}

	header := make(map[string][]string)
	for key, value := range req.Header {
		var logValue []string
		//do not log api keys, cookies and Authorization
		if key == "Rokwire-Api-Key" || key == "User-Id" || key == "Cookie" ||
			key == "Authorization" || key == "Rokwire-Hs-Api-Key" || key == "Group" ||
			key == "Rokwire-Acc-Id" || key == "Csrf" {
			logValue = append(logValue, "---")
		} else {
			logValue = value
		}
		header[key] = logValue
	}
	log.Printf("%s %s %s", method, path, header)
}

// GetLogUUIDValue prepares UUID to be logged.
func GetLogUUIDValue(identifier string) string {
	if len(identifier) < 26 {
		return fmt.Sprintf("bad identifier - %s", identifier)
	}

	sub := identifier[:26]
	return fmt.Sprintf("%s***", sub)
}

// GetLogValue prepares a sensitive data to be logged.
func GetLogValue(value string) string {
	if len(value) <= 3 {
		return "***"
	}
	last3 := value[len(value)-3:]
	return fmt.Sprintf("***%s", last3)
}

// DeepEqual checks whether a and b are “deeply equal”
func DeepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// Equal compares two slices
func Equal(a, b []string, strict bool) bool {
	if !strict {
		sort.Strings(a)
		sort.Strings(b)
	}

	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// EqualPointers compares two pointers slices
func EqualPointers(a, b *[]string) bool {
	if a == nil && b == nil {
		return true //equals
	}
	if a != nil && b == nil {
		return false // not equals
	}
	if a == nil && b != nil {
		return false // not equals
	}

	//both are not nil
	return Equal(*a, *b, true)
}

// GetInt gives the value which this pointer points. Gives 0 if the pointer is nil
func GetInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// GetString gives the value which this pointer points. Gives empty string if the pointer is nil
func GetString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// GetTime gives the value which this pointer points. Gives empty string if the pointer is nil
func GetTime(time *time.Time) string {
	if time == nil {
		return ""
	}
	return fmt.Sprintf("%s", time)
}

// SortVersions sorts the versions list. The format is x.x.x or x.x which is the short for x.x.0
func SortVersions(versions []string) {
	//sort
	sort.Slice(versions, func(i, j int) bool {
		v1 := versions[i]
		v2 := versions[j]
		return !IsVersionLess(v1, v2)
	})
}

// IsVersionLess checks if v1 is less than v2. The format is x.x.x or x.x which is the short for x.x.0
func IsVersionLess(v1 string, v2 string) bool {
	var v1Major, v1Minor, v1Patch int
	var v2Major, v2Minor, v2Patch int

	v1Elements := strings.Split(v1, ".")
	v2Elements := strings.Split(v2, ".")

	v1Major, _ = strconv.Atoi(v1Elements[0])
	v1Minor, _ = strconv.Atoi(v1Elements[1])
	if len(v1Elements) == 2 {
		v1Patch = 0
	} else {
		v1Patch, _ = strconv.Atoi(v1Elements[2])
	}

	v2Major, _ = strconv.Atoi(v2Elements[0])
	v2Minor, _ = strconv.Atoi(v2Elements[1])
	if len(v2Elements) == 2 {
		v2Patch = 0
	} else {
		v2Patch, _ = strconv.Atoi(v2Elements[2])
	}

	//1 first check major
	if v1Major < v2Major {
		return true
	}
	if v1Major > v2Major {
		return false
	}

	//2. majors are equals so check minors
	if v1Minor < v2Minor {
		return true
	}
	if v1Minor > v2Minor {
		return false
	}

	//3. minors are equals so check patch
	if v1Patch < v2Patch {
		return true
	}
	if v1Patch > v2Patch {
		return false
	}

	// they are equals
	return false
}

// Exist checks if the items exists in the list
func Exist[T listExistType](list []T, value T) bool {
	if len(list) == 0 {
		return false
	}

	for _, s := range list {
		if value == s {
			return true
		}
	}
	return false
}

type listExistType interface {
	string | int
}

// Hash hashes the s value
func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// DateEqual checks if date1 is the same as date2
func DateEqual(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// AnyToFloat64 Converts to float64
func AnyToFloat64(val any) float64 {
	switch i := val.(type) {
	case float64:
		return i
	case float32:
		return float64(i)
	case int64:
		return float64(i)
	case int32:
		return float64(i)
	default:
		return math.NaN()
	}
}

// AnyToArrayOfInt Converts to list of integers
func AnyToArrayOfInt(val any) []int {
	var result []int
	switch reflect.TypeOf(val).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(val)
		for i := 0; i < s.Len(); i++ {
			iVal := s.Index(i)
			intVal := iVal.Interface()
			switch intVal.(type) {
			case int:
				result = append(result, intVal.(int))
				break
			case int32:
				result = append(result, int(intVal.(int32)))
				break
			case int64:
				result = append(result, int(intVal.(int64)))
				break
			case float32:
				result = append(result, int(intVal.(float64)))
				break
			case float64:
				result = append(result, int(intVal.(float64)))
				break
			}

		}
		return result
	}
	return result
}

// GetValue returns the value corresponding to key in items; returns an error if missing and required or not the expected type
func GetValue[T any](items map[string]interface{}, key string, required bool) (T, error) {
	mapValue, ok := items[key]
	if required && !ok {
		return *new(T), errors.ErrorData(logutils.StatusMissing, "map value", &logutils.FieldArgs{"key": key, "required": required})
	}

	value, ok := mapValue.(T)
	if !ok {
		return *new(T), errors.ErrorData(logutils.StatusInvalid, "map value", &logutils.FieldArgs{"key": key, "required": required})
	}

	return value, nil
}

// StartTimer starts a timer with the given name, period, and function to call when the timer goes off
func StartTimer(timer *time.Timer, timerDone chan bool, initialDuration *time.Duration, period time.Duration, periodicFunc func(), name string, logger *logs.Logger) {
	if logger != nil {
		logger.Info("start timer for " + name)
	}

	//cancel if active
	if timer != nil {
		if logger != nil {
			logger.Info(name + " -> there is active timer, so cancel it")
		}

		timerDone <- true
		timer.Stop()
	}

	onTimer(timer, timerDone, initialDuration, period, periodicFunc, name, logger)
}

func onTimer(timer *time.Timer, timerDone chan bool, initialDuration *time.Duration, period time.Duration, periodicFunc func(), name string, logger *logs.Logger) {
	hasLogger := (logger != nil)
	if hasLogger {
		logger.Info(name)
	}

	duration := period
	if initialDuration != nil {
		duration = *initialDuration
	} else {
		periodicFunc()
	}
	timer = time.NewTimer(duration)

	if hasLogger {
		logger.Infof(name+" -> next call after %s", duration)
	}

	select {
	case <-timer.C:
		// timer expired
		if hasLogger {
			logger.Info(name + " -> timer expired")
		}
		timer = nil

		onTimer(timer, timerDone, nil, period, periodicFunc, name, logger)
	case <-timerDone:
		// timer aborted
		if hasLogger {
			logger.Info(name + " -> timer aborted")
		}
		timer = nil
	}
}
