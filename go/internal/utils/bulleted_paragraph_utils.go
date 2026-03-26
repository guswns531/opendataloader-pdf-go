// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package utils

import (
	"math"
	"regexp"
	"strings"

	"github.com/opendataloader-project/opendataloader-pdf-go/internal/entities"
)

const possibleLabels = "∘*+-.=‐‑‒–—―•‣․‧※⁃⁎→↳⇒⇨⇾∙■□▢▣▤▥▦▧▨▩▪▬▭▮▯▰▱▲△▴▵▶▷▸▹►▻▼▽▾▿◀◁◂◃◄◅◆◇◈◉◊○◌◍" +
	"◎●◐◑◒◓◔◕◖◗◘◙◢◣◤◥◦◧◨◩◪◫◬◭◮◯◰◱◲◳◴◵◶◷◸◹◺◻◼◽◾◿★☆☐☑☒☓☛☞♠♡♢♣♤♥♦♧⚪⚫⚬✓✔✕✖✗✘✙✚✛✜✝✞✟✦✧✨❍❏❐❑" +
	"❒❖➔➙➛➜➝➞➟➠➡➢➣➤➥➦➧➨➩➪➭➮➯➱⬛⬜⬝⬞⬟⬠⬡⬢⬣⬤⬥⬦⬧⬨⬩⬪⬫⬬⬭⬮⬯⭐⭑⭒⭓⭔⭕⭖⭗⭘⭙⯀⯁⯂⯃⯄⯅⯆⯇⯈⯌⯍⯎⯏⯐〇" +
	"󰁾󰋪󰋫󰋬󰋭󰋮󰋯󰋰󰋱󰋲󰋳󰋴󰋵󰋶󰋷󰋸󰋹󰋺󰋻󰋼"

const KoreanChapterRegex = `^(제\d+[장조절]).*`

var (
	koreanNumbersRegex = `[가나다라마바사아자차카타파하거너더러머버서어저처커터퍼허고노도로모보소오조초코토포호구누두루무부수우주추쿠투푸후그느드르므브스으즈츠크트프흐기니디리미비시이지치키티피히]`

	orderedBulletRegexes = []*regexp.Regexp{
		regexp.MustCompile(`^(\d+\.|[a-zA-Z]\.|[①-⑳]|[ⓐ-ⓩ]|\([0-9]+\)).*`),
		regexp.MustCompile(`^\(\d+\).*`),
		regexp.MustCompile(`^<\d+>.*`),
		regexp.MustCompile(`^\[\d+\].*`),
		regexp.MustCompile(`^\{\d+\}.*`),
		regexp.MustCompile(`^【\d+】.*`),
		regexp.MustCompile(`^\d+[\.\)]\s+.*`),
		regexp.MustCompile(`^\d+[ \.\]\)>].*`),
		regexp.MustCompile(`^[ㄱㄴㄷㄹㅁㅂㅅㅇㅈㅊㅋㅌㅍㅎ][\.\)\]>].*`),
		regexp.MustCompile(`^` + koreanNumbersRegex + `\..+`),
		regexp.MustCompile(`^` + koreanNumbersRegex + `[\)\]>].*`),
		regexp.MustCompile(`^` + koreanNumbersRegex + `(-\d+).*`),
		regexp.MustCompile(`^\(` + koreanNumbersRegex + `\).*`),
		regexp.MustCompile(`^<` + koreanNumbersRegex + `>.*`),
		regexp.MustCompile(`^\[` + koreanNumbersRegex + `\].*`),
		regexp.MustCompile(`^\{` + koreanNumbersRegex + `\}.*`),
		regexp.MustCompile(KoreanChapterRegex),
		regexp.MustCompile(`^법\.(제\d+조).*`),
		regexp.MustCompile(`^[I]\..*`),
		regexp.MustCompile(`^[Ⅰ-Ⅻ].*`),
		regexp.MustCompile(`^[ⅰ-ⅻ].*`),
		regexp.MustCompile(`^[①-⑳].*`),
		regexp.MustCompile(`^[⑴-⒇].*`),
		regexp.MustCompile(`^[⒈-⒛].*`),
		regexp.MustCompile(`^[⒜-⒵].*`),
		regexp.MustCompile(`^[Ⓐ-Ⓩ].*`),
		regexp.MustCompile(`^[ⓐ-ⓩ].*`),
		regexp.MustCompile(`^[⓵-⓾].*`),
		regexp.MustCompile(`^[❶-❿].*`),
		regexp.MustCompile(`^[➀-➉].*`),
		regexp.MustCompile(`^[➊-➓].*`),
		regexp.MustCompile(`^[㉮-㉻].*`),
		regexp.MustCompile(`^[-].*`),
		regexp.MustCompile(`^[-].*`),
	}
)

func IsBulletedParagraph(line *entities.TextLine) bool {
	return IsBulletedLine(line)
}

func IsBulletedLine(line *entities.TextLine) bool {
	return IsLabeledLine(line)
}

func IsLabeledLine(line *entities.TextLine) bool {
	if line == nil {
		return false
	}
	value := strings.TrimSpace(line.GetText())
	if value == "" {
		return false
	}
	first := []rune(value)[0]
	if strings.ContainsRune(possibleLabels, first) || line.LineArtBullet != nil {
		return true
	}
	for _, re := range orderedBulletRegexes {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func IsBulletedLineArtParagraph(line *entities.TextLine) bool {
	return line != nil && line.LineArtBullet != nil
}

func GetLabel(line *entities.TextLine) string {
	if line == nil {
		return ""
	}
	value := strings.TrimSpace(line.GetText())
	if value == "" {
		return ""
	}
	return string([]rune(value)[0])
}

func GetLabelRegex(line *entities.TextLine) string {
	if line == nil {
		return ""
	}
	value := strings.TrimSpace(line.GetText())
	for _, re := range orderedBulletRegexes {
		if re.MatchString(value) {
			return re.String()
		}
	}
	return ""
}

func IsOrderedBullet(text string) bool {
	value := strings.TrimSpace(text)
	for _, re := range orderedBulletRegexes {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func IsUnorderedBullet(text string) bool {
	value := strings.TrimSpace(text)
	if value == "" || IsOrderedBullet(value) {
		return false
	}
	r := []rune(value)
	if len(r) > 0 && strings.ContainsRune(possibleLabels, r[0]) {
		return true
	}
	return false
}

func GetIndentLevel(line *entities.TextLine, pageWidth float64) int {
	if line == nil || pageWidth <= 0 {
		return 0
	}
	fontSize := 12.0
	for _, chunk := range line.Chunks {
		if chunk != nil && chunk.FontStyle.FontSize > 0 {
			fontSize = chunk.FontStyle.FontSize
			break
		}
	}
	step := math.Max(fontSize*1.5, pageWidth*0.08)
	return int(math.Max(0, math.Floor(line.BBox.X/step)))
}
