package main

import (
	"log"
	"os"
	"testing"

	"github.com/dofusdude/doduda/mapping"
)

var TestingLangs map[string]mapping.LangDict
var TestingData *mapping.JSONGameData

func TestMain(m *testing.M) {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	TestingLangs = mapping.ParseRawLanguages(path)
	TestingData = mapping.ParseRawData(path)
	mapping.LoadPersistedElements(path)

	if TestingLangs == nil {
		log.Fatal("testingLangs is nil")
	}

	if TestingData == nil {
		log.Fatal("testingData is nil")
	}

	m.Run()
}

func TestParseSigness1(t *testing.T) {
	num, side := mapping.ParseSigness("-#1{~1~2 und -}#2")
	if !num {
		t.Error("num is false")
	}
	if !side {
		t.Error("side is false")
	}
}

func TestParseSigness2(t *testing.T) {
	num, side := mapping.ParseSigness("#1{~1~2 und -}#2")
	if num {
		t.Error("num is true")
	}
	if !side {
		t.Error("side is false")
	}
}

func TestParseSigness3(t *testing.T) {
	num, side := mapping.ParseSigness("#1{~1~2 und }#2")
	if side {
		t.Error("side is true")
	}
	if num {
		t.Error("num is true")
	}
}

func TestParseSigness4(t *testing.T) {
	num, side := mapping.ParseSigness("-#1{~1~2 und }-#2")
	if !side {
		t.Error("side is false")
	}
	if !num {
		t.Error("num is false")
	}
}

func TestParseConditionSimple(t *testing.T) {
	condition := mapping.ParseCondition("cs<25", &TestingLangs, TestingData)

	if len(condition) != 1 {
		t.Errorf("condition length is not 1: %d", len(condition))
	}
	if condition[0].Operator != "<" {
		t.Errorf("operator is not <: %s", condition[0].Operator)
	}
	if condition[0].Value != 25 {
		t.Errorf("value is not 25: %d", condition[0].Value)
	}
	if condition[0].Templated["de"] != "Stärke" {
		t.Errorf("templated is not Stärke: %s", condition[0].Templated["de"])
	}
}

func TestParseConditionMulti(t *testing.T) {
	condition := mapping.ParseCondition("CS>80&CV>40&CA>40", &TestingLangs, TestingData)

	if len(condition) != 3 {
		t.Errorf("condition length is not 3: %d", len(condition))
	}

	if condition[0].Operator != ">" {
		t.Errorf("operator is not >: %s", condition[0].Operator)
	}
	if condition[0].Value != 80 {
		t.Errorf("value is not 80: %d", condition[0].Value)
	}
	if condition[0].Templated["de"] != "Stärke" {
		t.Errorf("templated is not Stärke: %s", condition[0].Templated["de"])
	}

	if condition[1].Operator != ">" {
		t.Errorf("operator is not >: %s", condition[1].Operator)
	}
	if condition[1].Value != 40 {
		t.Errorf("value is not 40: %d", condition[1].Value)
	}
	if condition[1].Templated["de"] != "Vitalität" {
		t.Errorf("templated is not Vitalität: %s", condition[1].Templated["de"])
	}

	if condition[2].Operator != ">" {
		t.Errorf("operator is not >: %s", condition[2].Operator)
	}
	if condition[2].Value != 40 {
		t.Errorf("value is not 40: %d", condition[2].Value)
	}
	if condition[2].Templated["de"] != "Flinkheit" {
		t.Errorf("templated is not Flinkheit: %s", condition[2].Templated["de"])
	}
}

func TestDeleteNumHash(t *testing.T) {
	effect_name := mapping.DeleteDamageFormatter("Austauschbar ab: #1")
	if effect_name != "Austauschbar ab:" {
		t.Errorf("output is not as expected: %s", effect_name)
	}
}

func TestParseConditionEmpty(t *testing.T) {
	condition := mapping.ParseCondition("null", &TestingLangs, TestingData)
	if len(condition) > 0 {
		t.Errorf("condition should be empty")
	}
}

func TestParseSingularPluralFormatterNormal(t *testing.T) {
	formatted := mapping.SingularPluralFormatter("Filzpunkte", 1, "de")
	if formatted != "Filzpunkte" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestParseSingularPluralFormatterPlural(t *testing.T) {
	formatted := mapping.SingularPluralFormatter("Kommt in %1 Subgebiet{~pen} vor", 2, "es")
	if formatted != "Kommt in %1 Subgebieten vor" {
		t.Errorf("output is not as expected: %s", formatted)
	}

	formatted = mapping.SingularPluralFormatter("Punkt{~pe} erforderlich", 2, "es")
	if formatted != "Punkte erforderlich" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestParseSingularPluralFormatterPluralMulti(t *testing.T) {
	formatted := mapping.SingularPluralFormatter("Kommt in %1 Subgebiet{~pen} mit Punkt{~pe} vor", 2, "es")
	if formatted != "Kommt in %1 Subgebieten mit Punkte vor" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestParseSingularPluralFormatterSingularMulti(t *testing.T) {
	formatted := mapping.SingularPluralFormatter("Kommt in %1 Subgebiet{~sen} mit Punkt{~se} vor", 1, "es")
	if formatted != "Kommt in %1 Subgebieten mit Punkte vor" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestParseSingularPluralFormatterPluralComplexUnicode(t *testing.T) {
	formatted := mapping.SingularPluralFormatter("invocaç{~pões}", 2, "pt")
	if formatted != "invocações" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestParseSingularPluralFormatterPluralDeleteIfSingular(t *testing.T) {
	formatted := mapping.SingularPluralFormatter("invocaç{~pões}", 1, "pt")
	if formatted != "invocaç" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestDeleteDamageTemplate(t *testing.T) {
	formatted := mapping.DeleteDamageFormatter("#1{~1~2 bis }#2 (Erdschaden)")
	if formatted != "(Erdschaden)" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestDeleteDamageTemplateLevelEnBug(t *testing.T) {
	formatted := mapping.DeleteDamageFormatter("+#1{~1~2 to}level #2")
	if formatted != "level" {
		t.Errorf("output is not as expected: %s", formatted)
	}
}

func TestParseNumSpellNameFormatterItSpecial(t *testing.T) {
	input := "Ottieni: #1{~1~2 - }#2 kama"
	diceNum := 100
	diceSide := 233
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "it", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Ottieni: 100 - 233 kama" {
		t.Errorf("output is not as expected: %s", output)
	}

	if diceNum != 100 {
		t.Errorf("diceNum is not as expected: %d", diceNum)
	}

	if diceSide != 233 {
		t.Errorf("diceSide is not as expected: %d", diceSide)
	}
}

func TestParseNumSpellNameFormatterItSpecialSwitch(t *testing.T) {
	input := "#2: +#1 EP"
	diceNum := 100
	diceSide := 36
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "it", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "36: +100 EP" {
		t.Errorf("output is not as expected: %s", output)
	}

	if diceNum != 100 {
		t.Errorf("diceNum is not as expected: %d", diceNum)
	}

	if diceSide != 36 {
		t.Errorf("diceSide is not as expected: %d", diceSide)
	}
}

func TestParseNumSpellNameFormatterLearnSpellLevel(t *testing.T) {
	input := "Stufe #3 des Zauberspruchs erlernen"
	diceNum := 0
	diceSide := 0
	value := 1746
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Stufe 1746 des Zauberspruchs erlernen" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterLearnSpellLevel1(t *testing.T) {
	input := "Stufe #3 des Zauberspruchs erlernen"
	diceNum := 0
	diceSide := 1
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Stufe 1 des Zauberspruchs erlernen" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterDeNormal(t *testing.T) {
	input := "#1{~1~2 bis }#2 Kamagewinn"
	diceNum := 100
	diceSide := 233
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "100 bis 233 Kamagewinn" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterMultiValues(t *testing.T) {
	input := "Erfolgschance zwischen #1{~1~2 und }#2%"
	diceNum := 1
	diceSide := 2
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Erfolgschance zwischen 1 und 2%" {
		t.Errorf("output is not as expected: %s", output)
	}

	input = "Erfolgschance zwischen -#1{~1~2 und -}#2%"
	diceNum = 1
	diceSide = 2
	value = 0
	output, _ = mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Erfolgschance zwischen -1 und -2%" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterVitaRange(t *testing.T) {
	input := "+#1{~1~2 bis }#2 Vitalität"
	diceNum := 0
	diceSide := 300
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, true)

	if output != "0 bis 300 Vitalität" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterSingle(t *testing.T) {
	input := "Austauschbar ab: #1"
	diceNum := 1
	diceSide := 0
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Austauschbar ab: 1" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterMinMax(t *testing.T) {
	input := "Verbleib. Anwendungen: #2 / #3" // delete the min max
	diceNum := 2
	diceSide := 5
	value := 6
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)
	if output != "Verbleib. Anwendungen: 5 / 6" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterSpellDiceNum(t *testing.T) {
	input := "Zauberwurf: #1"
	diceNum := 15960
	diceSide := 0
	value := 0
	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)

	if output != "Zauberwurf: Mauschelei" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterEffectsRange(t *testing.T) {
	input := "-#1{~1~2 bis -}#2 Luftschaden"
	diceNum := 25
	diceSide := 50
	value := 0

	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)
	if output != "-25 bis -50 Luftschaden" {
		t.Errorf("output is not as expected: %s", output)
	}
}

func TestParseNumSpellNameFormatterMissingWhite(t *testing.T) {
	input := "+#1{~1~2 to}level #2"
	diceNum := 1
	diceSide := 0
	value := 0

	output, _ := mapping.NumSpellFormatter(input, "de", TestingData, &TestingLangs, &diceNum, &diceSide, &value, 0, false, false)
	if output != "1 level" {
		t.Errorf("output is not as expected: %s", output)
	}
}
