package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	ID                           int                             `json:"id"`
	AdminName                    string                          `json:"adminName"`
	Order                        int                             `json:"order"`
	NameID                       flexibleInt                     `json:"nameId"`
	DescriptionID                flexibleInt                     `json:"descriptionId"`
	TypeID                       int                             `json:"typeId"`
	IconID                       int                             `json:"iconId"`
	LevelIDs                     unityArray[int]                 `json:"spellLevels"`
	BasePreviewZone              rawZoneDescription              `json:"basePreviewZoneDescr"`
	BoundScriptUsageData         unityArray[rawSpellScriptUsage] `json:"boundScriptUsageData"`
	CriticalBoundScriptUsageData unityArray[rawSpellScriptUsage] `json:"criticalHitBoundScriptUsageData"`
}

type rawSpellScriptUsage struct {
	ID                          int             `json:"id"`
	ScriptID                    int             `json:"scriptId"`
	Order                       int             `json:"order"`
	Random                      int             `json:"random"`
	RandomGroup                 int             `json:"randomGroup"`
	SequenceGroup               int             `json:"sequenceGroup"`
	ActivationMask              string          `json:"activationMask"`
	CasterMask                  string          `json:"casterMask"`
	TargetMask                  string          `json:"targetMask"`
	ActivationZone              string          `json:"activationZone"`
	TargetZone                  string          `json:"targetZone"`
	Criterion                   string          `json:"criterion"`
	TargetsAffectedConcurrently int             `json:"areTargetsAffectedConcurrently"`
	TargetTreatedAsCaster       int             `json:"isTargetTreatedAsCaster"`
	SpellLevels                 unityArray[int] `json:"spellLevels"`
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
	MaxGlobalCastPerTurn   int                        `json:"maxGlobalCastPerTurn"`
	MaxGlobalCastPerTarget int                        `json:"maxGlobalCastPerTarget"`
	MinCastInterval        int                        `json:"minCastInterval"`
	InitialCooldown        int                        `json:"initialCooldown"`
	GlobalCooldown         int                        `json:"globalCooldown"`
	MinPlayerLevel         int                        `json:"minPlayerLevel"`
	IgnorePrisonGlyph      int                        `json:"ignorePrisonGlyph"`
	SpellBreed             int                        `json:"spellBreed"`
	StatesCriterion        string                     `json:"statesCriterion"`
	Effects                unityArray[rawSpellEffect] `json:"effects"`
	CriticalEffects        unityArray[rawSpellEffect] `json:"criticalEffect"`
	PreviewZones           unityArray[rawPreviewZone] `json:"previewZones"`
}

type rawSpellEffect struct {
	EffectID              int                `json:"actionId"`
	EffectUID             int                `json:"effectUid"`
	TargetID              int                `json:"targetId"`
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
	DisplayZero           int                `json:"displayZero"`
	Modificator           int                `json:"modificator"`
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
	ID                      int                                `json:"id"`
	NameID                  flexibleInt                        `json:"nameId"`
	GfxID                   int                                `json:"gfxId"`
	RaceID                  int                                `json:"race"`
	Spells                  unityArray[int]                    `json:"spells"`
	SpellGrades             unityArray[string]                 `json:"spellGrades"`
	Grades                  unityArray[rawMonsterGrade]        `json:"grades"`
	SubareaIDs              unityArray[int]                    `json:"subareas"`
	FavoriteSubareaID       int                                `json:"favoriteSubareaId"`
	Drops                   unityArray[rawMonsterDrop]         `json:"drops"`
	GlobalDrops             unityArray[rawGlobalDrop]          `json:"globalDrops"`
	Look                    string                             `json:"look"`
	CreatureBoneID          int                                `json:"creatureBoneId"`
	IsBounty                int                                `json:"isBounty"`
	CorrespondingMiniBossID int                                `json:"correspondingMiniBossId"`
	SummonCost              int                                `json:"summonCost"`
	SpeedAdjust             int                                `json:"speedAdjust"`
	ScaleGradeRef           int                                `json:"scaleGradeRef"`
	Achievements            unityArray[int]                    `json:"achievements"`
	Souls                   unityArray[rawMonsterSoul]         `json:"souls"`
	CharacteristicRatios    unityArray[rawCharacteristicRatio] `json:"characRatios"`
}

type rawMonsterDrop struct {
	DropID                 int                                    `json:"dropId"`
	ObjectID               int                                    `json:"objectId"`
	Count                  int                                    `json:"count"`
	PercentDropForGrade1   float64                                `json:"percentDropForGrade1"`
	PercentDropForGrade2   float64                                `json:"percentDropForGrade2"`
	PercentDropForGrade3   float64                                `json:"percentDropForGrade3"`
	PercentDropForGrade4   float64                                `json:"percentDropForGrade4"`
	PercentDropForGrade5   float64                                `json:"percentDropForGrade5"`
	Criterions             string                                 `json:"criterions"`
	HiddenIfInvalid        int                                    `json:"hiddenIfInvalidCriterions"`
	DisableDropModificator int                                    `json:"disableDropModificator"`
	SpecificCoefficients   unityArray[rawSpecificDropCoefficient] `json:"specificDropCoefficient"`
}

type rawSpecificDropCoefficient struct {
	MonsterID       int     `json:"monsterId"`
	MonsterGrade    int     `json:"monsterGrade"`
	DropCoefficient float64 `json:"dropCoefficient"`
	Criterions      string  `json:"criterions"`
}

type rawGlobalDrop struct {
	ObjectID               int     `json:"objectId"`
	MinPercent             float64 `json:"minPercentDrop"`
	MaxPercent             float64 `json:"maxPercentDrop"`
	ReceiverCriterion      string  `json:"receiverCriterion"`
	AlterationID           int     `json:"alterationId"`
	DisableDropModificator int     `json:"disableDropModificator"`
}

type rawMonsterSoul struct {
	ObjectID     int    `json:"objectId"`
	Criterion    string `json:"criterion"`
	HasCriterion int    `json:"hasCriterion"`
}

type rawCharacteristicRatio struct {
	Values unityArray[float64] `json:"values"`
}

type rawMonsterCharacteristics struct {
	APRemoval               int `json:"aPRemoval"`
	Agility                 int `json:"agility"`
	ActionPoints            int `json:"actionPoints"`
	MovementPoints          int `json:"movementPoints"`
	AirResistance           int `json:"reductionAir"`
	BonusAirDamage          int `json:"airDamageBonus"`
	BonusEarthDamage        int `json:"earthDamageBonus"`
	BonusFireDamage         int `json:"fireDamageBonus"`
	BonusWaterDamage        int `json:"waterDamageBonus"`
	Chance                  int `json:"chance"`
	EarthResistance         int `json:"reductionEarth"`
	FireResistance          int `json:"reductionFire"`
	Intelligence            int `json:"intelligence"`
	LifePoints              int `json:"lifePoints"`
	NeutralResistance       int `json:"reductionNeutral"`
	Strength                int `json:"strength"`
	TackleBlock             int `json:"tackleBonus"`
	TackleEvade             int `json:"tackleEvade"`
	WaterResistance         int `json:"reductionWater"`
	Wisdom                  int `json:"wisdom"`
	APAttack                int `json:"apAttack"`
	MPAttack                int `json:"mpAttack"`
	PADodge                 int `json:"paLostDodge"`
	PMDodge                 int `json:"mpLostDodge"`
	BonusRange              int `json:"rangeBonus"`
	DamageBonus             int `json:"damageBonus"`
	PercentDamageBonus      int `json:"percentDamageBonus"`
	CriticalHitBonus        int `json:"criticalHitBonus"`
	CriticalDamageBonus     int `json:"criticalDamageBonus"`
	CriticalDamageReduction int `json:"criticalDamageReduction"`
	PushDamageBonus         int `json:"pushDamageBonus"`
	PushDamageReduction     int `json:"pushDamageReduction"`
	TrapDamageBonus         int `json:"trapDamageBonus"`
	TrapDamageBonusPercent  int `json:"trapDamageBonusPercent"`
	HealBonus               int `json:"healBonus"`
	InitiativeBonus         int `json:"initiativeBonus"`
	MaxSummon               int `json:"maxSummon"`
	Vitality                int `json:"vitality"`
}

type rawMonsterGrade struct {
	ActionPoints            int                       `json:"actionPoints"`
	Agility                 int                       `json:"agility"`
	AirResistance           int                       `json:"reductionAir"`
	BonusCharacteristics    rawMonsterCharacteristics `json:"bonusCharacteristics"`
	BonusRange              int                       `json:"rangeBonus"`
	Chance                  int                       `json:"chance"`
	DamageReflect           int                       `json:"damageReflect"`
	EarthResistance         int                       `json:"reductionEarth"`
	FireResistance          int                       `json:"reductionFire"`
	Grade                   int                       `json:"grade"`
	GradeXP                 int                       `json:"xp"`
	Intelligence            int                       `json:"intelligence"`
	Level                   int                       `json:"level"`
	LifePoints              int                       `json:"lifePoints"`
	MovementPoints          int                       `json:"movementPoints"`
	NeutralResistance       int                       `json:"reductionNeutral"`
	PADodge                 int                       `json:"paLostDodge"`
	PMDodge                 int                       `json:"mpLostDodge"`
	StartingSpellID         int                       `json:"startingSpellId"`
	Strength                int                       `json:"strength"`
	Vitality                int                       `json:"vitality"`
	WaterResistance         int                       `json:"reductionWater"`
	Wisdom                  int                       `json:"wisdom"`
	APAttack                int                       `json:"apAttack"`
	MPAttack                int                       `json:"mpAttack"`
	DamageBonus             int                       `json:"damageBonus"`
	PercentDamageBonus      int                       `json:"percentDamageBonus"`
	CriticalHitBonus        int                       `json:"criticalHitBonus"`
	CriticalDamageBonus     int                       `json:"criticalDamageBonus"`
	CriticalDamageReduction int                       `json:"criticalDamageReduction"`
	PushDamageBonus         int                       `json:"pushDamageBonus"`
	PushDamageReduction     int                       `json:"pushDamageReduction"`
	TrapDamageBonus         int                       `json:"trapDamageBonus"`
	TrapDamageBonusPercent  int                       `json:"trapDamageBonusPercent"`
	HealBonus               int                       `json:"healBonus"`
	InitiativeBonus         int                       `json:"initiativeBonus"`
	TackleBonus             int                       `json:"tackleBonus"`
	TackleEvade             int                       `json:"tackleEvade"`
	MaxSummon               int                       `json:"maxSummon"`
	AirDamageBonus          int                       `json:"airDamageBonus"`
	EarthDamageBonus        int                       `json:"earthDamageBonus"`
	FireDamageBonus         int                       `json:"fireDamageBonus"`
	WaterDamageBonus        int                       `json:"waterDamageBonus"`
	NeutralDamageBonus      int                       `json:"neutralDamageBonus"`
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

type rawAchievement struct {
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

type mappedSpellLevel struct {
	AnkamaID               int                              `json:"ankama_id"`
	Grade                  int                              `json:"grade"`
	APCost                 int                              `json:"ap_cost"`
	MinRange               int                              `json:"min_range"`
	MaxRange               int                              `json:"max_range"`
	CriticalHitProbability int                              `json:"critical_hit_probability"`
	MaxStack               int                              `json:"max_stack"`
	MaxCastPerTurn         int                              `json:"max_cast_per_turn"`
	MaxCastPerTarget       int                              `json:"max_cast_per_target"`
	MaxGlobalCastPerTurn   int                              `json:"max_global_cast_per_turn"`
	MaxGlobalCastPerTarget int                              `json:"max_global_cast_per_target"`
	MinCastInterval        int                              `json:"min_cast_interval"`
	InitialCooldown        int                              `json:"initial_cooldown"`
	GlobalCooldown         int                              `json:"global_cooldown"`
	MinPlayerLevel         int                              `json:"min_player_level"`
	IgnorePrisonGlyph      bool                             `json:"ignore_prison_glyph"`
	SpellBreed             int                              `json:"spell_breed"`
	StatesCriterion        *mapping.ConditionTreeNodeMapped `json:"states_criterion"`
	PreviewZones           []mappedPreviewZone              `json:"preview_zones"`
	Effects                []mappedSpellEffect              `json:"effects"`
	CriticalEffects        []mappedSpellEffect              `json:"critical_effects"`
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
	EffectUID         int              `json:"effect_uid"`
	TargetID          int              `json:"target_id"`
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
	DisplayZero       bool             `json:"display_zero"`
	Modifier          int              `json:"modifier"`
}

type mappedSpell struct {
	AnkamaID        int                      `json:"ankama_id"`
	AdminName       string                   `json:"admin_name"`
	Order           int                      `json:"order"`
	Name            map[string]string        `json:"name"`
	Description     map[string]string        `json:"description"`
	Type            mappedSpellType          `json:"type"`
	IconID          int                      `json:"icon_id"`
	BasePreviewZone mappedZone               `json:"base_preview_zone"`
	Classes         []mappedNamedReference   `json:"classes"`
	Monsters        []mappedNamedReference   `json:"monsters"`
	Levels          []mappedSpellLevel       `json:"levels"`
	Scripts         []mappedSpellScriptUsage `json:"scripts"`
	CriticalScripts []mappedSpellScriptUsage `json:"critical_scripts"`
}

type mappedSpellScriptUsage struct {
	ConditionExpression         string                           `json:"condition_expression"`
	AnkamaID                    int                              `json:"ankama_id"`
	ScriptID                    int                              `json:"script_id"`
	Order                       int                              `json:"order"`
	Random                      int                              `json:"random"`
	RandomGroup                 int                              `json:"random_group"`
	SequenceGroup               int                              `json:"sequence_group"`
	ActivationMask              mappedTargetMask                 `json:"activation_mask"`
	CasterMask                  mappedTargetMask                 `json:"caster_mask"`
	TargetMask                  mappedTargetMask                 `json:"target_mask"`
	ActivationZone              mappedZone                       `json:"activation_zone"`
	TargetZone                  mappedZone                       `json:"target_zone"`
	Condition                   *mapping.ConditionTreeNodeMapped `json:"condition"`
	TargetsAffectedConcurrently bool                             `json:"targets_affected_concurrently"`
	TargetTreatedAsCaster       bool                             `json:"target_treated_as_caster"`
	SpellLevelIDs               []int                            `json:"spell_level_ids"`
}

type mappedMonsterRace struct {
	AnkamaID  int                  `json:"ankama_id"`
	Name      map[string]string    `json:"name"`
	SuperRace mappedNamedReference `json:"super_race"`
}

type mappedStartingSpell struct {
	SpellLevelID int               `json:"spell_level_id"`
	SpellID      int               `json:"spell_id,omitempty"`
	Grade        int               `json:"grade,omitempty"`
	Name         map[string]string `json:"name"`
}

type mappedMonsterCharacteristics struct {
	APRemoval               int `json:"ap_removal"`
	ActionPoints            int `json:"action_points"`
	MovementPoints          int `json:"movement_points"`
	Agility                 int `json:"agility"`
	AirResistance           int `json:"air_resistance"`
	BonusAirDamage          int `json:"bonus_air_damage"`
	BonusEarthDamage        int `json:"bonus_earth_damage"`
	BonusFireDamage         int `json:"bonus_fire_damage"`
	BonusWaterDamage        int `json:"bonus_water_damage"`
	Chance                  int `json:"chance"`
	EarthResistance         int `json:"earth_resistance"`
	FireResistance          int `json:"fire_resistance"`
	Intelligence            int `json:"intelligence"`
	LifePoints              int `json:"life_points"`
	NeutralResistance       int `json:"neutral_resistance"`
	Strength                int `json:"strength"`
	TackleBlock             int `json:"tackle_block"`
	TackleEvade             int `json:"tackle_evade"`
	WaterResistance         int `json:"water_resistance"`
	Wisdom                  int `json:"wisdom"`
	APAttack                int `json:"ap_attack"`
	MPAttack                int `json:"mp_attack"`
	PADodge                 int `json:"pa_dodge"`
	PMDodge                 int `json:"pm_dodge"`
	BonusRange              int `json:"bonus_range"`
	DamageBonus             int `json:"damage_bonus"`
	PercentDamageBonus      int `json:"percent_damage_bonus"`
	CriticalHitBonus        int `json:"critical_hit_bonus"`
	CriticalDamageBonus     int `json:"critical_damage_bonus"`
	CriticalDamageReduction int `json:"critical_damage_reduction"`
	PushDamageBonus         int `json:"push_damage_bonus"`
	PushDamageReduction     int `json:"push_damage_reduction"`
	TrapDamageBonus         int `json:"trap_damage_bonus"`
	TrapDamageBonusPercent  int `json:"trap_damage_bonus_percent"`
	HealBonus               int `json:"heal_bonus"`
	InitiativeBonus         int `json:"initiative_bonus"`
	MaxSummon               int `json:"max_summon"`
	Vitality                int `json:"vitality"`
}

type mappedMonsterGrade struct {
	ActionPoints            int                          `json:"action_points"`
	Agility                 int                          `json:"agility"`
	AirResistance           int                          `json:"air_resistance"`
	BonusCharacteristics    mappedMonsterCharacteristics `json:"bonus_characteristics"`
	BonusRange              int                          `json:"bonus_range"`
	Chance                  int                          `json:"chance"`
	DamageReflect           int                          `json:"damage_reflect"`
	EarthResistance         int                          `json:"earth_resistance"`
	FireResistance          int                          `json:"fire_resistance"`
	Grade                   int                          `json:"grade"`
	GradeXP                 int                          `json:"grade_xp"`
	Intelligence            int                          `json:"intelligence"`
	Level                   int                          `json:"level"`
	LifePoints              int                          `json:"life_points"`
	MovementPoints          int                          `json:"movement_points"`
	NeutralResistance       int                          `json:"neutral_resistance"`
	PADodge                 int                          `json:"pa_dodge"`
	PMDodge                 int                          `json:"pm_dodge"`
	StartingSpell           *mappedStartingSpell         `json:"starting_spell"`
	Strength                int                          `json:"strength"`
	Vitality                int                          `json:"vitality"`
	WaterResistance         int                          `json:"water_resistance"`
	Wisdom                  int                          `json:"wisdom"`
	Spells                  []mappedSpell                `json:"spells"`
	APAttack                int                          `json:"ap_attack"`
	MPAttack                int                          `json:"mp_attack"`
	DamageBonus             int                          `json:"damage_bonus"`
	PercentDamageBonus      int                          `json:"percent_damage_bonus"`
	CriticalHitBonus        int                          `json:"critical_hit_bonus"`
	CriticalDamageBonus     int                          `json:"critical_damage_bonus"`
	CriticalDamageReduction int                          `json:"critical_damage_reduction"`
	PushDamageBonus         int                          `json:"push_damage_bonus"`
	PushDamageReduction     int                          `json:"push_damage_reduction"`
	TrapDamageBonus         int                          `json:"trap_damage_bonus"`
	TrapDamageBonusPercent  int                          `json:"trap_damage_bonus_percent"`
	HealBonus               int                          `json:"heal_bonus"`
	InitiativeBonus         int                          `json:"initiative_bonus"`
	TackleBonus             int                          `json:"tackle_bonus"`
	TackleEvade             int                          `json:"tackle_evade"`
	MaxSummon               int                          `json:"max_summon"`
	DamageBonuses           mappedElementalDamageBonuses `json:"damage_bonuses"`
}

type mappedElementalDamageBonuses struct {
	Air     int `json:"air"`
	Earth   int `json:"earth"`
	Fire    int `json:"fire"`
	Water   int `json:"water"`
	Neutral int `json:"neutral"`
}

type mappedMonster struct {
	AnkamaID              int                    `json:"ankama_id"`
	Name                  map[string]string      `json:"name"`
	GfxID                 int                    `json:"gfx_id"`
	Race                  mappedMonsterRace      `json:"race"`
	Grades                []mappedMonsterGrade   `json:"grades"`
	SpawnLocations        []mappedSpawnLocation  `json:"spawn_locations"`
	Drops                 []mappedMonsterDrop    `json:"drops"`
	GlobalDrops           []mappedGlobalDrop     `json:"global_drops"`
	Appearance            *mappedClassAppearance `json:"appearance"`
	CreatureBoneID        int                    `json:"creature_bone_id"`
	IsBounty              bool                   `json:"is_bounty"`
	CorrespondingMiniBoss *mappedNamedReference  `json:"corresponding_mini_boss"`
	SummonCost            int                    `json:"summon_cost"`
	SpeedAdjust           int                    `json:"speed_adjust"`
	ScaleGradeReference   int                    `json:"scale_grade_reference"`
	Achievements          []mappedNamedReference `json:"achievements"`
	Souls                 []mappedMonsterSoul    `json:"souls"`
	CharacteristicRatios  [][]float64            `json:"characteristic_ratios"`
}

type mappedMonsterDrop struct {
	ConditionExpression       string                           `json:"condition_expression"`
	DropID                    int                              `json:"drop_id"`
	Item                      mappedNamedReference             `json:"item"`
	RatesPercent              []mappedGradeRate                `json:"rates_percent"`
	HasConditions             bool                             `json:"has_conditions"`
	Condition                 *mapping.ConditionTreeNodeMapped `json:"condition"`
	Count                     int                              `json:"count"`
	HiddenIfInvalidConditions bool                             `json:"hidden_if_invalid_conditions"`
	DisableDropModifier       bool                             `json:"disable_drop_modifier"`
	SpecificCoefficients      []mappedSpecificDropCoefficient  `json:"specific_coefficients"`
}

type mappedSpecificDropCoefficient struct {
	ConditionExpression string                           `json:"condition_expression"`
	MonsterGrade        int                              `json:"monster_grade"`
	Coefficient         float64                          `json:"coefficient"`
	Condition           *mapping.ConditionTreeNodeMapped `json:"condition"`
}

type mappedGlobalDrop struct {
	HasConditions       bool                             `json:"has_conditions"`
	ConditionExpression string                           `json:"condition_expression"`
	Item                mappedNamedReference             `json:"item"`
	MinPercent          float64                          `json:"min_percent"`
	MaxPercent          float64                          `json:"max_percent"`
	Condition           *mapping.ConditionTreeNodeMapped `json:"condition"`
	AlterationID        int                              `json:"alteration_id"`
	DisableDropModifier bool                             `json:"disable_drop_modifier"`
}

type mappedMonsterSoul struct {
	HasConditions       bool                             `json:"has_conditions"`
	ConditionExpression string                           `json:"condition_expression"`
	Item                mappedNamedReference             `json:"item"`
	Condition           *mapping.ConditionTreeNodeMapped `json:"condition"`
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
				MappedMultilangEffect: *effect, EffectUID: raw.EffectUID, TargetID: raw.TargetID,
				Zone:       mapZoneDescription(raw.Zone),
				TargetMask: mapTargetMask(raw.TargetMask, states, monsters, breeds, *languages),
				Triggers:   mapTriggers(raw.Triggers, states, monsters, breeds, spells, *languages), Duration: raw.Duration,
				Delay: raw.Delay, Random: raw.Random, Group: raw.Group, Order: raw.Order,
				Dispellable: raw.Dispellable, TriggeredDuration: raw.EffectTriggerDuration,
				DisplayZero: raw.DisplayZero != 0, Modifier: raw.Modificator,
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

func mapSpellScripts(raw []rawSpellScriptUsage, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, languages map[string]mapping.LangDictUnity, gameData *mapping.JSONGameDataUnity) []mappedSpellScriptUsage {
	result := make([]mappedSpellScriptUsage, 0, len(raw))
	for _, usage := range raw {
		result = append(result, mappedSpellScriptUsage{
			AnkamaID: usage.ID, ScriptID: usage.ScriptID, Order: usage.Order,
			Random: usage.Random, RandomGroup: usage.RandomGroup, SequenceGroup: usage.SequenceGroup,
			ActivationMask: mapTargetMask(usage.ActivationMask, states, monsters, breeds, languages),
			CasterMask:     mapTargetMask(usage.CasterMask, states, monsters, breeds, languages),
			TargetMask:     mapTargetMask(usage.TargetMask, states, monsters, breeds, languages),
			ActivationZone: parseRawZone(usage.ActivationZone), TargetZone: parseRawZone(usage.TargetZone),
			Condition:                   mapping.ParseConditionUnity(usage.Criterion, &languages, gameData),
			ConditionExpression:         usage.Criterion,
			TargetsAffectedConcurrently: usage.TargetsAffectedConcurrently != 0,
			TargetTreatedAsCaster:       usage.TargetTreatedAsCaster != 0,
			SpellLevelIDs:               usage.SpellLevels.Array,
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

func mapSpellLevel(level rawSpellLevel, states map[int]rawSpellState, monsters map[int]rawMonster, breeds map[int]rawBreed, spells map[int]rawSpell, gameData *mapping.JSONGameDataUnity, languages *map[string]mapping.LangDictUnity) mappedSpellLevel {
	return mappedSpellLevel{
		AnkamaID: level.ID, Grade: level.Grade, APCost: level.APCost,
		MinRange: level.MinRange, MaxRange: level.Range,
		CriticalHitProbability: level.CriticalHitProbability, MaxStack: level.MaxStack,
		MaxCastPerTurn: level.MaxCastPerTurn, MaxCastPerTarget: level.MaxCastPerTarget,
		MaxGlobalCastPerTurn: level.MaxGlobalCastPerTurn, MaxGlobalCastPerTarget: level.MaxGlobalCastPerTarget,
		MinCastInterval: level.MinCastInterval, InitialCooldown: level.InitialCooldown,
		GlobalCooldown: level.GlobalCooldown, MinPlayerLevel: level.MinPlayerLevel,
		IgnorePrisonGlyph: level.IgnorePrisonGlyph != 0, SpellBreed: level.SpellBreed,
		StatesCriterion: mapping.ParseConditionUnity(level.StatesCriterion, languages, gameData),
		PreviewZones:    mapPreviewZones(level.PreviewZones.Array, states, monsters, breeds, *languages),
		Effects:         mapSpellEffects(gameData, level.Effects.Array, states, monsters, breeds, spells, languages),
		CriticalEffects: mapSpellEffects(gameData, level.CriticalEffects.Array, states, monsters, breeds, spells, languages),
	}
}

type monsterSpellAssignment struct {
	MonsterGrade int
	MonsterLevel int
	SpellGrade   int
}

func parseMonsterSpellGrades(encoded string) ([]monsterSpellAssignment, error) {
	if encoded == "" || encoded == "null" {
		return []monsterSpellAssignment{}, nil
	}
	parts := strings.Split(encoded, ";")
	assignments := make([]monsterSpellAssignment, 0, len(parts))
	for index, part := range parts {
		bounds := strings.Split(part, ",")
		if len(bounds) != 2 {
			return nil, fmt.Errorf("invalid monster spell grade assignment %q", part)
		}
		// Ankama pads the list to five entries even when a monster has fewer
		// grades. A null monster level marks one of those absent grades.
		if bounds[1] == "null" {
			continue
		}
		spellGrade, spellErr := strconv.Atoi(bounds[0])
		monsterLevel, levelErr := strconv.Atoi(bounds[1])
		if spellErr != nil || levelErr != nil {
			return nil, fmt.Errorf("invalid monster spell grade assignment %q", part)
		}
		assignments = append(assignments, monsterSpellAssignment{
			MonsterGrade: index + 1, MonsterLevel: monsterLevel, SpellGrade: spellGrade,
		})
	}
	return assignments, nil
}

func spellAtResolvedGrade(spell mappedSpell, requestedGrade int) (mappedSpell, bool) {
	if requestedGrade <= 0 || len(spell.Levels) == 0 {
		return mappedSpell{}, false
	}
	selected := spell.Levels[0]
	for _, level := range spell.Levels {
		if level.Grade == requestedGrade {
			selected = level
			break
		}
		if level.Grade <= requestedGrade && level.Grade > selected.Grade {
			selected = level
		}
		if requestedGrade < selected.Grade && level.Grade < selected.Grade {
			selected = level
		}
	}
	resolved := spell
	resolved.Levels = []mappedSpellLevel{selected}
	return resolved, true
}

func mapCharacteristics(value rawMonsterCharacteristics) mappedMonsterCharacteristics {
	return mappedMonsterCharacteristics{
		APRemoval: value.APRemoval, ActionPoints: value.ActionPoints, MovementPoints: value.MovementPoints,
		Agility: value.Agility, AirResistance: value.AirResistance,
		BonusAirDamage: value.BonusAirDamage, BonusEarthDamage: value.BonusEarthDamage,
		BonusFireDamage: value.BonusFireDamage, BonusWaterDamage: value.BonusWaterDamage,
		Chance: value.Chance, EarthResistance: value.EarthResistance, FireResistance: value.FireResistance,
		Intelligence: value.Intelligence, LifePoints: value.LifePoints, NeutralResistance: value.NeutralResistance,
		Strength: value.Strength, TackleBlock: value.TackleBlock, TackleEvade: value.TackleEvade,
		WaterResistance: value.WaterResistance, Wisdom: value.Wisdom,
		APAttack: value.APAttack, MPAttack: value.MPAttack, PADodge: value.PADodge, PMDodge: value.PMDodge,
		BonusRange: value.BonusRange, DamageBonus: value.DamageBonus, PercentDamageBonus: value.PercentDamageBonus,
		CriticalHitBonus: value.CriticalHitBonus, CriticalDamageBonus: value.CriticalDamageBonus,
		CriticalDamageReduction: value.CriticalDamageReduction, PushDamageBonus: value.PushDamageBonus,
		PushDamageReduction: value.PushDamageReduction, TrapDamageBonus: value.TrapDamageBonus,
		TrapDamageBonusPercent: value.TrapDamageBonusPercent, HealBonus: value.HealBonus,
		InitiativeBonus: value.InitiativeBonus, MaxSummon: value.MaxSummon, Vitality: value.Vitality,
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
	// Some monsters encode several conditional looks separated by `},{`.
	// The first entry is the stable base appearance exposed by this mapping.
	baseLook := raw
	if separator := strings.Index(baseLook, "},{"); separator >= 0 {
		baseLook = baseLook[:separator+1]
	}
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(baseLook, "{"), "}"), "|")
	if len(parts) == 0 || parts[0] == "" {
		return mappedClassAppearance{}, fmt.Errorf("invalid entity look %q", raw)
	}
	bonesValue := strings.SplitN(parts[0], "$", 2)[0]
	bonesID, err := strconv.Atoi(bonesValue)
	if err != nil {
		return mappedClassAppearance{}, fmt.Errorf("parse entity look bones %q: %w", raw, err)
	}
	if strings.Contains(parts[0], "$") {
		if colors == nil {
			colors = []int{}
		}
		return mappedClassAppearance{
			BonesID: bonesID, SkinIDs: []int{}, Scale: mappedScale{X: 100, Y: 100}, DefaultColors: colors,
		}, nil
	}
	skins := []int{}
	if len(parts) > 1 {
		skins, err = parseIntegerList(parts[1])
		if err != nil {
			return mappedClassAppearance{}, fmt.Errorf("parse entity look skins %q: %w", raw, err)
		}
	}
	scales := []int{}
	if len(parts) > 3 {
		scaleValue := strings.SplitN(parts[3], "$", 2)[0]
		scales, err = parseIntegerList(scaleValue)
		if err != nil {
			return mappedClassAppearance{}, fmt.Errorf("parse entity look scale %q: %w", raw, err)
		}
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
		APAttack: grade.APAttack, MPAttack: grade.MPAttack, DamageBonus: grade.DamageBonus,
		PercentDamageBonus: grade.PercentDamageBonus, CriticalHitBonus: grade.CriticalHitBonus,
		CriticalDamageBonus: grade.CriticalDamageBonus, CriticalDamageReduction: grade.CriticalDamageReduction,
		PushDamageBonus: grade.PushDamageBonus, PushDamageReduction: grade.PushDamageReduction,
		TrapDamageBonus: grade.TrapDamageBonus, TrapDamageBonusPercent: grade.TrapDamageBonusPercent,
		HealBonus: grade.HealBonus, InitiativeBonus: grade.InitiativeBonus, TackleBonus: grade.TackleBonus,
		TackleEvade: grade.TackleEvade, MaxSummon: grade.MaxSummon,
		DamageBonuses: mappedElementalDamageBonuses{Air: grade.AirDamageBonus, Earth: grade.EarthDamageBonus,
			Fire: grade.FireDamageBonus, Water: grade.WaterDamageBonus, Neutral: grade.NeutralDamageBonus},
	}
}

func MapSpellEntitiesUnity(dir string, gameData *mapping.JSONGameDataUnity, languages map[string]mapping.LangDictUnity, spawnByMonster map[int][]mappedSpawnLocation) ([]mappedClass, []mappedMonster, []mappedSpell, error) {
	err := loadConditionReferences(dir, gameData, languages)
	if err != nil {
		return nil, nil, nil, err
	}
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
	achievements, err := readUnityRecords[rawAchievement](dir, "achievements.json", "AchievementData")
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
	achievementByID := make(map[int]rawAchievement, len(achievements))
	for _, value := range achievements {
		achievementByID[value.ID] = value
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
		Assignments []monsterSpellAssignment
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
				assignments, err := parseMonsterSpellGrades(encoded)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("monster %d spell %d: %w", monster.ID, spellID, err)
				}
				monsterSpellLinks[monster.ID] = append(monsterSpellLinks[monster.ID], pendingMonsterSpell{
					SpellID: spellID, Assignments: assignments,
				})
				monsterRefsBySpell[spellID] = append(monsterRefsBySpell[spellID], monsterRef)
			}
		}
		mappedGrades := make([]mappedMonsterGrade, 0, len(monster.Grades.Array))
		for _, grade := range monster.Grades.Array {
			mappedGrade := mapMonsterGrade(grade, levelByID, spellByID, languages)
			mappedGrade.Spells = []mappedSpell{}
			mappedGrades = append(mappedGrades, mappedGrade)
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
			coefficients := make([]mappedSpecificDropCoefficient, 0, len(drop.SpecificCoefficients.Array))
			for _, coefficient := range drop.SpecificCoefficients.Array {
				if coefficient.MonsterID != 0 && coefficient.MonsterID != monster.ID {
					continue
				}
				coefficients = append(coefficients, mappedSpecificDropCoefficient{
					MonsterGrade: coefficient.MonsterGrade, Coefficient: coefficient.DropCoefficient,
					Condition:           mapping.ParseConditionUnity(coefficient.Criterions, &languages, gameData),
					ConditionExpression: coefficient.Criterions,
				})
			}
			drops = append(drops, mappedMonsterDrop{
				DropID: drop.DropID,
				Item:   namedReference(drop.ObjectID, flexibleInt(item.NameId), true, languages),
				RatesPercent: []mappedGradeRate{
					{Grade: 1, Percent: drop.PercentDropForGrade1},
					{Grade: 2, Percent: drop.PercentDropForGrade2},
					{Grade: 3, Percent: drop.PercentDropForGrade3},
					{Grade: 4, Percent: drop.PercentDropForGrade4},
					{Grade: 5, Percent: drop.PercentDropForGrade5},
				},
				HasConditions:       drop.Criterions != "",
				Condition:           mapping.ParseConditionUnity(drop.Criterions, &languages, gameData),
				ConditionExpression: drop.Criterions,
				Count:               drop.Count, HiddenIfInvalidConditions: drop.HiddenIfInvalid != 0,
				DisableDropModifier: drop.DisableDropModificator != 0, SpecificCoefficients: coefficients,
			})
		}
		globalDrops := make([]mappedGlobalDrop, 0, len(monster.GlobalDrops.Array))
		for _, drop := range monster.GlobalDrops.Array {
			item, found := gameData.Items[drop.ObjectID]
			if !found {
				continue
			}
			globalDrops = append(globalDrops, mappedGlobalDrop{
				Item:       namedReference(drop.ObjectID, flexibleInt(item.NameId), true, languages),
				MinPercent: drop.MinPercent, MaxPercent: drop.MaxPercent,
				Condition:     mapping.ParseConditionUnity(drop.ReceiverCriterion, &languages, gameData),
				HasConditions: drop.ReceiverCriterion != "", ConditionExpression: drop.ReceiverCriterion,
				AlterationID: drop.AlterationID, DisableDropModifier: drop.DisableDropModificator != 0,
			})
		}
		var appearance *mappedClassAppearance
		if monster.Look != "" {
			mappedAppearance, err := mapClassAppearance(monster.Look, nil)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("map appearance for monster %d: %w", monster.ID, err)
			}
			appearance = &mappedAppearance
		}
		var correspondingMiniBoss *mappedNamedReference
		if rawMiniBoss, found := monsterByID[monster.CorrespondingMiniBossID]; found {
			ref := namedReference(rawMiniBoss.ID, rawMiniBoss.NameID, true, languages)
			correspondingMiniBoss = &ref
		}
		souls := make([]mappedMonsterSoul, 0, len(monster.Souls.Array))
		for _, soul := range monster.Souls.Array {
			item, found := gameData.Items[soul.ObjectID]
			if !found {
				continue
			}
			souls = append(souls, mappedMonsterSoul{
				Item:          namedReference(soul.ObjectID, flexibleInt(item.NameId), true, languages),
				Condition:     mapping.ParseConditionUnity(soul.Criterion, &languages, gameData),
				HasConditions: soul.Criterion != "", ConditionExpression: soul.Criterion,
			})
		}
		ratios := make([][]float64, 0, len(monster.CharacteristicRatios.Array))
		for _, ratio := range monster.CharacteristicRatios.Array {
			ratios = append(ratios, ratio.Values.Array)
		}
		achievementRefs := make([]mappedNamedReference, 0, len(monster.Achievements.Array))
		for _, achievementID := range monster.Achievements.Array {
			if achievement, found := achievementByID[achievementID]; found {
				achievementRefs = append(achievementRefs, namedReference(achievement.ID, achievement.NameID, true, languages))
			}
		}
		mappedMonsters = append(mappedMonsters, mappedMonster{
			AnkamaID: monster.ID, Name: monsterRef.Name, GfxID: monster.GfxID, Race: race,
			Grades: mappedGrades, SpawnLocations: spawnLocations, Drops: drops, GlobalDrops: globalDrops,
			Appearance: appearance, CreatureBoneID: monster.CreatureBoneID, IsBounty: monster.IsBounty != 0,
			CorrespondingMiniBoss: correspondingMiniBoss, SummonCost: monster.SummonCost,
			SpeedAdjust: monster.SpeedAdjust, ScaleGradeReference: monster.ScaleGradeRef,
			Achievements: achievementRefs, Souls: souls, CharacteristicRatios: ratios,
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
			AnkamaID: spell.ID, AdminName: spell.AdminName, Order: spell.Order, Name: localizedText(spell.NameID, languages),
			Description: localizedText(spell.DescriptionID, languages), Type: mappedType, IconID: spell.IconID,
			BasePreviewZone: mapZoneDescription(spell.BasePreviewZone), Classes: classes, Monsters: monsterRefs, Levels: levels,
			Scripts:         mapSpellScripts(spell.BoundScriptUsageData.Array, stateByID, monsterByID, breedByID, languages, gameData),
			CriticalScripts: mapSpellScripts(spell.CriticalBoundScriptUsageData.Array, stateByID, monsterByID, breedByID, languages, gameData),
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
			spell, found := mappedSpellByID[link.SpellID]
			if !found {
				continue
			}
			for _, assignment := range link.Assignments {
				if assignment.SpellGrade == 0 {
					continue
				}
				gradeIndex := -1
				for candidate := range mappedMonsters[index].Grades {
					if mappedMonsters[index].Grades[candidate].Grade == assignment.MonsterGrade {
						gradeIndex = candidate
						break
					}
				}
				if gradeIndex < 0 {
					return nil, nil, nil, fmt.Errorf("monster %d spell %d refers to missing monster grade %d", mappedMonsters[index].AnkamaID, link.SpellID, assignment.MonsterGrade)
				}
				monsterGrade := &mappedMonsters[index].Grades[gradeIndex]
				if monsterGrade.Level != assignment.MonsterLevel {
					return nil, nil, nil, fmt.Errorf("monster %d grade %d level is %d, spell assignment says %d", mappedMonsters[index].AnkamaID, assignment.MonsterGrade, monsterGrade.Level, assignment.MonsterLevel)
				}
				gradeSpell, available := spellAtResolvedGrade(spell, assignment.SpellGrade)
				if !available {
					continue
				}
				monsterGrade.Spells = append(monsterGrade.Spells, gradeSpell)
			}
		}
	}
	return mappedClasses, mappedMonsters, mappedSpells, nil
}
