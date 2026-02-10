package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const Layout = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	start, err := time.Parse(Layout, dstart)
	if err != nil {
		return "", errors.New("Invalid date format")
	}

	if repeat == "" {
		return "", nil
	}

	parts := strings.Split(repeat, " ")
	rule := parts[0]

	afterNow := func(date time.Time, now time.Time) bool {
		return date.Format(Layout) > now.Format(Layout)
	}

	switch rule {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("Invalid format: d <days>")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("Invalid interval for d: must be 1-400")
		}
		date := start.AddDate(0, 0, days)
		for !afterNow(date, now) {
			date = date.AddDate(0, 0, days)
		}
		return date.Format(Layout), nil

	case "y":
		if len(parts) != 1 {
			return "", errors.New("Invalid format: y <years>")
		}
		date := start.AddDate(1, 0, 0)
		for !afterNow(date, now) {
			date = date.AddDate(1, 0, 0)
		}
		return date.Format(Layout), nil

	case "w":
		if len(parts) != 2 {
			return "", errors.New("Invalid format: w <weekdays>")
		}
		dayStrs := strings.Split(parts[1], ",")
		targetDays := make(map[int]bool)
		for _, s := range dayStrs {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 7 {
				return "", errors.New("Invalid interval for w: must be 1-7")
			}
			targetDays[n] = true
		}

		date := start
		for {
			date = date.AddDate(0, 0, 1)
			weekday := int(date.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			if targetDays[weekday] && afterNow(date, now) {
				return date.Format(Layout), nil
			}
		}

	case "m":
		if len(parts) < 2 || len(parts) > 3 {
			return "", errors.New("Invalid format: m <months>")
		}

		dayStrs := strings.Split(parts[1], ",")
		var targetDays []int
		for _, s := range dayStrs {
			n, err := strconv.Atoi(s)
			if err != nil || n < -31 || n > 31 || n == 0 {
				return "", errors.New("Invalid day for m: must be -31..-1 or 1-31")
			}
			targetDays = append(targetDays, n)
		}

		// Подсчитываем количество отрицательных дней и проверяем наличие -1
		negativeCount := 0
		hasMinusOne := false
		for _, d := range targetDays {
			if d < 0 {
				negativeCount++
				if d == -1 {
					hasMinusOne = true
				}
			}
		}

		// Запрещаем: если больше одного отрицательного дня и при этом нет -1
		if negativeCount > 1 && !hasMinusOne {
			return "", errors.New("Multiple negative days must include -1")
		}

		var targetMonths []int
		if len(parts) == 3 {
			monthStrs := strings.Split(parts[2], ",")
			for _, s := range monthStrs {
				n, err := strconv.Atoi(s)
				if err != nil || n < 1 || n > 12 {
					return "", errors.New("Invalid interval for m: must be 1-12")
				}
				targetMonths = append(targetMonths, n)
			}
		} else {
			for i := 1; i <= 12; i++ {
				targetMonths = append(targetMonths, i)
			}
		}

		date := start
		maxIter := 365 * 5
		for i := 0; i < maxIter; i++ {
			year, month, day := date.Date()
			monthInt := int(month)

			inMonth := false
			for _, m := range targetMonths {
				if m == monthInt {
					inMonth = true
					break
				}
			}
			if !inMonth {
				date = date.AddDate(0, 0, 1)
				continue
			}

			lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()

			match := false
			for _, d := range targetDays {
				if d > 0 && d == day {
					match = true
					break
				}
				if d < 0 {
					expected := lastDay + d + 1
					if day == expected {
						match = true
						break
					}
				}
			}
			if match && afterNow(date, now) {
				return date.Format(Layout), nil
			}
			date = date.AddDate(0, 0, 1)
		}

		return "", errors.New("No valid date found within 5 years")

	default:
		return "", errors.New("Unsupported repeat rule")
	}
}

func nextDayHandler(w http.ResponseWriter, r *http.Request) {
	nowParam := r.FormValue("now")
	dateParam := r.FormValue("date")
	repeatParam := r.FormValue("repeat")

	now := time.Now()
	if nowParam != "" {
		var err error
		now, err = time.Parse(Layout, nowParam)
		if err != nil {
			http.Error(w, "Invalid now date format", http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(now, dateParam, repeatParam)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	if nextDate == "" {
		w.Write([]byte(""))
	} else {
		w.Write([]byte(nextDate))
	}
}
