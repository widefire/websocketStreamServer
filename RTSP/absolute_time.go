package rtsp

import (
	"errors"
	"strconv"
	"strings"
)

//UTCDate ...
type UTCDate struct {
	YYYY int
	MM   int
	DD   int
}

//UTCTime ...
type UTCTime struct {
	HH       int
	MM       int
	SS       int
	Fraction int
}

//UTCDateTime ...
type UTCDateTime struct {
	Date *UTCDate
	Time *UTCTime
}

//UTCRange ...
type UTCRange struct {
	BeginDateTime *UTCDateTime
	EndDateTime   *UTCDateTime
}

//IsUTC ...
func IsUTC(line string) bool {
	return strings.HasPrefix(line, "clock=")
}

//ParseUTC ...
func ParseUTC(line string) (dateTimeRange *UTCRange, err error) {
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

	dateTimeRange = &UTCRange{}
	dateTimeRange.BeginDateTime, err = parseDateTime(times[0])
	if err != nil {
		return
	}
	if len(times[1]) != 0 {
		dateTimeRange.EndDateTime, err = parseDateTime(times[1])
		if err != nil {
			return
		}
	}

	return
}

func parseDateTime(strDateTime string) (dateTime *UTCDateTime, err error) {
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

	dateTime = &UTCDateTime{}
	dateTime.Date = &UTCDate{}
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

		dateTime.Time = &UTCTime{}
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
