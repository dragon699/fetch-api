package utils

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ToInt(val string) (int64, error) {
	i, err := strconv.Atoi(val)

	if err != nil {
		return 0, fmt.Errorf("expected int-like value, got %s", val)
	}
	return int64(i), nil
}

func ToBool(val string) (bool, error) {
	b, err := strconv.ParseBool(val)

	if err != nil {
		return false, fmt.Errorf("expected bool-like value, got %s", val)
	}
	return b, nil
}

func HasItemWithPrefix(items []string, prefix string) bool {
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			return true
		}
	}

	return false
}

func ReadFile(path string) (string, error) {
	contents, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func TrimKubeTime(time metav1.Time) string {
	return time.Time.Format("2006-01-02T15:04:05")
}

func TimeGetOrdinalDay(day int) string {
	if day >= 11 && day <= 13 {
		return "th"
	}

	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func TimeFromUnix(unixTime int64) string {
	timeObj := time.Unix(unixTime, 0)
	day := timeObj.Day()

	return fmt.Sprintf(
		"%d%s of %s %d at %s",
		day,
		TimeGetOrdinalDay(day),
		timeObj.Format("Jan"),
		timeObj.Year(),
		timeObj.Format("15:04"),
	)
}

func TimeFromGrafana(stringTime string) string {
	t, err := time.Parse(time.RFC3339Nano, stringTime)

	if err != nil {
		return stringTime
	}

	loc, err := time.LoadLocation("Europe/Sofia")

	if err != nil {
		loc = time.UTC
	}

	t.In(loc)
	return fmt.Sprintf("on *%02d/%02d* at *%02d:%02d*", t.Day(), t.Month(), t.Hour(), t.Minute())
}

func BytesToMegabytes(bytes int64) int64 {
	return bytes / (1024 * 1024)
}

func BytesToGigabytes(bytes int64) float64 {
	return RoundToTwoDecimals(float64(bytes) / (1024 * 1024 * 1024))
}

func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

func ProgressToPercentage(progress float64) float64 {
	if progress >= 1 {
		return 100
	}
	if progress <= 0 {
		return 0
	}

	percentage := progress * 100

	if progress >= 0.97 {
		return math.Floor(percentage*100) / 100
	}

	return RoundToTwoDecimals(percentage)
}

func BytesPerSecondToMBps(bytesPerSec int64) float64 {
	mbps := float64(bytesPerSec) / (1024 * 1024)
	return RoundToTwoDecimals(mbps)
}

func SecondsToMinutes(seconds int64) int64 {
	if seconds < 60 {
		return 1
	}
	return seconds / 60
}

func CapitalizeString(text string) string {
	if text == "" {
		return text
	}

	return strings.ToUpper(text[:1]) + strings.ToLower(text[1:])
}

var seasonEpisodeRe = regexp.MustCompile(`(?i)\b(?:S(?:eason)?\s*(\d{1,2})\s*(?:E(?:pisode)?\s*(\d{1,2}))|E(?:pisode)?\s*(\d{1,2})\s*S(?:eason)?\s*(\d{1,2}))\b`)
var seasonOnlyRe = regexp.MustCompile(`(?i)\bS(?:eason)?\s*(\d{1,2})\b`)
var episodeOnlyRe = regexp.MustCompile(`(?i)\bE(?:pisode)?\s*(\d{1,2})\b`)
var yearRe = regexp.MustCompile(`\b(19|20)\d{2}\b`)
var collectionRangeRe = regexp.MustCompile(`\b(\d{1,2})\s*-\s*(\d{1,2})\b`)
var collectionWordRe = regexp.MustCompile(`(?i)\bcollection\b`)

var junkWords = map[string]struct{}{
	"1080p": {}, "720p": {}, "480p": {},
	"2160p": {}, "4k": {},
	"bluray": {}, "brrip": {}, "webrip": {}, "webdl": {}, "web": {},
	"hdrip": {}, "dvdrip": {}, "remux": {}, "proper": {}, "repack": {},
	"x264": {}, "x265": {}, "h264": {}, "h265": {}, "hevc": {},
	"aac": {}, "dts": {}, "ac3": {}, "dd5": {}, "ddp5": {},
	"cam": {}, "ts": {}, "hc": {},
}

func normalize(name string) string {
	replacer := strings.NewReplacer(
		".", " ",
		"-", " ",
		"_", " ",
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
	)
	name = replacer.Replace(name)
	name = strings.Join(strings.Fields(name), " ")
	return name
}

func removeJunk(title string) string {
	words := strings.Fields(title)
	var clean []string

	for _, w := range words {
		l := strings.ToLower(w)

		if len(l) > 4 && isNumeric(l) {
			continue
		}

		if _, exists := junkWords[l]; exists {
			continue
		}

		clean = append(clean, w)
	}

	return strings.Join(clean, " ")
}

func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func titleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		words[i] = strings.ToUpper(string(w[0])) + w[1:]
	}
	return strings.Join(words, " ")
}

func BeautifyMovieName(name string) string {
	rawName := name
	name = normalize(name)

	if m := seasonEpisodeRe.FindStringSubmatchIndex(name); m != nil {
		season := 0
		episode := 0

		if m[2] != -1 && m[3] != -1 && m[4] != -1 && m[5] != -1 {
			season, _ = strconv.Atoi(name[m[2]:m[3]])
			episode, _ = strconv.Atoi(name[m[4]:m[5]])
		} else if m[6] != -1 && m[7] != -1 && m[8] != -1 && m[9] != -1 {
			episode, _ = strconv.Atoi(name[m[6]:m[7]])
			season, _ = strconv.Atoi(name[m[8]:m[9]])
		}

		prefix := strings.TrimSpace(name[:m[0]])
		suffix := strings.TrimSpace(name[m[1]:])
		prefix = titleCase(removeJunk(prefix))
		suffix = titleCase(removeJunk(suffix))

		label := fmt.Sprintf("S%02dE%02d", season, episode)
		if prefix == "" && suffix == "" {
			return label
		}

		if prefix == "" {
			return fmt.Sprintf("%s %s", label, suffix)
		}
		if suffix == "" {
			return fmt.Sprintf("%s %s", prefix, label)
		}
		return fmt.Sprintf("%s %s %s", prefix, label, suffix)
	}

	if m := seasonOnlyRe.FindStringSubmatchIndex(name); m != nil {
		season, _ := strconv.Atoi(name[m[2]:m[3]])
		prefix := strings.TrimSpace(name[:m[0]])
		suffix := strings.TrimSpace(name[m[1]:])
		prefix = titleCase(removeJunk(prefix))
		suffix = titleCase(removeJunk(suffix))

		label := fmt.Sprintf("S%02d", season)
		if prefix == "" && suffix == "" {
			return label
		}
		if prefix == "" {
			return fmt.Sprintf("%s %s", label, suffix)
		}
		if suffix == "" {
			return fmt.Sprintf("%s %s", prefix, label)
		}
		return fmt.Sprintf("%s %s %s", prefix, label, suffix)
	}

	if m := episodeOnlyRe.FindStringSubmatchIndex(name); m != nil {
		episode, _ := strconv.Atoi(name[m[2]:m[3]])
		prefix := strings.TrimSpace(name[:m[0]])
		suffix := strings.TrimSpace(name[m[1]:])
		prefix = titleCase(removeJunk(prefix))
		suffix = titleCase(removeJunk(suffix))

		label := fmt.Sprintf("E%02d", episode)
		if prefix == "" && suffix == "" {
			return label
		}
		if prefix == "" {
			return fmt.Sprintf("%s %s", label, suffix)
		}
		if suffix == "" {
			return fmt.Sprintf("%s %s", prefix, label)
		}
		return fmt.Sprintf("%s %s %s", prefix, label, suffix)
	}

	years := yearRe.FindAllString(name, -1)
	var year string
	if len(years) > 0 {
		year = years[len(years)-1]
	}

	if year != "" {
		loc := yearRe.FindStringIndex(name)
		title := strings.TrimSpace(name[:loc[0]])

		if cm := collectionRangeRe.FindStringSubmatch(rawName); cm != nil {
			start := cm[1]
			end := cm[2]

			rangeInTitleRe := regexp.MustCompile(`\b` + regexp.QuoteMeta(start) + `\s+` + regexp.QuoteMeta(end) + `\b`)
			title = rangeInTitleRe.ReplaceAllString(title, " ")
			title = collectionWordRe.ReplaceAllString(title, " ")
			title = strings.Join(strings.Fields(title), " ")
			title = removeJunk(title)
			title = titleCase(title)

			return fmt.Sprintf("%s %s-%s (%s)", title, start, end, year)
		}

		title = removeJunk(title)
		title = titleCase(title)

		return fmt.Sprintf("%s (%s)", title, year)
	}

	name = removeJunk(name)
	return titleCase(name)
}

func templateJSON(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return `""`
	}

	return string(body)
}

func RenderTemplate(templatePath string, templateVars any, embeddedFiles embed.FS) (string, error) {
	templateBytes, err := embeddedFiles.ReadFile(templatePath)

	if err != nil {
		return "", fmt.Errorf("[%s] Invalid or missing template file: %s", err, templatePath)
	}

	templateContent, err := renderTemplateContent(templatePath, string(templateBytes), templateVars)

	if err != nil {
		return "", fmt.Errorf("[%s] Failed to render embedded template file: %w", templatePath, err)
	}

	return templateContent, nil
}

func renderTemplateContent(templateName string, templateContent string, vars any) (string, error) {
	tpl, err := template.New(templateName).Funcs(template.FuncMap{
		"json":         templateJSON,
		"beautifyTime": TimeFromGrafana,
		"hasPrefix":    strings.HasPrefix,
		"contains":     strings.Contains,
		"capitalize":   CapitalizeString,
		"toUpper":      strings.ToUpper,
	}).Parse(templateContent)

	if err != nil {
		return "", err
	}

	var out strings.Builder

	if err := tpl.Execute(&out, vars); err != nil {
		return "", err
	}

	return out.String(), nil
}
