package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	mapping "github.com/dofusdude/dodumap"
)

type mappedCombatCondition struct {
	Code          string                `json:"code"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	Subject       string                `json:"subject,omitempty"`
	Operator      string                `json:"operator,omitempty"`
	ReferenceType string                `json:"reference_type,omitempty"`
	Value         *int                  `json:"value,omitempty"`
	Reference     *mappedNamedReference `json:"reference,omitempty"`
}

type mappedTargetMask struct {
	Raw        string                  `json:"raw"`
	Logic      string                  `json:"logic"`
	Selectors  []mappedCombatCondition `json:"selectors"`
	Conditions []mappedCombatCondition `json:"conditions"`
}

var targetSelectorNames = map[string][2]string{
	"A": {"enemies", "enemy fighters"},
	"a": {"allies", "allied fighters, including the caster"},
	"C": {"caster", "the spell caster"},
	"c": {"caster", "the spell caster"},
	"g": {"allies_except_caster", "allied fighters other than the caster"},
	"H": {"enemy_players", "enemy player characters"},
	"h": {"allied_players", "allied player characters"},
	"L": {"enemy_players_or_companions", "enemy players or companions"},
	"l": {"allied_players_or_companions", "allied players or companions"},
	"M": {"enemy_monsters", "enemy monsters"},
	"m": {"allied_monsters", "allied monsters"},
	"D": {"enemy_companions", "enemy companions"},
	"d": {"allied_companions", "allied companions"},
	"I": {"enemy_controllable_summons", "enemy controllable summons"},
	"i": {"allied_controllable_summons", "allied controllable summons"},
	"J": {"enemy_summons", "enemy summons"},
	"j": {"allied_summons", "allied summons"},
	"S": {"enemy_static_summons", "enemy static or non-playable summons"},
	"s": {"allied_static_summons", "allied static or non-playable summons"},
}

type targetConditionDefinition struct {
	Name, Description, Operator, ReferenceKind string
}

var targetConditionNames = map[string]targetConditionDefinition{
	"B":   {"breed_is", "target has the specified breed", "equals", "breed"},
	"b":   {"breed_is_not", "target does not have the specified breed", "not_equals", "breed"},
	"E":   {"has_state", "subject has the specified spell state", "contains", "state"},
	"e":   {"does_not_have_state", "subject does not have the specified spell state", "not_contains", "state"},
	"F":   {"monster_is", "target is the specified monster", "equals", "monster"},
	"f":   {"monster_is_not", "target is not the specified monster", "not_equals", "monster"},
	"K":   {"is_killed", "target is killed by the triggering action", "equals", ""},
	"O":   {"is_triggering_spell_caster", "target is the caster of the triggering spell", "equals", ""},
	"P":   {"is_summon", "target is a summon", "equals", ""},
	"p":   {"is_not_summon", "target is not a summon", "not_equals", ""},
	"Q":   {"is_summoner", "target is the summoner of the triggering summon", "equals", ""},
	"q":   {"is_not_summoner", "target is not the summoner of the triggering summon", "not_equals", ""},
	"R":   {"is_revived", "target has been revived", "equals", ""},
	"r":   {"is_not_revived", "target has not been revived", "not_equals", ""},
	"T":   {"is_telefrag_target", "target is part of the triggering telefrag", "equals", ""},
	"U":   {"is_summoned_by_caster", "target was summoned by the caster", "equals", ""},
	"u":   {"is_not_summoned_by_caster", "target was not summoned by the caster", "not_equals", ""},
	"W":   {"is_portal", "target is a portal", "equals", ""},
	"Z":   {"creature_family_is", "target belongs to the specified creature family", "equals", "creature_family"},
	"z":   {"creature_family_is_not", "target does not belong to the specified creature family", "not_equals", "creature_family"},
	"V":   {"life_percent_at_most", "target life percentage is at most the specified value", "less_than_or_equal", ""},
	"v":   {"life_percent_above", "target life percentage is above the specified value", "greater_than", ""},
	"AP":  {"action_points_at_least", "target has at least the specified action points", "greater_than_or_equal", ""},
	"ap":  {"action_points_below", "target has fewer than the specified action points", "less_than", ""},
	"MP":  {"movement_points_at_least", "target has at least the specified movement points", "greater_than_or_equal", ""},
	"mp":  {"movement_points_below", "target has fewer than the specified movement points", "less_than", ""},
	"PB":  {"shield_points_at_least", "target has at least the specified shield points", "greater_than_or_equal", ""},
	"pb":  {"shield_points_below", "target has fewer than the specified shield points", "less_than", ""},
	"PR":  {"is_primary_target", "target is the primary target of the triggering action", "equals", ""},
	"pr":  {"is_not_primary_target", "target is not the primary target of the triggering action", "not_equals", ""},
	"o":   {"is_not_triggering_spell_caster", "target is not the caster of the triggering spell", "not_equals", ""},
	"x":   {"is_original_target", "target is the original target of the triggering action", "equals", ""},
	"Atq": {"attack_team", "fighters on the scenario attack team", "equals", ""},
	"Def": {"defense_team", "fighters on the scenario defense team", "equals", ""},
	"Sce": {"scenario_entities", "entities controlled by the fight scenario", "equals", ""},
}

var targetTokenPattern = regexp.MustCompile(`^([A-Za-z]+?)(-?\d+)?$`)

func mapTargetMask(raw string, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, languages map[string]mapping.LangDictUnity) mappedTargetMask {
	result := mappedTargetMask{Raw: raw, Logic: "selector_matches_and_all_conditions_match", Selectors: []mappedCombatCondition{}, Conditions: []mappedCombatCondition{}}
	if raw == "" {
		return result
	}
	for _, encoded := range strings.Split(raw, ",") {
		subject := "target"
		if strings.HasPrefix(encoded, "*") {
			subject = "caster"
			encoded = strings.TrimPrefix(encoded, "*")
		}
		if selector, ok := targetSelectorNames[encoded]; ok {
			result.Selectors = append(result.Selectors, mappedCombatCondition{Code: encoded, Name: selector[0], Description: selector[1], Subject: subject})
			continue
		}
		matches := targetTokenPattern.FindStringSubmatch(encoded)
		if len(matches) == 0 {
			panic(fmt.Sprintf("unmapped target-mask token %q", encoded))
		}
		prefix := matches[1]
		definition, ok := targetConditionNames[prefix]
		if !ok {
			panic(fmt.Sprintf("unmapped target-mask prefix %q in %q", prefix, raw))
		}
		condition := mappedCombatCondition{Code: encoded, Name: definition.Name, Description: definition.Description, Subject: subject, Operator: definition.Operator, ReferenceType: definition.ReferenceKind}
		if matches[2] != "" {
			value, _ := strconv.Atoi(matches[2])
			condition.Value = &value
			condition.Reference = targetReference(definition.ReferenceKind, value, states, monsters, breeds, languages)
		}
		result.Conditions = append(result.Conditions, condition)
	}
	return result
}

func targetReference(kind string, id int, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, languages map[string]mapping.LangDictUnity) *mappedNamedReference {
	switch kind {
	case "state":
		state, ok := states[id]
		return pointerToNamedReference(namedReference(id, state.NameID, ok, languages))
	case "monster":
		monster, ok := monsters[id]
		return pointerToNamedReference(namedReference(id, monster.NameID, ok, languages))
	case "breed":
		breed, ok := breeds[id]
		return pointerToNamedReference(namedReference(id, breed.ShortNameID, ok, languages))
	default:
		return nil
	}
}

func pointerToNamedReference(value mappedNamedReference) *mappedNamedReference { return &value }

type mappedTrigger struct {
	Code        string                `json:"code"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Subject     string                `json:"subject,omitempty"`
	Operator    string                `json:"operator,omitempty"`
	Parameter   *int                  `json:"parameter,omitempty"`
	Reference   *mappedNamedReference `json:"reference,omitempty"`
	TargetMask  *mappedTargetMask     `json:"target_mask,omitempty"`
}

type mappedTriggers struct {
	Raw    string          `json:"raw"`
	Logic  string          `json:"logic"`
	Events []mappedTrigger `json:"events"`
}

var triggerNames = map[string][2]string{
	"I": {"immediate", "applies immediately when the spell is cast"}, "D": {"damage_received", "target receives damage"},
	"DA": {"air_damage_received", "target receives Air damage"}, "DBA": {"damage_received_from_ally", "target receives damage from an ally"},
	"DBE": {"damage_received_from_enemy", "target receives damage from an enemy"}, "DCAC": {"close_combat_damage_received", "target receives close-combat damage"},
	"DCCBE": {"critical_damage_received_from_enemy", "target receives a critical hit from an enemy"}, "DE": {"earth_damage_received", "target receives Earth damage"},
	"DF": {"fire_damage_received", "target receives Fire damage"}, "DG": {"glyph_damage_received", "target receives glyph damage"},
	"DI": {"indirect_damage_received", "target receives indirect damage"}, "DIS": {"shield_damage_received", "target loses shield points"},
	"DM": {"melee_damage_received", "target receives damage at melee range"}, "DN": {"neutral_damage_received", "target receives Neutral damage"},
	"DR": {"ranged_damage_received", "target receives damage at range"}, "DS": {"spell_damage_received", "target receives spell damage"},
	"DT": {"trap_damage_received", "target receives trap damage"}, "DTB": {"pushback_damage_received", "target receives pushback damage"},
	"DTE": {"erosion_damage_received", "target loses maximum life through erosion"}, "DV": {"life_stolen", "life is stolen from the target"},
	"DW": {"water_damage_received", "target receives Water damage"}, "PD": {"pushback", "target is pushed"},
	"PMD": {"pushback_damage", "target takes pushback damage"}, "PPD": {"pull", "target is pulled"},
	"PDT": {"pushback_collision", "a push causes a collision"}, "PO": {"portal_used", "an entity uses a portal"},
	"PST": {"position_swap", "two entities swap positions"}, "PT": {"attract", "target is attracted"},
	"TB": {"turn_begin", "target's turn begins"}, "TE": {"turn_end", "target's turn ends"},
	"T": {"teleport", "target is teleported"}, "TP": {"telefrag", "a telefrag is generated"},
	"M": {"movement", "target moves"}, "MA": {"movement_away", "target moves away"},
	"MPA": {"movement_points_changed", "target's movement points change"}, "MS": {"movement_start", "target starts moving"},
	"CMPA": {"caster_movement_points_changed", "caster's movement points change"}, "CMPARR": {"caster_movement_stopped", "caster stops moving"},
	"CMPAS": {"caster_movement_points_spent", "caster spends movement points"}, "CMPDEP": {"caster_movement_started", "caster starts moving"},
	"CCMPARR": {"caster_movement_stopped_by_target", "caster stops moving because of the target"}, "CCMPDEP": {"caster_moves_from_target", "caster starts moving from the target"},
	"APA": {"action_points_changed", "target's action points change"}, "CAP": {"caster_action_points_changed", "caster's action points change"},
	"CAPA": {"caster_action_points_added", "caster gains action points"}, "CAPAS": {"caster_action_points_spent", "caster spends action points"},
	"H": {"healed", "target is healed"}, "V": {"life_changed", "target's life changes"},
	"VA": {"life_gained", "target gains life"}, "VE": {"life_lost", "target loses life"}, "VM": {"maximum_life_changed", "target's maximum life changes"},
	"K": {"fighter_killed", "a fighter is killed"}, "KE": {"enemy_killed", "an enemy is killed"},
	"KEDT": {"enemy_killed_by_trap", "an enemy is killed by trap damage"},
	"KEWS": {"enemy_killed_by_weapon_or_spell", "an enemy is killed by weapon or spell damage"},
	"KHA":  {"ally_player_killed", "an allied player is killed"}, "KHE": {"enemy_player_killed", "an enemy player is killed"},
	"KIE": {"enemy_summon_killed", "an enemy summon is killed"}, "KMA": {"allied_monster_killed", "an allied monster is killed"},
	"KME": {"enemy_monster_killed", "an enemy monster is killed"}, "KWW": {"summon_killed", "a summon is killed"},
	"R": {"revived", "a fighter is revived"}, "LPU": {"portal_state_changed", "a portal becomes usable or unusable"},
	"EON": {"state_entered", "the specified spell state is entered"}, "EOFF": {"state_left", "the specified spell state is left"},
	"EACT": {"effect_activated", "the specified effect is activated"}, "ION": {"invisibility_started", "target becomes invisible"},
	"IOFF": {"invisibility_ended", "target becomes visible"}, "OEIOFF": {"other_invisibility_ended", "another fighter becomes visible"},
	"CI": {"caster_invisible", "caster becomes invisible"}, "CIOFF": {"caster_invisibility_ended", "caster becomes visible"},
	"CD": {"caster_deals_damage", "caster deals damage"}, "CDA": {"caster_deals_air_damage", "caster deals Air damage"},
	"CDBA": {"caster_damages_ally", "caster damages an ally"}, "CDBE": {"caster_damages_enemy", "caster damages an enemy"},
	"CDCAC": {"caster_deals_close_combat_damage", "caster deals close-combat damage"}, "CDE": {"caster_deals_earth_damage", "caster deals Earth damage"},
	"CDF": {"caster_deals_fire_damage", "caster deals Fire damage"}, "CDM": {"caster_deals_melee_damage", "caster deals damage at melee range"},
	"CDN": {"caster_deals_neutral_damage", "caster deals Neutral damage"}, "CDR": {"caster_deals_ranged_damage", "caster deals damage at range"},
	"CDS": {"caster_deals_spell_damage", "caster deals spell damage"}, "CDW": {"caster_deals_water_damage", "caster deals Water damage"},
	"CC": {"critical_hit", "a critical hit occurs"}, "CH": {"caster_heals", "caster heals a fighter"},
	"CS": {"spell_cast", "caster casts a spell"}, "CT": {"trap_triggered", "caster's trap is triggered"},
	"CPD": {"caster_pushes", "caster pushes a fighter"}, "CPT": {"caster_attracts", "caster attracts a fighter"},
	"P": {"position_changed", "target's position changes"}, "SREF": {"reflected_spell", "a spell is reflected"},
	"TR": {"spell_triggered", "the specified spell is triggered"}, "X": {"other_fighter_died", "another fighter dies"},
}

var parameterizedTrigger = regexp.MustCompile(`^(EACT|EON|EOFF|TR)(\d+)$`)
var effectCountTrigger = regexp.MustCompile(`^EC:(<=|>=|=|<|>)(\d+):(.+)$`)

func mapTriggers(raw string, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, spells map[int]rawSpell, languages map[string]mapping.LangDictUnity) mappedTriggers {
	result := mappedTriggers{Raw: raw, Logic: "any", Events: []mappedTrigger{}}
	if raw == "" {
		return result
	}
	for _, code := range strings.Split(raw, "|") {
		if matches := effectCountTrigger.FindStringSubmatch(code); len(matches) != 0 {
			count, _ := strconv.Atoi(matches[2])
			mask := mapTargetMask(matches[3], states, monsters, breeds, languages)
			result.Events = append(result.Events, mappedTrigger{
				Code: code, Name: "matching_fighter_count",
				Description: "the number of fighters matching the target mask satisfies the comparison",
				Subject:     "fight", Operator: comparisonOperatorName(matches[1]), Parameter: &count, TargetMask: &mask,
			})
			continue
		}
		if strings.HasPrefix(code, "EK:") {
			mask := mapTargetMask(strings.TrimPrefix(code, "EK:"), states, monsters, breeds, languages)
			result.Events = append(result.Events, mappedTrigger{Code: code, Name: "target_matches", Description: "triggering target matches the embedded target mask", TargetMask: &mask})
			continue
		}
		base := code
		var parameter *int
		if matches := parameterizedTrigger.FindStringSubmatch(code); len(matches) != 0 {
			base = matches[1]
			value, _ := strconv.Atoi(matches[2])
			parameter = &value
		}
		lookup := base
		subject := "target"
		if strings.HasPrefix(base, "X") && base != "X" {
			if _, ok := triggerNames[strings.TrimPrefix(base, "X")]; ok {
				lookup = strings.TrimPrefix(base, "X")
				subject = "other_fighter"
			}
		}
		definition, ok := triggerNames[lookup]
		if !ok {
			panic(fmt.Sprintf("unmapped trigger %q", code))
		}
		var reference *mappedNamedReference
		if parameter != nil {
			switch base {
			case "EACT", "EON", "EOFF":
				state, found := states[*parameter]
				reference = pointerToNamedReference(namedReference(*parameter, state.NameID, found, languages))
			case "TR":
				spell, found := spells[*parameter]
				reference = pointerToNamedReference(namedReference(*parameter, spell.NameID, found, languages))
			}
		}
		name, description := definition[0], definition[1]
		if subject == "other_fighter" {
			name = "other_fighter_" + name
			description = strings.Replace(description, "target", "another fighter", 1)
		}
		result.Events = append(result.Events, mappedTrigger{Code: code, Name: name, Description: description, Subject: subject, Parameter: parameter, Reference: reference})
	}
	return result
}

func comparisonOperatorName(operator string) string {
	switch operator {
	case "=":
		return "equals"
	case "<":
		return "less_than"
	case ">":
		return "greater_than"
	case "<=":
		return "less_than_or_equal"
	case ">=":
		return "greater_than_or_equal"
	default:
		panic(fmt.Sprintf("unmapped comparison operator %q", operator))
	}
}
