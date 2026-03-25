// Copyright 2025-2026 Hancom Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0

package utils

import (
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
	leftRatio := line.BBox.X / pageWidth
	switch {
	case leftRatio < 0.08:
		return 0
	case leftRatio < 0.16:
		return 1
	case leftRatio < 0.24:
		return 2
	case leftRatio < 0.32:
		return 3
	case leftRatio < 0.40:
		return 4
	default:
		return 5
	}
}
