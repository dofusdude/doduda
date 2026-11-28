package main

import (
	"encoding/json"
	"testing"

	mapping "github.com/dofusdude/dodumap"
)

func TestCurrentSpellEffectFieldsUnmarshal(t *testing.T) {
	var effect rawSpellEffect
	if err := json.Unmarshal([]byte(`{"actionId":195,"effectUid":42,"targetId":7,"displayZero":1,"modificator":3,"zoneDescr":{"shape":80}}`), &effect); err != nil {
		t.Fatal(err)
	}
	if effect.EffectID != 195 || effect.EffectUID != 42 || effect.TargetID != 7 || effect.DisplayZero != 1 || effect.Modificator != 3 {
		t.Fatalf("current spell effect fields were not mapped: %#v", effect)
	}
}

func TestCurrentMonsterGradeFieldsUnmarshal(t *testing.T) {
	var grade rawMonsterGrade
	if err := json.Unmarshal([]byte(`{"xp":2112,"reductionEarth":6,"reductionWater":-9,"paLostDodge":-1,"rangeBonus":2}`), &grade); err != nil {
		t.Fatal(err)
	}
	if grade.GradeXP != 2112 || grade.EarthResistance != 6 || grade.WaterResistance != -9 || grade.PADodge != -1 || grade.BonusRange != 2 {
		t.Fatalf("current monster grade fields were not mapped: %#v", grade)
	}
}

func TestSpellLevelUsesSharedConditionParser(t *testing.T) {
	languages := map[string]mapping.LangDictUnity{"en": {Texts: map[int]string{}}}
	data := &mapping.JSONGameDataUnity{ConditionReferences: map[string]map[int]mapping.ConditionReference{
		"HS": {
			7: {AnkamaID: 7, Name: map[string]string{"en": "Rooted"}},
			8: {AnkamaID: 8, Name: map[string]string{"en": "Invulnerable"}},
		},
	}}
	for _, raw := range []string{"HS=7&HS!8", ""} {
		level := mapSpellLevel(rawSpellLevel{StatesCriterion: raw}, nil, nil, nil, nil, data, &languages)
		want := mapping.ParseConditionUnity(raw, &languages, data)
		gotJSON, err := json.Marshal(level.StatesCriterion)
		if err != nil {
			t.Fatal(err)
		}
		wantJSON, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		if string(gotJSON) != string(wantJSON) {
			t.Fatalf("spell criterion differs from shared parser: %s != %s", gotJSON, wantJSON)
		}
		if raw != "" && level.StatesCriterion == nil {
			t.Fatal("spell condition disappeared")
		}
	}
}

func TestParseMonsterSpellGrades(t *testing.T) {
	grades, err := parseMonsterSpellGrades("1,16;2,17")
	if err != nil {
		t.Fatal(err)
	}
	if len(grades) != 2 {
		t.Fatalf("got %d grade assignments, want 2", len(grades))
	}
	if grades[1].MonsterGrade != 2 || grades[1].SpellGrade != 2 || grades[1].MonsterLevel != 17 {
		t.Fatalf("unexpected second grade assignment: %#v", grades[1])
	}
}

func TestParseMonsterSpellGradesRejectsMalformedValues(t *testing.T) {
	if _, err := parseMonsterSpellGrades("1,16;broken"); err == nil {
		t.Fatal("expected malformed assignment to fail")
	}
}

func TestParseMonsterSpellGradesSkipsNullPaddedMonsterGrades(t *testing.T) {
	grades, err := parseMonsterSpellGrades("1,1;1,66;1,132;1,null;1,null")
	if err != nil {
		t.Fatal(err)
	}
	if len(grades) != 3 || grades[2].MonsterGrade != 3 || grades[2].MonsterLevel != 132 {
		t.Fatalf("unexpected assignments: %#v", grades)
	}
}

func TestSpellAtResolvedGradeUsesExactGradeAndClampsLegacyValues(t *testing.T) {
	spell := mappedSpell{AnkamaID: 861, Levels: []mappedSpellLevel{{Grade: 1}, {Grade: 2}, {Grade: 3}}}
	exact, ok := spellAtResolvedGrade(spell, 2)
	if !ok || len(exact.Levels) != 1 || exact.Levels[0].Grade != 2 {
		t.Fatalf("exact grade was not selected: %#v", exact.Levels)
	}
	clamped, ok := spellAtResolvedGrade(spell, 6)
	if !ok || len(clamped.Levels) != 1 || clamped.Levels[0].Grade != 3 {
		t.Fatalf("legacy grade was not clamped: %#v", clamped.Levels)
	}
	if _, ok := spellAtResolvedGrade(spell, 0); ok {
		t.Fatal("grade zero must mean unavailable")
	}
}

func TestParseRawZone(t *testing.T) {
	zone := parseRawZone("X1,0,10,1")
	if zone.Shape.Code != "X" || zone.Shape.Name != "cross" {
		t.Fatalf("unexpected parsed shape: %#v", zone.Shape)
	}
	if zone.Size == nil || *zone.Size != 1 || zone.MinSize == nil || *zone.MinSize != 0 {
		t.Fatalf("unexpected named geometry: size=%v min_size=%v", zone.Size, zone.MinSize)
	}
	if zone.DamageDecreaseStepPercent != 10 || zone.MaxDamageDecreaseApplyCount != 1 {
		t.Fatalf("unexpected degression: %#v", zone)
	}
}

func TestMapClassAppearanceDecodesEntityLook(t *testing.T) {
	appearance, err := mapClassAppearance("{1|1438||52}", []int{0x112233, 0x445566})
	if err != nil {
		t.Fatal(err)
	}
	if appearance.BonesID != 1 || len(appearance.SkinIDs) != 1 || appearance.SkinIDs[0] != 1438 {
		t.Fatalf("unexpected decoded appearance: %#v", appearance)
	}
	if appearance.Scale.X != 52 || appearance.Scale.Y != 52 {
		t.Fatalf("unexpected uniform scale: %#v", appearance.Scale)
	}
	if len(appearance.DefaultColors) != 2 {
		t.Fatalf("default colors were not retained: %#v", appearance.DefaultColors)
	}
}

func TestPointZoneNormalizesEncodedDefaults(t *testing.T) {
	zone := parseRawZone("P")
	if zone.Size == nil || *zone.Size != 0 || zone.MinSize != nil {
		t.Fatalf("point geometry should be size 0 with no minimum: %#v", zone)
	}
}

func TestWholeMapZoneHasNoArtificialSize(t *testing.T) {
	zone := parseRawZone("A")
	if zone.Size != nil || zone.MinSize != nil {
		t.Fatalf("whole-map geometry should not expose an artificial size: %#v", zone)
	}
}

func TestAllCurrentZoneShapesAreNamed(t *testing.T) {
	for _, code := range []string{"", "#", "*", "+", "-", "/", ";", "A", "B", "C", "D", "F", "G", "I", "L", "O", "P", "Q", "R", "T", "U", "V", "W", "X", "Z", "a", "l"} {
		shape := mapZoneShape(code)
		if shape.Name == "" {
			t.Fatalf("shape %q has no semantic name", code)
		}
	}
}

func TestMapTargetMaskResolvesConditions(t *testing.T) {
	languages := map[string]mapping.LangDictUnity{"en": {Texts: map[int]string{100: "Rooted", 200: "Gobball"}}}
	states := map[int]rawSpellState{7: {ID: 7, NameID: 100}}
	monsters := map[int]rawMonster{42: {ID: 42, NameID: 200}}

	mask := mapTargetMask("a,*E7,F42,V50", states, monsters, nil, languages)
	if len(mask.Selectors) != 1 || mask.Selectors[0].Name != "allies" {
		t.Fatalf("unexpected selectors: %#v", mask.Selectors)
	}
	if len(mask.Conditions) != 3 || mask.Conditions[0].Subject != "caster" || mask.Conditions[0].Reference.Name["en"] != "Rooted" {
		t.Fatalf("state condition was not resolved: %#v", mask.Conditions)
	}
	if mask.Conditions[1].Reference.Name["en"] != "Gobball" || mask.Conditions[2].Name != "life_percent_at_most" {
		t.Fatalf("unexpected mapped conditions: %#v", mask.Conditions)
	}
}

func TestMapTriggersResolvesParametersAndOtherFighterScope(t *testing.T) {
	languages := map[string]mapping.LangDictUnity{"en": {Texts: map[int]string{100: "Rooted", 300: "Punishment"}}}
	states := map[int]rawSpellState{7: {ID: 7, NameID: 100}}
	spells := map[int]rawSpell{99: {ID: 99, NameID: 300}}

	triggers := mapTriggers("I|XD|EON7|TR99|EK:m", states, nil, nil, spells, languages)
	if len(triggers.Events) != 5 || triggers.Events[1].Name != "other_fighter_damage_received" || triggers.Events[1].Subject != "other_fighter" {
		t.Fatalf("other-fighter trigger was not mapped: %#v", triggers.Events)
	}
	if triggers.Events[2].Reference.Name["en"] != "Rooted" || triggers.Events[3].Reference.Name["en"] != "Punishment" {
		t.Fatalf("parameter references were not resolved: %#v", triggers.Events)
	}
	if triggers.Events[4].TargetMask == nil || triggers.Events[4].TargetMask.Selectors[0].Name != "allied_monsters" {
		t.Fatalf("embedded target mask was not mapped: %#v", triggers.Events[4])
	}
}

func TestMapTriggersResolvesMatchingFighterCount(t *testing.T) {
	triggers := mapTriggers("EC:>0:m,h", nil, nil, nil, nil, nil)
	if len(triggers.Events) != 1 {
		t.Fatalf("unexpected events: %#v", triggers.Events)
	}
	event := triggers.Events[0]
	if event.Name != "matching_fighter_count" || event.Operator != "greater_than" || event.Parameter == nil || *event.Parameter != 0 {
		t.Fatalf("count trigger was not decoded: %#v", event)
	}
	if event.TargetMask == nil || len(event.TargetMask.Selectors) != 2 {
		t.Fatalf("embedded target mask was not decoded: %#v", event.TargetMask)
	}
}

func TestMapTargetMaskParsesSetsAndComparisons(t *testing.T) {
	mask := mapTargetMask("a,A,HCH{11,12,13},*Ct33276<3,ich11,F50000", nil, nil, nil, map[string]mapping.LangDictUnity{})
	if len(mask.Selectors) != 2 || len(mask.Conditions) != 4 {
		t.Fatalf("unexpected split: %#v", mask)
	}
	set := mask.Conditions[0]
	if set.Code != "HCH{11,12,13}" || set.Name != "" || set.Operator != "" || len(set.Values) != 3 || set.Values[2] != 13 {
		t.Fatalf("unconfirmed set condition must be code and values only: %#v", set)
	}
	counter := mask.Conditions[1]
	if counter.Subject != "caster" || counter.Name != "" || counter.Reference != nil || counter.Value == nil || *counter.Value != 33276 ||
		counter.Comparison == nil || counter.Comparison.Operator != "less_than" || counter.Comparison.Value != 3 {
		t.Fatalf("unconfirmed counter condition must be code, value and comparison only: %#v", counter)
	}
	if mask.Conditions[2].Name != "" || mask.Conditions[2].Operator != "" || *mask.Conditions[2].Value != 11 {
		t.Fatalf("unexpected unconfirmed condition: %#v", mask.Conditions[2])
	}
	if mask.Conditions[3].Name != "monster_is" {
		t.Fatalf("confirmed condition lost its name: %#v", mask.Conditions[3])
	}
}

func TestMapTriggersParsesSpellCounters(t *testing.T) {
	triggers := mapTriggers("I|CTI33276|CTT33276>0|NWS", nil, nil, nil, nil, map[string]mapping.LangDictUnity{})
	if len(triggers.Events) != 4 || triggers.Events[0].Name != "immediate" {
		t.Fatalf("unexpected events: %#v", triggers.Events)
	}
	if e := triggers.Events[1]; e.Code != "CTI33276" || e.Name != "" || e.Subject != "" || e.Reference != nil || *e.Parameter != 33276 {
		t.Fatalf("unconfirmed trigger must be code and parameter only: %#v", e)
	}
	if e := triggers.Events[2]; e.Name != "" || *e.Parameter != 33276 || e.Comparison == nil || e.Comparison.Operator != "greater_than" || e.Comparison.Value != 0 {
		t.Fatalf("unexpected threshold trigger: %#v", e)
	}
	if e := triggers.Events[3]; e.Code != "NWS" || e.Name != "" || e.Parameter != nil {
		t.Fatalf("unexpected NWS trigger: %#v", e)
	}
}

func TestMapTargetMaskParsesAtFlag(t *testing.T) {
	mask := mapTargetMask("C,at", nil, nil, nil, map[string]mapping.LangDictUnity{})
	if len(mask.Selectors) != 1 || len(mask.Conditions) != 1 || mask.Conditions[0].Code != "at" || mask.Conditions[0].Name != "" || mask.Conditions[0].Value != nil {
		t.Fatalf("unexpected at mask: %#v", mask)
	}
}

func TestMonsterDropRatesSupportsBothFormats(t *testing.T) {
	legacy := monsterDropRates(rawMonsterDrop{PercentDropForGrade1: 10, PercentDropForGrade5: 50})
	if len(legacy) != 5 || legacy[0].Percent != 10 || legacy[4].Grade != 5 || legacy[4].Percent != 50 {
		t.Fatalf("unexpected legacy rates: %#v", legacy)
	}
	current := monsterDropRates(rawMonsterDrop{PercentDropsForGrades: unityArray[float64]{Array: []float64{1, 2, 3, 4, 5}}})
	if len(current) != 5 || current[0].Percent != 1 || current[4].Grade != 5 || current[4].Percent != 5 {
		t.Fatalf("unexpected array rates: %#v", current)
	}
}
