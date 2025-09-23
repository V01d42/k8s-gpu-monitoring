package timeutil

import "time"

// NowJST returns the current time in JST as "YYYY/MM/DD HH:MM:SS".
func NowJST() string {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}

	return time.Now().In(loc).Format("2006/01/02 15:04:05")
}
