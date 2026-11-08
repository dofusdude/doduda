package main

import (
	"fmt"
	"strconv"
	"strings"

	mapping "github.com/dofusdude/dodumap"
)

type rawZoneDescription struct {
	CellIDs                     unityArray[int] `json:"cellIds"`
	DamageDecreaseStepPercent   int             `json:"damageDecreaseStepPercent"`
	ForcedDirection             int             `json:"forcedDirection"`
	IncludeCarried              int             `json:"includeCarried"`
	IsStopAtTarget              int             `json:"isStopAtTarget"`
	MaxDamageDecreaseApplyCount int             `json:"maxDamageDecreaseApplyCount"`
	OnlyAffectIfInSightLine     int             `json:"onlyAffectIfInSightLine"`
	Param1                      int             `json:"param1"`
	Param2                      int             `json:"param2"`
	Shape                       int             `json:"shape"`
}

type mappedZoneShape struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type mappedZone struct {
	Raw                         string          `json:"raw,omitempty"`
	Shape                       mappedZoneShape `json:"shape"`
	Size                        *int            `json:"size,omitempty"`
	MinSize                     *int            `json:"min_size,omitempty"`
	CellIDs                     []int           `json:"cell_ids"`
	DamageDecreaseStepPercent   int             `json:"damage_decrease_step_percent"`
	MaxDamageDecreaseApplyCount int             `json:"max_damage_decrease_apply_count"`
	ForcedDirection             int             `json:"forced_direction"`
	IncludeCarried              bool            `json:"include_carried"`
	StopAtTarget                bool            `json:"stop_at_target"`
	OnlyAffectInSightLine       bool            `json:"only_affect_in_sight_line"`
}

const globalZoneSize = 63

var zoneShapesWithMinSize = map[string]bool{
	"#": true,
	"+": true,
	"C": true,
	"Q": true,
	"R": true,
	"X": true,
	"l": true,
}

func intPtr(value int) *int {
	return &value
}

// effectiveZoneSize translates the client's generic param1/param2 fields into
// stable geometry fields. Special shapes are normalized to how they actually
// behave in combat, rather than exposing their encoded defaults.
func effectiveZoneSize(code string, size int, minSize int) (*int, *int) {
	switch code {
	case "", ";", "A", "a":
		return nil, nil
	case "P":
		return intPtr(0), nil
	case "I":
		return intPtr(globalZoneSize), intPtr(size)
	case "O":
		return intPtr(size), intPtr(size)
	default:
		if zoneShapesWithMinSize[code] {
			return intPtr(size), intPtr(minSize)
		}
		return intPtr(size), nil
	}
}

// Complete spell-area alphabet used by the current Unity client. The names
// follow the client's SpellZone behavior classes and UI localization keys.
var zoneShapeNames = map[string]string{
	"":  "empty",
	"#": "cross_without_center",
	"*": "star",
	"+": "plus",
	"-": "perpendicular_line",
	"/": "diagonal_line",
	";": "custom_cells",
	"A": "whole_map",
	"B": "boomerang",
	"C": "circle",
	"D": "checkerboard",
	"F": "fork",
	"G": "square",
	"I": "outside_circle",
	"L": "line",
	"O": "ring",
	"P": "point",
	"Q": "diagonal_cross_without_center",
	"R": "rectangle",
	"T": "perpendicular_line",
	"U": "half_circle",
	"V": "cone",
	"W": "square_without_diagonals",
	"X": "diagonal_cross",
	"Z": "outside_complex_circle",
	"a": "whole_map",
	"l": "line_from_caster",
}

func mapZoneShape(code string) mappedZoneShape {
	name, ok := zoneShapeNames[code]
	if !ok {
		panic(fmt.Sprintf("unmapped zone shape %q", code))
	}
	return mappedZoneShape{Code: code, Name: name}
}

func mapZoneDescription(raw rawZoneDescription) mappedZone {
	code := ""
	if raw.Shape > 0 {
		code = string(rune(raw.Shape))
	}
	cells := raw.CellIDs.Array
	if cells == nil {
		cells = []int{}
	}
	size, minSize := effectiveZoneSize(code, raw.Param1, raw.Param2)
	return mappedZone{
		Shape:                       mapZoneShape(code),
		Size:                        size,
		MinSize:                     minSize,
		CellIDs:                     cells,
		DamageDecreaseStepPercent:   raw.DamageDecreaseStepPercent,
		MaxDamageDecreaseApplyCount: raw.MaxDamageDecreaseApplyCount,
		ForcedDirection:             raw.ForcedDirection,
		IncludeCarried:              raw.IncludeCarried != 0,
		StopAtTarget:                raw.IsStopAtTarget != 0,
		OnlyAffectInSightLine:       raw.OnlyAffectIfInSightLine != 0,
	}
}

func parseRawZone(raw string) mappedZone {
	parts := strings.Split(raw, ",")
	code := ""
	encoded := []int{}
	if len(parts) > 0 && parts[0] != "" {
		code = string([]rune(parts[0])[0])
		firstParameter := strings.TrimPrefix(parts[0], code)
		if firstParameter != "" {
			if parameter, err := strconv.Atoi(firstParameter); err == nil {
				encoded = append(encoded, parameter)
			}
		}
	}
	for _, value := range parts[1:] {
		parameter, err := strconv.Atoi(value)
		if err == nil {
			encoded = append(encoded, parameter)
		}
	}

	// The historical "l" encoding stores its two geometric values in reverse.
	if code == "l" && len(encoded) >= 2 {
		encoded[0], encoded[1] = encoded[1], encoded[0]
	}

	size := 1
	minSize := 0
	damageDecrease := 10
	maxDamageDecrease := 4
	stopAtTarget := false
	if len(encoded) > 0 {
		size = encoded[0]
	}
	next := 1
	if zoneShapesWithMinSize[code] && len(encoded) > next {
		minSize = encoded[next]
		next++
	}
	if len(encoded) > next {
		damageDecrease = encoded[next]
		next++
	}
	if len(encoded) > next {
		maxDamageDecrease = encoded[next]
		next++
	}
	if len(encoded) > next {
		stopAtTarget = encoded[next] != 0
	}
	effectiveSize, effectiveMinSize := effectiveZoneSize(code, size, minSize)

	return mappedZone{
		Raw:                         raw,
		Shape:                       mapZoneShape(code),
		Size:                        effectiveSize,
		MinSize:                     effectiveMinSize,
		CellIDs:                     []int{},
		DamageDecreaseStepPercent:   damageDecrease,
		MaxDamageDecreaseApplyCount: maxDamageDecrease,
		StopAtTarget:                stopAtTarget,
	}
}

type rawItemTypeZone struct {
	ID      int    `json:"id"`
	RawZone string `json:"rawZone"`
}

type rawWeapon struct {
	ID             int `json:"id"`
	CastInLine     int `json:"castInLine"`
	CastInDiagonal int `json:"castInDiagonal"`
	CastTestLOS    int `json:"castTestLos"`
}

type mappedWeaponCombat struct {
	Zone           mappedZone `json:"zone"`
	CastInLine     bool       `json:"cast_in_line"`
	CastInDiagonal bool       `json:"cast_in_diagonal"`
	RequiresLOS    bool       `json:"requires_line_of_sight"`
}

type mappedItemWithZone struct {
	mapping.MappedMultilangItemUnity
	Weapon *mappedWeaponCombat `json:"weapon,omitempty"`
}

func MapItemZonesUnity(dir string, items []mapping.MappedMultilangItemUnity) ([]mappedItemWithZone, error) {
	types, err := readUnityRecords[rawItemTypeZone](dir, "item_types.json", "ItemTypeData")
	if err != nil {
		return nil, err
	}
	weapons, err := readUnityRecords[rawWeapon](dir, "items.json", "WeaponData")
	if err != nil {
		return nil, err
	}
	typeZones := make(map[int]string, len(types))
	for _, itemType := range types {
		typeZones[itemType.ID] = itemType.RawZone
	}
	weaponsByID := make(map[int]rawWeapon, len(weapons))
	for _, weapon := range weapons {
		weaponsByID[weapon.ID] = weapon
	}

	result := make([]mappedItemWithZone, 0, len(items))
	for _, item := range items {
		mapped := mappedItemWithZone{MappedMultilangItemUnity: item}
		if weapon, ok := weaponsByID[item.AnkamaId]; ok {
			rawZone, found := typeZones[item.Type.Id]
			if !found {
				return nil, fmt.Errorf("weapon %d references missing item type %d", item.AnkamaId, item.Type.Id)
			}
			mapped.Weapon = &mappedWeaponCombat{
				Zone: parseRawZone(rawZone), CastInLine: weapon.CastInLine != 0,
				CastInDiagonal: weapon.CastInDiagonal != 0, RequiresLOS: weapon.CastTestLOS != 0,
			}
		}
		result = append(result, mapped)
	}
	return result, nil
}
