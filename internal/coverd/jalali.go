package coverd

import (
	"fmt"
	"time"
)

// Jalali conversion: faithful port of the jalaali-js algorithm (Borkowski's
// astronomical approximation with the standard breaks table). This matches
// the official Iranian Solar Hijri calendar for modern years — a naive
// 33-year arithmetic approximation drifts by a day around Nowruz.

type jalaliBreaks = []int

var breaks = jalaliBreaks{
	-61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210,
	1635, 2060, 2097, 2192, 2262, 2324, 2394, 2456, 3178,
}

func jalDiv(a, b int) int { return a / b }
func jalMod(a, b int) int { return a - (a/b)*b }

type jalCalResult struct {
	leap  int
	gy    int
	march int
}

func jalCal(jy int) jalCalResult {
	bl := len(breaks)
	gy := jy + 621
	leapJ := -14
	jp := breaks[0]
	jm := 0
	jump := 0

	if jy < jp || jy >= breaks[bl-1] {
		// Out of the supported range; clamp behavior is acceptable for the
		// cover site's modern dates, but fail loudly to catch bugs.
		panic("coverd: jalali year out of supported range")
	}

	for i := 1; i < bl; i++ {
		jm = breaks[i]
		jump = jm - jp
		if jy < jm {
			break
		}
		leapJ += jalDiv(jump, 33)*8 + jalDiv(jalMod(jump, 33), 4)
		jp = jm
	}
	n := jy - jp

	leapJ += jalDiv(n, 33)*8 + jalDiv(jalMod(n, 33)+3, 4)
	if jalMod(jump, 33) == 4 && jump-n == 4 {
		leapJ++
	}

	leapG := jalDiv(gy, 4) - jalDiv((jalDiv(gy, 100)+1)*3, 4) - 150
	march := 20 + leapJ - leapG

	if jump-n < 6 {
		n = n - jump + jalDiv(jump+4, 33)*33
	}
	leap := jalMod(jalMod(n+1, 33)-1, 4)
	if leap == -1 {
		leap = 4
	}
	return jalCalResult{leap: leap, gy: gy, march: march}
}

// g2d converts a Gregorian date to a Julian Day Number (jalaali-js).
func g2d(gy, gm, gd int) int {
	d := jalDiv((gy+jalDiv(gm-8, 6)+100100)*1461, 4) +
		jalDiv(153*jalMod(gm+9, 12)+2, 5) +
		gd - 34840408
	d = d - jalDiv(jalDiv(gy+100100+jalDiv(gm-8, 6), 100)*3, 4) + 752
	return d
}

// d2g converts a Julian Day Number back to a Gregorian date (jalaali-js).
func d2g(jdn int) (gy, gm, gd int) {
	j := 4*jdn + 139361631
	j += jalDiv(jalDiv(4*jdn+183187720, 146097)*3, 4)*4 - 3908
	i := jalDiv(jalMod(j, 1461), 4)*5 + 308
	gd = jalDiv(jalMod(i, 153), 5) + 1
	gm = jalMod(jalDiv(i, 153), 12) + 1
	gy = jalDiv(j, 1461) - 100100 + jalDiv(8-gm, 6)
	return gy, gm, gd
}

// j2d converts a Jalali date to a Julian Day Number (jalaali-js).
func j2d(jy, jm, jd int) int {
	r := jalCal(jy)
	return g2d(r.gy, 3, r.march) + (jm-1)*31 - jalDiv(jm, 7)*(jm-7) + jd - 1
}

// d2j converts a Julian Day Number to a Jalali date (jalaali-js).
func d2j(jdn int) (jy, jm, jd int) {
	gy, _, _ := d2g(jdn)
	jy = gy - 621
	r := jalCal(jy)
	jdn1f := g2d(r.gy, 3, r.march)

	k := jdn - jdn1f
	if k >= 0 {
		if k <= 185 {
			jm = 1 + jalDiv(k, 31)
			jd = jalMod(k, 31) + 1
			return jy, jm, jd
		}
		k -= 186
	} else {
		jy--
		k += 179
		if r.leap == 1 {
			k++
		}
	}
	jm = 7 + jalDiv(k, 30)
	jd = jalMod(k, 30) + 1
	return jy, jm, jd
}

// Jalali converts a Gregorian time to the Jalali (Solar Hijri) calendar.
func Jalali(t time.Time) (jy int, jm int, jd int) {
	return d2j(g2d(t.Year(), int(t.Month()), t.Day()))
}

var jalaliMonths = [12]string{
	"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور",
	"مهر", "آبان", "آذر", "دی", "بهمن", "اسفند",
}

var persianWeekdays = map[time.Weekday]string{
	time.Saturday:  "شنبه",
	time.Sunday:    "یک‌شنبه",
	time.Monday:    "دوشنبه",
	time.Tuesday:   "سه‌شنبه",
	time.Wednesday: "چهارشنبه",
	time.Thursday:  "پنج‌شنبه",
	time.Friday:    "جمعه",
}

// PersianDigits renders Latin digits as Persian digits.
func PersianDigits(s string) string {
	const persian = "۰۱۲۳۴۵۶۷۸۹"
	digits := []rune(persian)
	out := []rune(s)
	for i, r := range out {
		if r >= '0' && r <= '9' {
			out[i] = digits[r-'0']
		}
	}
	return string(out)
}

// JalaliDateLong renders e.g. «شنبه ۲۴ شهریور ۱۴۰۵».
func JalaliDateLong(t time.Time) string {
	jy, jm, jd := Jalali(t)
	weekday := persianWeekdays[t.Weekday()]
	return fmt.Sprintf("%s %s %s %s", weekday, PersianDigits(fmt.Sprintf("%d", jd)), jalaliMonths[jm-1], PersianDigits(fmt.Sprintf("%d", jy)))
}

// JalaliDateShort renders e.g. «۱۴۰۵/۰۶/۲۴».
func JalaliDateShort(t time.Time) string {
	jy, jm, jd := Jalali(t)
	return PersianDigits(fmt.Sprintf("%04d/%02d/%02d", jy, jm, jd))
}
