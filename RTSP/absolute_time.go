package RTSP

import (
	"errors"
	"strconv"
	"strings"
)

type UTC_Date struct {
	YYYY int
	MM   int
	DD   int
}

type UTC_Time struct {
	HH       int
	MM       int
	SS       int
	Fraction int
}

type UTC_DateTime struct {
	Date *UTC_Date
	Time *UTC_Time
}

type UTC_Range struct {
	BeginDateTime *UTC_DateTime
	EndDateTime   *UTC_DateTime
}

func IsUTC(line string) bool {
	return strings.HasPrefix(line, "clock=")
}

func ParseUTC(line string) (err error, dateTimeRange *UTC_Range) {
	if !IsUTC(line) {
		err = errors.New("not valid absolute UTC time")
		return
	}

	strRange := strings.TrimPrefix(line, "clock=")

	times := strings.Split(strRange, "-")
	if len(times) != 2 || len(times[0]) == 0 {
		err = errors.New("not valid utc time")
		return
	}

	dateTimeRange = &UTC_Range{}
	err, dateTimeRange.BeginDateTime = parseDateTime(times[0])
	if err != nil {
		return
	}
	if len(times[1]) != 0 {
		err, dateTimeRange.EndDateTime = parseDateTime(times[1])
		if err != nil {
			return
		}
	}

	return
}

func parseDateTime(strDateTime string) (err error, dateTime *UTC_DateTime) {
	if !strings.HasSuffix(strDateTime, "Z") {
		err = errors.New("utc time need suffix Z")
		return
	}
	if len(strDateTime) < 16 {
		err = errors.New("utc need at least 16 byte")
		return
	}

	if 'T' != strDateTime[8] {
		err = errors.New("utc need T at 9")
		return
	}

	ymdAndHms := strings.Split(strDateTime, "T")
	if len(ymdAndHms) != 2 || len(ymdAndHms[0]) != 8 {
		err = errors.New("invalid utc ymd or hms")
		return
	}

	dateTime = &UTC_DateTime{}
	dateTime.Date = &UTC_Date{}
	ymd, err := strconv.Atoi(ymdAndHms[0][0:8])
	if err != nil {
		return
	}
	if ymd < 0 {
		err = errors.New("ymd must > 0")
		return
	}
	dateTime.Date.YYYY = ymd / 10000
	dateTime.Date.MM = (ymd - dateTime.Date.YYYY*10000) / 100
	dateTime.Date.DD = ymd % 100

	if len(ymdAndHms[1]) > 0 {
		strHmsFraction := strings.TrimSuffix(ymdAndHms[1], "Z")
		strHmsFractionArr := strings.Split(strHmsFraction, ".")
		strHms := strHmsFractionArr[0]
		if len(strHms) != 6 {
			err = errors.New("hh mm  ss must 6 byte")
			return
		}

		if len(strHmsFractionArr) > 2 {
			err = errors.New("hh mm  ss can only 1 dot")
			return
		}

		dateTime.Time = &UTC_Time{}
		hms := 0
		hms, err = strconv.Atoi(strHms)
		if err != nil {
			return
		}

		dateTime.Time.HH = hms / 10000
		dateTime.Time.MM = (hms - dateTime.Time.HH*10000) / 100
		dateTime.Time.SS = hms % 100

		if dateTime.Time.HH < 0 || dateTime.Time.HH > 23 {
			err = errors.New("invalid date time hh ,need 0-23")
			return
		}

		if dateTime.Time.MM < 0 || dateTime.Time.MM > 59 {
			err = errors.New("invalid date time mm ,need 0-59")
			return
		}

		if dateTime.Time.SS < 0 || dateTime.Time.SS > 59 {
			err = errors.New("invalid date time mm ,need 0-59")
			return
		}

		if len(strHmsFractionArr) == 2 && len(strHmsFractionArr[1]) > 0 {
			dateTime.Time.Fraction, err = strconv.Atoi(strHmsFractionArr[1])
			if err != nil {
				return
			}
		}

	}

	return
}
