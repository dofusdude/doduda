package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	mapping "github.com/dofusdude/dodumap"
)

type unityArray[T any] struct {
	Array []T `json:"Array"`
}

type unityReference struct {
	Data json.RawMessage `json:"data"`
	Type struct {
		Class string `json:"class"`
	} `json:"type"`
}

type unityAsset struct {
	References struct {
		RefIDs []unityReference `json:"RefIds"`
	} `json:"references"`
}

type flexibleInt int

func (i *flexibleInt) UnmarshalJSON(data []byte) error {
	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*i = flexibleInt(number)
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == "" {
		*i = 0
		return nil
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("parse integer %q: %w", value, err)
	}
	*i = flexibleInt(number)
	return nil
}

type rawBreed struct {
	ID                    int                         `json:"id"`
	ShortNameID           flexibleInt                 `json:"shortNameId"`
	DescriptionID         flexibleInt                 `json:"descriptionId"`
	GameplayDescriptionID flexibleInt                 `json:"gameplayDescriptionId"`
	SpellIDs              unityArray[int]             `json:"breedSpellsId"`
	Complexity            int                         `json:"complexity"`
	SortIndex             int                         `json:"sortIndex"`
	CreatureBonesID       int                         `json:"creatureBonesId"`
	MaleLook              string                      `json:"maleLook"`
	FemaleLook            string                      `json:"femaleLook"`
	MaleColors            unityArray[int]             `json:"maleColors"`
	FemaleColors          unityArray[int]             `json:"femaleColors"`
	Roles                 unityArray[rawBreedRoleUse] `json:"breedRoles"`
	StrengthCosts         unityArray[rawStatCost]     `json:"statsPointsForStrength"`
	IntelligenceCosts     unityArray[rawStatCost]     `json:"statsPointsForIntelligence"`
	ChanceCosts           unityArray[rawStatCost]     `json:"statsPointsForChance"`
	AgilityCosts          unityArray[rawStatCost]     `json:"statsPointsForAgility"`
	VitalityCosts         unityArray[rawStatCost]     `json:"statsPointsForVitality"`
	WisdomCosts           unityArray[rawStatCost]     `json:"statsPointsForWisdom"`
}

type rawBreedRoleUse struct {
	RoleID int `json:"roleId"`
	Value  int `json:"value"`
	Order  int `json:"order"`
}

type rawBreedRole struct {
	ID            int         `json:"id"`
	NameID        flexibleInt `json:"nameId"`
	DescriptionID flexibleInt `json:"descriptionId"`
	AssetID       int         `json:"assetId"`
	Color         int         `json:"color"`
}

type rawStatCost struct {
	Values unityArray[int] `json:"values"`
}

type rawSpell struct {
	ID              int                `json:"id"`
	NameID          flexibleInt        `json:"nameId"`
	DescriptionID   flexibleInt        `json:"descriptionId"`
	TypeID          int                `json:"typeId"`
	IconID          int                `json:"iconId"`
	LevelIDs        unityArray[int]    `json:"spellLevels"`
	BasePreviewZone rawZoneDescription `json:"basePreviewZoneDescr"`
}

type rawSpellType struct {
	ID          int         `json:"id"`
	ShortNameID flexibleInt `json:"shortNameId"`
	LongNameID  flexibleInt `json:"longNameId"`
}

type rawSpellState struct {
	ID     int         `json:"id"`
	NameID flexibleInt `json:"nameId"`
}

type rawSpellLevel struct {
	ID                     int                        `json:"id"`
	SpellID                int                        `json:"spellId"`
	Grade                  int                        `json:"grade"`
	APCost                 int                        `json:"apCost"`
	MinRange               int                        `json:"minRange"`
	Range                  int                        `json:"range"`
	CriticalHitProbability int                        `json:"criticalHitProbability"`
	MaxStack               int                        `json:"maxStack"`
	MaxCastPerTurn         int                        `json:"maxCastPerTurn"`
	MaxCastPerTarget       int                        `json:"maxCastPerTarget"`
	MinCastInterval        int                        `json:"minCastInterval"`
	InitialCooldown        int                        `json:"initialCooldown"`
	GlobalCooldown         int                        `json:"globalCooldown"`
	MinPlayerLevel         int                        `json:"minPlayerLevel"`
	StatesCriterion        string                     `json:"statesCriterion"`
	Effects                unityArray[rawSpellEffect] `json:"effects"`
	CriticalEffects        unityArray[rawSpellEffect] `json:"criticalEffect"`
	PreviewZones           unityArray[rawPreviewZone] `json:"previewZones"`
}

type rawSpellEffect struct {
	EffectID              int                `json:"effectId"`
	MinimumValue          int                `json:"diceNum"`
	MaximumValue          int                `json:"diceSide"`
	Value                 int                `json:"value"`
	BaseEffectID          int                `json:"baseEffectId"`
	EffectElement         int                `json:"effectElement"`
	Dispellable           int                `json:"dispellable"`
	SpellID               int                `json:"spellId"`
	Duration              int                `json:"duration"`
	Delay                 int                `json:"delay"`
	Random                float64            `json:"random"`
	Group                 int                `json:"group"`
	Order                 int                `json:"order"`
	TargetMask            string             `json:"targetMask"`
	Triggers              string             `json:"triggers"`
	EffectTriggerDuration int                `json:"effectTriggerDuration"`
	Zone                  rawZoneDescription `json:"zoneDescr"`
}

func (effect rawSpellEffect) possibleEffect() mapping.JSONGameItemPossibleEffectUnity {
	return mapping.JSONGameItemPossibleEffectUnity{
		EffectId: effect.EffectID, MinimumValue: effect.MinimumValue, MaximumValue: effect.MaximumValue,
		Value: effect.Value, BaseEffectId: effect.BaseEffectID, EffectElement: effect.EffectElement,
		Dispellable: effect.Dispellable, SpellId: effect.SpellID, Duration: effect.Duration,
	}
}

type rawPreviewZone struct {
	ID             int                `json:"id"`
	ActivationMask string             `json:"activationMask"`
	CasterMask     string             `json:"casterMask"`
	Hidden         int                `json:"isPreviewZoneHidden"`
	ActivationZone rawZoneDescription `json:"activationZoneDescr"`
	DisplayZone    rawZoneDescription `json:"displayZoneDescr"`
}

type rawMonster struct {
	ID                int                         `json:"id"`
	NameID            flexibleInt                 `json:"nameId"`
	GfxID             int                         `json:"gfxId"`
	RaceID            int                         `json:"race"`
	Spells            unityArray[int]             `json:"spells"`
	SpellGrades       unityArray[string]          `json:"spellGrades"`
	Grades            unityArray[rawMonsterGrade] `json:"grades"`
	SubareaIDs        unityArray[int]             `json:"subareas"`
	FavoriteSubareaID int                         `json:"favoriteSubareaId"`
	Drops             unityArray[rawMonsterDrop]  `json:"drops"`
}

type rawMonsterDrop struct {
	ObjectID               int     `json:"objectId"`
	PercentDropForGrade1   float64 `json:"percentDropForGrade1"`
	PercentDropForGrade2   float64 `json:"percentDropForGrade2"`
	PercentDropForGrade3   float64 `json:"percentDropForGrade3"`
	PercentDropForGrade4   float64 `json:"percentDropForGrade4"`
	PercentDropForGrade5   float64 `json:"percentDropForGrade5"`
	Criterions             string  `json:"criterions"`
	HiddenIfInvalid        int     `json:"hiddenIfInvalidCriterions"`
	DisableDropModificator int     `json:"disableDropModificator"`
}

type rawMonsterCharacteristics struct {
	APRemoval         int `json:"aPRemoval"`
	Agility           int `json:"agility"`
	AirResistance     int `json:"airResistance"`
	BonusAirDamage    int `json:"bonusAirDamage"`
	BonusEarthDamage  int `json:"bonusEarthDamage"`
	BonusFireDamage   int `json:"bonusFireDamage"`
	BonusWaterDamage  int `json:"bonusWaterDamage"`
	Chance            int `json:"chance"`
	EarthResistance   int `json:"earthResistance"`
	FireResistance    int `json:"fireResistance"`
	Intelligence      int `json:"intelligence"`
	LifePoints        int `json:"lifePoints"`
	NeutralResistance int `json:"neutralResistance"`
	Strength          int `json:"strength"`
	TackleBlock       int `json:"tackleBlock"`
	TackleEvade       int `json:"tackleEvade"`
	WaterResistance   int `json:"waterResistance"`
	Wisdom            int `json:"wisdom"`
}

type rawMonsterGrade struct {
	ActionPoints         int                       `json:"actionPoints"`
	Agility              int                       `json:"agility"`
	AirResistance        int                       `json:"airResistance"`
	BonusCharacteristics rawMonsterCharacteristics `json:"bonusCharacteristics"`
	BonusRange           int                       `json:"bonusRange"`
	Chance               int                       `json:"chance"`
	DamageReflect        int                       `json:"damageReflect"`
	EarthResistance      int                       `json:"earthResistance"`
	FireResistance       int                       `json:"fireResistance"`
	Grade                int                       `json:"grade"`
	GradeXP              int                       `json:"gradeXp"`
	Intelligence         int                       `json:"intelligence"`
	Level                int                       `json:"level"`
	LifePoints           int                       `json:"lifePoints"`
	MovementPoints       int                       `json:"movementPoints"`
	NeutralResistance    int                       `json:"neutralResistance"`
	PADodge              int                       `json:"paDodge"`
	PMDodge              int                       `json:"pmDodge"`
	StartingSpellID      int                       `json:"startingSpellId"`
	Strength             int                       `json:"strength"`
	Vitality             int                       `json:"vitality"`
	WaterResistance      int                       `json:"waterResistance"`
	Wisdom               int                       `json:"wisdom"`
}

type rawMonsterRace struct {
	ID          int         `json:"id"`
	NameID      flexibleInt `json:"nameId"`
	SuperRaceID int         `json:"superRaceId"`
}

type rawMonsterSuperRace struct {
	ID     int         `json:"id"`
	NameID flexibleInt `json:"nameId"`
}

type mappedNamedReference struct {
	AnkamaID int               `json:"ankama_id"`
	Name     map[string]string `json:"name"`
}

type mappedClass struct {
	AnkamaID            int                         `json:"ankama_id"`
	Name                map[string]string           `json:"name"`
	Description         map[string]string           `json:"description"`
	GameplayDescription map[string]string           `json:"gameplay_description"`
	Complexity          int                         `json:"complexity"`
	SortIndex           int                         `json:"sort_index"`
	CreatureBonesID     int                         `json:"creature_bones_id"`
	Looks               mappedClassLooks            `json:"looks"`
	Roles               []mappedClassRole           `json:"roles"`
	CharacteristicCosts map[string][]mappedStatCost `json:"characteristic_costs"`
	Spells              []mappedSpell               `json:"spells"`
}

type mappedClassLooks struct {
	Male   mappedClassAppearance `json:"male"`
	Female mappedClassAppearance `json:"female"`
}

type mappedClassAppearance struct {
	BonesID       int         `json:"bones_id"`
	SkinIDs       []int       `json:"skin_ids"`
	Scale         mappedScale `json:"scale_percent"`
	DefaultColors []int       `json:"default_colors"`
}

type mappedScale struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type mappedClassRole struct {
	AnkamaID    int               `json:"ankama_id"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	Value       int               `json:"value"`
	Order       int               `json:"order"`
	AssetID     int               `json:"asset_id,omitempty"`
	Color       int               `json:"color,omitempty"`
}

type mappedStatCost struct {
	From int `json:"from"`
	Cost int `json:"cost"`
}

type mappedSpellType struct {
	AnkamaID  int               `json:"ankama_id"`
	ShortName map[string]string `json:"short_name"`
	LongName  map[string]string `json:"long_name"`
}

type mappedStateCriterion struct {
	Raw                string            `json:"raw"`
	Templated          map[string]string `json:"templated"`
	StateIDs           []int             `json:"state_ids"`
	UnresolvedStateIDs []int             `json:"unresolved_state_ids"`
}

type mappedSpellLevel struct {
	AnkamaID               int                   `json:"ankama_id"`
	Grade                  int                   `json:"grade"`
	APCost                 int                   `json:"ap_cost"`
	MinRange               int                   `json:"min_range"`
	MaxRange               int                   `json:"max_range"`
	CriticalHitProbability int                   `json:"critical_hit_probability"`
	MaxStack               int                   `json:"max_stack"`
	MaxCastPerTurn         int                   `json:"max_cast_per_turn"`
	MaxCastPerTarget       int                   `json:"max_cast_per_target"`
	MinCastInterval        int                   `json:"min_cast_interval"`
	InitialCooldown        int                   `json:"initial_cooldown"`
	GlobalCooldown         int                   `json:"global_cooldown"`
	MinPlayerLevel         int                   `json:"min_player_level"`
	StatesCriterion        *mappedStateCriterion `json:"states_criterion"`
	PreviewZones           []mappedPreviewZone   `json:"preview_zones"`
	Effects                []mappedSpellEffect   `json:"effects"`
	CriticalEffects        []mappedSpellEffect   `json:"critical_effects"`
}

type mappedPreviewZone struct {
	AnkamaID       int              `json:"ankama_id"`
	ActivationMask mappedTargetMask `json:"activation_mask"`
	CasterMask     mappedTargetMask `json:"caster_mask"`
	Hidden         bool             `json:"hidden"`
	ActivationZone mappedZone       `json:"activation_zone"`
	DisplayZone    mappedZone       `json:"display_zone"`
}

type mappedSpellEffect struct {
	mapping.MappedMultilangEffect
	Zone              mappedZone       `json:"zone"`
	TargetMask        mappedTargetMask `json:"target_mask"`
	Triggers          mappedTriggers   `json:"triggers"`
	Duration          int              `json:"duration"`
	Delay             int              `json:"delay"`
	Random            float64          `json:"random"`
	Group             int              `json:"group"`
	Order             int              `json:"order"`
	Dispellable       int              `json:"dispellable"`
	TriggeredDuration int              `json:"triggered_duration"`
}

type mappedSpell struct {
	AnkamaID        int                    `json:"ankama_id"`
	Name            map[string]string      `json:"name"`
	Description     map[string]string      `json:"description"`
	Type            mappedSpellType        `json:"type"`
	IconID          int                    `json:"icon_id"`
	BasePreviewZone mappedZone             `json:"base_preview_zone"`
	Classes         []mappedNamedReference `json:"classes"`
	Monsters        []mappedNamedReference `json:"monsters"`
	Levels          []mappedSpellLevel     `json:"levels"`
}

type mappedMonsterRace struct {
	AnkamaID  int                  `json:"ankama_id"`
	Name      map[string]string    `json:"name"`
	SuperRace mappedNamedReference `json:"super_race"`
}

type mappedMonsterSpellGrade struct {
	MonsterGrade  int `json:"monster_grade"`
	MinSpellGrade int `json:"min_spell_grade"`
	MaxSpellGrade int `json:"max_spell_grade"`
}

type mappedMonsterSpell struct {
	Spell       mappedSpell               `json:"spell"`
	GradeRanges []mappedMonsterSpellGrade `json:"grade_ranges"`
}

type mappedStartingSpell struct {
	SpellLevelID int               `json:"spell_level_id"`
	SpellID      int               `json:"spell_id,omitempty"`
	Grade        int               `json:"grade,omitempty"`
	Name         map[string]string `json:"name"`
}

type mappedMonsterCharacteristics struct {
	APRemoval         int `json:"ap_removal"`
	Agility           int `json:"agility"`
	AirResistance     int `json:"air_resistance"`
	BonusAirDamage    int `json:"bonus_air_damage"`
	BonusEarthDamage  int `json:"bonus_earth_damage"`
	BonusFireDamage   int `json:"bonus_fire_damage"`
	BonusWaterDamage  int `json:"bonus_water_damage"`
	Chance            int `json:"chance"`
	EarthResistance   int `json:"earth_resistance"`
	FireResistance    int `json:"fire_resistance"`
	Intelligence      int `json:"intelligence"`
	LifePoints        int `json:"life_points"`
	NeutralResistance int `json:"neutral_resistance"`
	Strength          int `json:"strength"`
	TackleBlock       int `json:"tackle_block"`
	TackleEvade       int `json:"tackle_evade"`
	WaterResistance   int `json:"water_resistance"`
	Wisdom            int `json:"wisdom"`
}

type mappedMonsterGrade struct {
	ActionPoints         int                          `json:"action_points"`
	Agility              int                          `json:"agility"`
	AirResistance        int                          `json:"air_resistance"`
	BonusCharacteristics mappedMonsterCharacteristics `json:"bonus_characteristics"`
	BonusRange           int                          `json:"bonus_range"`
	Chance               int                          `json:"chance"`
	DamageReflect        int                          `json:"damage_reflect"`
	EarthResistance      int                          `json:"earth_resistance"`
	FireResistance       int                          `json:"fire_resistance"`
	Grade                int                          `json:"grade"`
	GradeXP              int                          `json:"grade_xp"`
	Intelligence         int                          `json:"intelligence"`
	Level                int                          `json:"level"`
	LifePoints           int                          `json:"life_points"`
	MovementPoints       int                          `json:"movement_points"`
	NeutralResistance    int                          `json:"neutral_resistance"`
	PADodge              int                          `json:"pa_dodge"`
	PMDodge              int                          `json:"pm_dodge"`
	StartingSpell        *mappedStartingSpell         `json:"starting_spell"`
	Strength             int                          `json:"strength"`
	Vitality             int                          `json:"vitality"`
	WaterResistance      int                          `json:"water_resistance"`
	Wisdom               int                          `json:"wisdom"`
}

type mappedMonster struct {
	AnkamaID       int                   `json:"ankama_id"`
	Name           map[string]string     `json:"name"`
	GfxID          int                   `json:"gfx_id"`
	Race           mappedMonsterRace     `json:"race"`
	Grades         []mappedMonsterGrade  `json:"grades"`
	Spells         []mappedMonsterSpell  `json:"spells"`
	SpawnLocations []mappedSpawnLocation `json:"spawn_locations"`
	Drops          []mappedMonsterDrop   `json:"drops"`
}

type mappedMonsterDrop struct {
	Item          mappedNamedReference `json:"item"`
	RatesPercent  []mappedGradeRate    `json:"rates_percent"`
	HasConditions bool                 `json:"has_conditions"`
}

type mappedGradeRate struct {
	Grade   int     `json:"grade"`
	Percent float64 `json:"percent"`
}

func readUnityRecords[T any](dir, filename, className string) ([]T, error) {
	file, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return nil, err
	}
	var asset unityAsset
	if err := json.Unmarshal(file, &asset); err != nil {
		return nil, fmt.Errorf("decode %s: %w", filename, err)
	}
	records := make([]T, 0)
	for _, reference := range asset.References.RefIDs {
		if reference.Type.Class != className {
			continue
		}
		var record T
		if err := json.Unmarshal(reference.Data, &record); err != nil {
			return nil, fmt.Errorf("decode %s %s record: %w", filename, className, err)
		}
		records = append(records, record)
	}
	return records, nil
}

func localizedText(id flexibleInt, languages map[string]mapping.LangDictUnity) map[string]string {
	result := make(map[string]string, len(languages))
	for language, dictionary := range languages {
		if value, ok := dictionary.Texts[int(id)]; ok {
			result[language] = value
		}
	}
	return result
}

// mapSpellEffects formats spell effects with the same localized representation
// used by the existing item and set mappers.
func mapSpellEffects(gameData *mapping.JSONGameDataUnity, effects []rawSpellEffect, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, spells map[int]rawSpell, languages *map[string]mapping.LangDictUnity) []mappedSpellEffect {
	effectPointers := make([]*mapping.JSONGameItemPossibleEffectUnity, len(effects))
	for index := range effects {
		possible := effects[index].possibleEffect()
		effectPointers[index] = &possible
	}
	mappedGroups := mapping.ParseEffectsUnity(gameData, [][]*mapping.JSONGameItemPossibleEffectUnity{effectPointers}, languages)
	if len(mappedGroups) == 0 {
		return []mappedSpellEffect{}
	}
	mappedEffects := make([]mappedSpellEffect, 0, len(mappedGroups[0]))
	for index, effect := range mappedGroups[0] {
		if effect != nil {
			raw := effects[index]
			mappedEffects = append(mappedEffects, mappedSpellEffect{
				MappedMultilangEffect: *effect, Zone: mapZoneDescription(raw.Zone),
				TargetMask: mapTargetMask(raw.TargetMask, states, monsters, breeds, *languages),
				Triggers:   mapTriggers(raw.Triggers, states, monsters, breeds, spells, *languages), Duration: raw.Duration,
				Delay: raw.Delay, Random: raw.Random, Group: raw.Group, Order: raw.Order,
				Dispellable: raw.Dispellable, TriggeredDuration: raw.EffectTriggerDuration,
			})
		}
	}
	return mappedEffects
}

func mapPreviewZones(raw []rawPreviewZone, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, languages map[string]mapping.LangDictUnity) []mappedPreviewZone {
	result := make([]mappedPreviewZone, 0, len(raw))
	for _, zone := range raw {
		result = append(result, mappedPreviewZone{
			AnkamaID:       zone.ID,
			ActivationMask: mapTargetMask(zone.ActivationMask, states, monsters, breeds, languages),
			CasterMask:     mapTargetMask(zone.CasterMask, states, monsters, breeds, languages),
			Hidden:         zone.Hidden != 0, ActivationZone: mapZoneDescription(zone.ActivationZone),
			DisplayZone: mapZoneDescription(zone.DisplayZone),
		})
	}
	return result
}

func namedReference(id int, nameID flexibleInt, resolved bool, languages map[string]mapping.LangDictUnity) mappedNamedReference {
	name := map[string]string{}
	if resolved {
		name = localizedText(nameID, languages)
	}
	return mappedNamedReference{AnkamaID: id, Name: name}
}

var stateCriterionToken = regexp.MustCompile(`HS([=!])(\d+)`)

type stateCriterionWords struct{ Has, HasNot, And, Or string }

var stateCriterionTranslations = map[string]stateCriterionWords{
	"de": {"hat den Zustand %s", "hat nicht den Zustand %s", " und ", " oder "},
	"en": {"has state %s", "does not have state %s", " and ", " or "},
	"es": {"tiene el estado %s", "no tiene el estado %s", " y ", " o "},
	"fr": {"a l'état %s", "n'a pas l'état %s", " et ", " ou "},
	"pt": {"tem o estado %s", "não tem o estado %s", " e ", " ou "},
}

func mapStateCriterion(raw string, states map[int]rawSpellState, languages map[string]mapping.LangDictUnity) *mappedStateCriterion {
	if raw == "" {
		return nil
	}
	found := stateCriterionToken.FindAllStringSubmatch(raw, -1)
	stateIDs := make([]int, 0, len(found))
	unresolved := make([]int, 0)
	seen, missing := map[int]bool{}, map[int]bool{}
	for _, token := range found {
		id, _ := strconv.Atoi(token[2])
		if !seen[id] {
			stateIDs = append(stateIDs, id)
			seen[id] = true
		}
		if _, ok := states[id]; !ok && !missing[id] {
			unresolved = append(unresolved, id)
			missing[id] = true
		}
	}
	templated := make(map[string]string, len(languages))
	for language := range languages {
		words, ok := stateCriterionTranslations[language]
		if !ok {
			words = stateCriterionTranslations["en"]
		}
		value := stateCriterionToken.ReplaceAllStringFunc(raw, func(token string) string {
			parts := stateCriterionToken.FindStringSubmatch(token)
			id, _ := strconv.Atoi(parts[2])
			name := strconv.Itoa(id)
			if state, ok := states[id]; ok {
				if localized := localizedText(state.NameID, languages)[language]; localized != "" {
					name = localized
				}
			}
			if parts[1] == "=" {
				return fmt.Sprintf(words.Has, name)
			}
			return fmt.Sprintf(words.HasNot, name)
		})
		value = strings.ReplaceAll(value, "&", words.And)
		value = strings.ReplaceAll(value, "|", words.Or)
		templated[language] = value
	}
	return &mappedStateCriterion{Raw: raw, Templated: templated, StateIDs: stateIDs, UnresolvedStateIDs: unresolved}
}

func mapSpellLevel(level rawSpellLevel, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, spells map[int]rawSpell, gameData *mapping.JSONGameDataUnity, languages *map[string]mapping.LangDictUnity) mappedSpellLevel {
	return mappedSpellLevel{
		AnkamaID: level.ID, Grade: level.Grade, APCost: level.APCost,
		MinRange: level.MinRange, MaxRange: level.Range,
		CriticalHitProbability: level.CriticalHitProbability, MaxStack: level.MaxStack,
		MaxCastPerTurn: level.MaxCastPerTurn, MaxCastPerTarget: level.MaxCastPerTarget,
		MinCastInterval: level.MinCastInterval, InitialCooldown: level.InitialCooldown,
		GlobalCooldown: level.GlobalCooldown, MinPlayerLevel: level.MinPlayerLevel,
		StatesCriterion: mapStateCriterion(level.StatesCriterion, states, *languages),
		PreviewZones:    mapPreviewZones(level.PreviewZones.Array, states, monsters, breeds, *languages),
		Effects:         mapSpellEffects(gameData, level.Effects.Array, states, monsters, breeds, spells, languages),
		CriticalEffects: mapSpellEffects(gameData, level.CriticalEffects.Array, states, monsters, breeds, spells, languages),
	}
}

func parseMonsterSpellGrades(encoded string) []mappedMonsterSpellGrade {
	parts := strings.Split(encoded, ";")
	ranges := make([]mappedMonsterSpellGrade, 0, len(parts))
	for index, part := range parts {
		bounds := strings.Split(part, ",")
		if len(bounds) != 2 {
			continue
		}
		minimum, minErr := strconv.Atoi(bounds[0])
		maximum, maxErr := strconv.Atoi(bounds[1])
		if minErr == nil && maxErr == nil {
			ranges = append(ranges, mappedMonsterSpellGrade{index + 1, minimum, maximum})
		}
	}
	return ranges
}

func mapCharacteristics(value rawMonsterCharacteristics) mappedMonsterCharacteristics {
	return mappedMonsterCharacteristics{
		APRemoval: value.APRemoval, Agility: value.Agility, AirResistance: value.AirResistance,
		BonusAirDamage: value.BonusAirDamage, BonusEarthDamage: value.BonusEarthDamage,
		BonusFireDamage: value.BonusFireDamage, BonusWaterDamage: value.BonusWaterDamage,
		Chance: value.Chance, EarthResistance: value.EarthResistance, FireResistance: value.FireResistance,
		Intelligence: value.Intelligence, LifePoints: value.LifePoints, NeutralResistance: value.NeutralResistance,
		Strength: value.Strength, TackleBlock: value.TackleBlock, TackleEvade: value.TackleEvade,
		WaterResistance: value.WaterResistance, Wisdom: value.Wisdom,
	}
}

func mapStatCosts(raw unityArray[rawStatCost]) []mappedStatCost {
	result := make([]mappedStatCost, 0, len(raw.Array))
	for _, entry := range raw.Array {
		if len(entry.Values.Array) >= 2 {
			result = append(result, mappedStatCost{From: entry.Values.Array[0], Cost: entry.Values.Array[1]})
		}
	}
	return result
}

func parseIntegerList(raw string) ([]int, error) {
	if raw == "" {
		return []int{}, nil
	}
	values := strings.Split(raw, ",")
	result := make([]int, 0, len(values))
	for _, value := range values {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("parse integer list value %q: %w", value, err)
		}
		result = append(result, parsed)
	}
	return result, nil
}

func mapClassAppearance(raw string, colors []int) (mappedClassAppearance, error) {
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(raw, "{"), "}"), "|")
	if len(parts) < 4 {
		return mappedClassAppearance{}, fmt.Errorf("invalid entity look %q", raw)
	}
	bonesID, err := strconv.Atoi(parts[0])
	if err != nil {
		return mappedClassAppearance{}, fmt.Errorf("parse entity look bones %q: %w", raw, err)
	}
	skins, err := parseIntegerList(parts[1])
	if err != nil {
		return mappedClassAppearance{}, fmt.Errorf("parse entity look skins %q: %w", raw, err)
	}
	scales, err := parseIntegerList(parts[3])
	if err != nil {
		return mappedClassAppearance{}, fmt.Errorf("parse entity look scale %q: %w", raw, err)
	}
	scale := mappedScale{X: 100, Y: 100}
	if len(scales) == 1 {
		scale.X, scale.Y = scales[0], scales[0]
	} else if len(scales) >= 2 {
		scale.X, scale.Y = scales[0], scales[1]
	}
	if skins == nil {
		skins = []int{}
	}
	if colors == nil {
		colors = []int{}
	}
	return mappedClassAppearance{BonesID: bonesID, SkinIDs: skins, Scale: scale, DefaultColors: colors}, nil
}

func mapMonsterGrade(grade rawMonsterGrade, levels map[int]rawSpellLevel, spells map[int]rawSpell, languages map[string]mapping.LangDictUnity) mappedMonsterGrade {
	var startingSpell *mappedStartingSpell
	if grade.StartingSpellID > 0 {
		if level, ok := levels[grade.StartingSpellID]; ok {
			if spell, ok := spells[level.SpellID]; ok {
				startingSpell = &mappedStartingSpell{
					SpellLevelID: grade.StartingSpellID,
					SpellID:      level.SpellID,
					Grade:        level.Grade,
					Name:         localizedText(spell.NameID, languages),
				}
			}
		}
	}
	return mappedMonsterGrade{
		ActionPoints: grade.ActionPoints, Agility: grade.Agility, AirResistance: grade.AirResistance,
		BonusCharacteristics: mapCharacteristics(grade.BonusCharacteristics), BonusRange: grade.BonusRange,
		Chance: grade.Chance, DamageReflect: grade.DamageReflect, EarthResistance: grade.EarthResistance,
		FireResistance: grade.FireResistance, Grade: grade.Grade, GradeXP: grade.GradeXP,
		Intelligence: grade.Intelligence, Level: grade.Level, LifePoints: grade.LifePoints,
		MovementPoints: grade.MovementPoints, NeutralResistance: grade.NeutralResistance,
		PADodge: grade.PADodge, PMDodge: grade.PMDodge, StartingSpell: startingSpell,
		Strength: grade.Strength, Vitality: grade.Vitality, WaterResistance: grade.WaterResistance, Wisdom: grade.Wisdom,
	}
}

func MapSpellEntitiesUnity(dir string, gameData *mapping.JSONGameDataUnity, languages map[string]mapping.LangDictUnity, spawnByMonster map[int][]mappedSpawnLocation) ([]mappedClass, []mappedMonster, []mappedSpell, error) {
	breeds, err := readUnityRecords[rawBreed](dir, "breeds.json", "BreedData")
	if err != nil {
		return nil, nil, nil, err
	}
	monsters, err := readUnityRecords[rawMonster](dir, "monsters.json", "MonsterData")
	if err != nil {
		return nil, nil, nil, err
	}
	spells, err := readUnityRecords[rawSpell](dir, "spells.json", "SpellData")
	if err != nil {
		return nil, nil, nil, err
	}
	spellLevels, err := readUnityRecords[rawSpellLevel](dir, "spell_levels.json", "SpellLevelData")
	if err != nil {
		return nil, nil, nil, err
	}
	spellTypes, err := readUnityRecords[rawSpellType](dir, "spell_types.json", "SpellTypeData")
	if err != nil {
		return nil, nil, nil, err
	}
	spellStates, err := readUnityRecords[rawSpellState](dir, "spell_states.json", "SpellStateData")
	if err != nil {
		return nil, nil, nil, err
	}
	races, err := readUnityRecords[rawMonsterRace](dir, "monster_races.json", "MonsterRaceData")
	if err != nil {
		return nil, nil, nil, err
	}
	superRaces, err := readUnityRecords[rawMonsterSuperRace](dir, "monsters_super_races.json", "MonsterSuperRaceData")
	if err != nil {
		return nil, nil, nil, err
	}
	breedRoles, err := readUnityRecords[rawBreedRole](dir, "breed_roles.json", "BreedRoleData")
	if err != nil {
		return nil, nil, nil, err
	}

	spellByID := make(map[int]rawSpell, len(spells))
	for _, value := range spells {
		spellByID[value.ID] = value
	}
	levelByID := make(map[int]rawSpellLevel, len(spellLevels))
	for _, value := range spellLevels {
		levelByID[value.ID] = value
	}
	stateByID := make(map[int]rawSpellState, len(spellStates))
	for _, value := range spellStates {
		stateByID[value.ID] = value
	}
	typeByID := make(map[int]rawSpellType, len(spellTypes))
	for _, value := range spellTypes {
		typeByID[value.ID] = value
	}
	raceByID := make(map[int]rawMonsterRace, len(races))
	for _, value := range races {
		raceByID[value.ID] = value
	}
	superRaceByID := make(map[int]rawMonsterSuperRace, len(superRaces))
	for _, value := range superRaces {
		superRaceByID[value.ID] = value
	}
	breedRoleByID := make(map[int]rawBreedRole, len(breedRoles))
	for _, value := range breedRoles {
		breedRoleByID[value.ID] = value
	}
	breedByID := make(map[int]rawBreed, len(breeds))
	for _, value := range breeds {
		breedByID[value.ID] = value
	}
	monsterByID := make(map[int]rawMonster, len(monsters))
	for _, value := range monsters {
		monsterByID[value.ID] = value
	}

	classRefsBySpell := make(map[int][]mappedNamedReference)
	classSpellIDs := make(map[int][]int, len(breeds))
	mappedClasses := make([]mappedClass, 0, len(breeds))
	for _, breed := range breeds {
		ref := namedReference(breed.ID, breed.ShortNameID, true, languages)
		for _, spellID := range breed.SpellIDs.Array {
			if _, found := spellByID[spellID]; found {
				classSpellIDs[breed.ID] = append(classSpellIDs[breed.ID], spellID)
				classRefsBySpell[spellID] = append(classRefsBySpell[spellID], ref)
			}
		}
		roles := make([]mappedClassRole, 0, len(breed.Roles.Array))
		for _, roleUse := range breed.Roles.Array {
			role, found := breedRoleByID[roleUse.RoleID]
			if !found {
				continue
			}
			roles = append(roles, mappedClassRole{
				AnkamaID: roleUse.RoleID, Name: localizedText(role.NameID, languages),
				Description: localizedText(role.DescriptionID, languages), Value: roleUse.Value,
				Order: roleUse.Order, AssetID: role.AssetID, Color: role.Color,
			})
		}
		male, err := mapClassAppearance(breed.MaleLook, breed.MaleColors.Array)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("map male appearance for class %d: %w", breed.ID, err)
		}
		female, err := mapClassAppearance(breed.FemaleLook, breed.FemaleColors.Array)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("map female appearance for class %d: %w", breed.ID, err)
		}
		mappedClasses = append(mappedClasses, mappedClass{
			AnkamaID: breed.ID, Name: ref.Name, Description: localizedText(breed.DescriptionID, languages),
			GameplayDescription: localizedText(breed.GameplayDescriptionID, languages), Complexity: breed.Complexity,
			SortIndex: breed.SortIndex, CreatureBonesID: breed.CreatureBonesID,
			Looks: mappedClassLooks{Male: male, Female: female},
			Roles: roles,
			CharacteristicCosts: map[string][]mappedStatCost{
				"strength": mapStatCosts(breed.StrengthCosts), "intelligence": mapStatCosts(breed.IntelligenceCosts),
				"chance": mapStatCosts(breed.ChanceCosts), "agility": mapStatCosts(breed.AgilityCosts),
				"vitality": mapStatCosts(breed.VitalityCosts), "wisdom": mapStatCosts(breed.WisdomCosts),
			},
			Spells: []mappedSpell{},
		})
	}

	type pendingMonsterSpell struct {
		SpellID     int
		GradeRanges []mappedMonsterSpellGrade
	}
	monsterRefsBySpell := make(map[int][]mappedNamedReference)
	monsterSpellLinks := make(map[int][]pendingMonsterSpell, len(monsters))
	mappedMonsters := make([]mappedMonster, 0, len(monsters))
	for _, monster := range monsters {
		monsterRef := namedReference(monster.ID, monster.NameID, true, languages)
		for index, spellID := range monster.Spells.Array {
			if spellID <= 0 {
				continue
			}
			encoded := ""
			if index < len(monster.SpellGrades.Array) {
				encoded = monster.SpellGrades.Array[index]
			}
			if _, found := spellByID[spellID]; found {
				monsterSpellLinks[monster.ID] = append(monsterSpellLinks[monster.ID], pendingMonsterSpell{
					SpellID: spellID, GradeRanges: parseMonsterSpellGrades(encoded),
				})
				monsterRefsBySpell[spellID] = append(monsterRefsBySpell[spellID], monsterRef)
			}
		}
		mappedGrades := make([]mappedMonsterGrade, 0, len(monster.Grades.Array))
		for _, grade := range monster.Grades.Array {
			mappedGrades = append(mappedGrades, mapMonsterGrade(grade, levelByID, spellByID, languages))
		}
		race := mappedMonsterRace{AnkamaID: monster.RaceID, Name: map[string]string{}, SuperRace: mappedNamedReference{Name: map[string]string{}}}
		if rawRace, ok := raceByID[monster.RaceID]; ok {
			race.Name = localizedText(rawRace.NameID, languages)
			if superRace, ok := superRaceByID[rawRace.SuperRaceID]; ok {
				race.SuperRace = namedReference(superRace.ID, superRace.NameID, true, languages)
			}
		}
		spawnLocations := spawnByMonster[monster.ID]
		if spawnLocations == nil {
			spawnLocations = []mappedSpawnLocation{}
		}
		drops := make([]mappedMonsterDrop, 0, len(monster.Drops.Array))
		for _, drop := range monster.Drops.Array {
			item, found := gameData.Items[drop.ObjectID]
			if !found {
				continue
			}
			drops = append(drops, mappedMonsterDrop{
				Item: namedReference(drop.ObjectID, flexibleInt(item.NameId), true, languages),
				RatesPercent: []mappedGradeRate{
					{Grade: 1, Percent: drop.PercentDropForGrade1},
					{Grade: 2, Percent: drop.PercentDropForGrade2},
					{Grade: 3, Percent: drop.PercentDropForGrade3},
					{Grade: 4, Percent: drop.PercentDropForGrade4},
					{Grade: 5, Percent: drop.PercentDropForGrade5},
				},
				HasConditions: drop.Criterions != "",
			})
		}
		mappedMonsters = append(mappedMonsters, mappedMonster{
			AnkamaID: monster.ID, Name: monsterRef.Name, GfxID: monster.GfxID, Race: race,
			Grades: mappedGrades, Spells: []mappedMonsterSpell{}, SpawnLocations: spawnLocations, Drops: drops,
		})
	}

	levelsBySpell := make(map[int][]mappedSpellLevel)
	for _, level := range spellLevels {
		levelsBySpell[level.SpellID] = append(levelsBySpell[level.SpellID], mapSpellLevel(level, stateByID, monsterByID, breedByID, spellByID, gameData, &languages))
	}
	mappedSpells := make([]mappedSpell, 0, len(spells))
	for _, spell := range spells {
		mappedType := mappedSpellType{AnkamaID: spell.TypeID, ShortName: map[string]string{}, LongName: map[string]string{}}
		if rawType, ok := typeByID[spell.TypeID]; ok {
			mappedType.ShortName = localizedText(rawType.ShortNameID, languages)
			mappedType.LongName = localizedText(rawType.LongNameID, languages)
		}
		levels := levelsBySpell[spell.ID]
		if levels == nil {
			levels = []mappedSpellLevel{}
		}
		classes := classRefsBySpell[spell.ID]
		if classes == nil {
			classes = []mappedNamedReference{}
		}
		monsterRefs := monsterRefsBySpell[spell.ID]
		if monsterRefs == nil {
			monsterRefs = []mappedNamedReference{}
		}
		mappedSpells = append(mappedSpells, mappedSpell{
			AnkamaID: spell.ID, Name: localizedText(spell.NameID, languages),
			Description: localizedText(spell.DescriptionID, languages), Type: mappedType, IconID: spell.IconID,
			BasePreviewZone: mapZoneDescription(spell.BasePreviewZone), Classes: classes, Monsters: monsterRefs, Levels: levels,
		})
	}

	mappedSpellByID := make(map[int]mappedSpell, len(mappedSpells))
	for _, spell := range mappedSpells {
		mappedSpellByID[spell.AnkamaID] = spell
	}
	for index := range mappedClasses {
		for _, spellID := range classSpellIDs[mappedClasses[index].AnkamaID] {
			if spell, found := mappedSpellByID[spellID]; found {
				mappedClasses[index].Spells = append(mappedClasses[index].Spells, spell)
			}
		}
	}
	for index := range mappedMonsters {
		for _, link := range monsterSpellLinks[mappedMonsters[index].AnkamaID] {
			if spell, found := mappedSpellByID[link.SpellID]; found {
				mappedMonsters[index].Spells = append(mappedMonsters[index].Spells, mappedMonsterSpell{
					Spell: spell, GradeRanges: link.GradeRanges,
				})
			}
		}
	}
	return mappedClasses, mappedMonsters, mappedSpells, nil
}
