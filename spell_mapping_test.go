package main

import (
	"testing"

	mapping "github.com/dofusdude/dodumap"
)

func TestMapStateCriterion(t *testing.T) {
	languages := map[string]mapping.LangDictUnity{
		"en": {Texts: map[int]string{100: "Rooted", 101: "Invulnerable"}},
		"fr": {Texts: map[int]string{100: "Enraciné", 101: "Invulnérable"}},
	}
	states := map[int]rawSpellState{
		7: {ID: 7, NameID: 100},
		8: {ID: 8, NameID: 101},
	}

	criterion := mapStateCriterion("HS=7&HS!8|HS=999", states, languages)
	if criterion == nil {
		t.Fatal("expected a mapped criterion")
	}
	if got, want := criterion.Templated["en"], "has state Rooted and does not have state Invulnerable or has state 999"; got != want {
		t.Fatalf("English criterion = %q, want %q", got, want)
	}
	if len(criterion.StateIDs) != 3 || criterion.StateIDs[0] != 7 || criterion.StateIDs[1] != 8 || criterion.StateIDs[2] != 999 {
		t.Fatalf("state IDs were not retained in encounter order: %v", criterion.StateIDs)
	}
	if len(criterion.UnresolvedStateIDs) != 1 || criterion.UnresolvedStateIDs[0] != 999 {
		t.Fatalf("unexpected unresolved states: %v", criterion.UnresolvedStateIDs)
	}
}

func TestMapStateCriterionEmpty(t *testing.T) {
	if criterion := mapStateCriterion("", nil, nil); criterion != nil {
		t.Fatalf("empty criterion should map to null, got %#v", criterion)
	}
}

func TestParseMonsterSpellGrades(t *testing.T) {
	grades := parseMonsterSpellGrades("1,16;2,17;broken")
	if len(grades) != 2 {
		t.Fatalf("got %d grade ranges, want 2", len(grades))
	}
	if grades[1].MonsterGrade != 2 || grades[1].MinSpellGrade != 2 || grades[1].MaxSpellGrade != 17 {
		t.Fatalf("unexpected second grade range: %#v", grades[1])
	}
}

func TestParseRawZone(t *testing.T) {
	zone := parseRawZone("X1,0,10,1")
	if zone.Shape.Code != "X" || zone.Shape.Name != "diagonal_cross" {
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
