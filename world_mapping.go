package main

import (
	"fmt"

	mapping "github.com/dofusdude/dodumap"
)

type rawSuperArea struct {
	ID         int         `json:"id"`
	NameID     flexibleInt `json:"nameId"`
	WorldMapID int         `json:"worldmapId"`
}

type rawArea struct {
	ID          int             `json:"id"`
	NameID      flexibleInt     `json:"nameId"`
	SuperAreaID int             `json:"superAreaId"`
	SubareaIDs  unityArray[int] `json:"subareaIds"`
}

type rawSubarea struct {
	ID         int             `json:"id"`
	NameID     flexibleInt     `json:"nameId"`
	AreaID     int             `json:"areaId"`
	Level      int             `json:"level"`
	MapIDs     unityArray[int] `json:"mapIds"`
	MonsterIDs unityArray[int] `json:"monsters"`
	DungeonID  int             `json:"dungeonId"`
}

type rawMapInformation struct {
	ID        int `json:"id"`
	PosX      int `json:"posX"`
	PosY      int `json:"posY"`
	SubareaID int `json:"subAreaId"`
	WorldMap  int `json:"worldMap"`
}

type mappedMapReference struct {
	AnkamaID  int `json:"ankama_id"`
	PositionX int `json:"position_x"`
	PositionY int `json:"position_y"`
	WorldMap  int `json:"world_map_id"`
}

type mappedLocationReference struct {
	AnkamaID int               `json:"ankama_id"`
	Name     map[string]string `json:"name"`
}

type mappedSpawnLocation struct {
	Subarea     mappedLocationReference `json:"subarea"`
	Area        mappedLocationReference `json:"area"`
	SuperArea   mappedLocationReference `json:"super_area"`
	SubareaMaps []mappedMapReference    `json:"subarea_maps"`
	Preferred   bool                    `json:"preferred"`
}

type mappedSubareaReference struct {
	AnkamaID int               `json:"ankama_id"`
	Name     map[string]string `json:"name"`
}

type mappedArea struct {
	AnkamaID  int                      `json:"ankama_id"`
	Name      map[string]string        `json:"name"`
	SuperArea mappedLocationReference  `json:"super_area"`
	Subareas  []mappedSubareaReference `json:"subareas"`
}

type mappedSuperArea struct {
	AnkamaID   int                       `json:"ankama_id"`
	Name       map[string]string         `json:"name"`
	WorldMapID int                       `json:"world_map_id"`
	Areas      []mappedLocationReference `json:"areas"`
}

type mappedSubarea struct {
	AnkamaID  int                     `json:"ankama_id"`
	Name      map[string]string       `json:"name"`
	Level     int                     `json:"level"`
	DungeonID int                     `json:"dungeon_id,omitempty"`
	Area      mappedLocationReference `json:"area"`
	SuperArea mappedLocationReference `json:"super_area"`
	Maps      []mappedMapReference    `json:"maps"`
	Monsters  []mappedNamedReference  `json:"monsters"`
}

func locationReference(id int, nameID flexibleInt, found bool, languages map[string]mapping.LangDictUnity) mappedLocationReference {
	name := map[string]string{}
	if found {
		name = localizedText(nameID, languages)
	}
	return mappedLocationReference{AnkamaID: id, Name: name}
}

func MapWorldAreasUnity(dir string, languages map[string]mapping.LangDictUnity) ([]mappedSuperArea, []mappedArea, []mappedSubarea, map[int][]mappedSpawnLocation, error) {
	superAreas, err := readUnityRecords[rawSuperArea](dir, "superareas.json", "SuperAreaData")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	areas, err := readUnityRecords[rawArea](dir, "areas.json", "AreaData")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	subareas, err := readUnityRecords[rawSubarea](dir, "subareas.json", "SubAreaData")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	maps, err := readUnityRecords[rawMapInformation](dir, "maps_information.json", "MapInformationData")
	if err != nil {
		return nil, nil, nil, nil, err
	}
	monsters, err := readUnityRecords[rawMonster](dir, "monsters.json", "MonsterData")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	superByID := make(map[int]rawSuperArea, len(superAreas))
	for _, value := range superAreas {
		superByID[value.ID] = value
	}
	areaByID := make(map[int]rawArea, len(areas))
	for _, value := range areas {
		areaByID[value.ID] = value
	}
	subareaByID := make(map[int]rawSubarea, len(subareas))
	for _, value := range subareas {
		subareaByID[value.ID] = value
	}
	mapByID := make(map[int]rawMapInformation, len(maps))
	for _, value := range maps {
		mapByID[value.ID] = value
	}
	monsterByID := make(map[int]rawMonster, len(monsters))
	for _, value := range monsters {
		monsterByID[value.ID] = value
	}
	areasBySuper := make(map[int][]mappedLocationReference)
	for _, area := range areas {
		areasBySuper[area.SuperAreaID] = append(areasBySuper[area.SuperAreaID], locationReference(area.ID, area.NameID, true, languages))
	}
	mappedSuperAreas := make([]mappedSuperArea, 0, len(superAreas))
	for _, super := range superAreas {
		areaRefs := areasBySuper[super.ID]
		if areaRefs == nil {
			areaRefs = []mappedLocationReference{}
		}
		mappedSuperAreas = append(mappedSuperAreas, mappedSuperArea{
			AnkamaID: super.ID, Name: localizedText(super.NameID, languages), WorldMapID: super.WorldMapID, Areas: areaRefs,
		})
	}

	mappedAreas := make([]mappedArea, 0, len(areas))
	for _, area := range areas {
		super, found := superByID[area.SuperAreaID]
		if !found {
			return nil, nil, nil, nil, fmt.Errorf("area %d references missing super area %d", area.ID, area.SuperAreaID)
		}
		subareaRefs := make([]mappedSubareaReference, 0, len(area.SubareaIDs.Array))
		for _, id := range area.SubareaIDs.Array {
			subarea, ok := subareaByID[id]
			if !ok {
				continue
			}
			subareaRefs = append(subareaRefs, mappedSubareaReference{AnkamaID: id, Name: localizedText(subarea.NameID, languages)})
		}
		mappedAreas = append(mappedAreas, mappedArea{
			AnkamaID: area.ID, Name: localizedText(area.NameID, languages),
			SuperArea: locationReference(area.SuperAreaID, super.NameID, found, languages), Subareas: subareaRefs,
		})
	}

	monsterIDsBySubarea := make(map[int][]int)
	for _, monster := range monsters {
		for _, subareaID := range monster.SubareaIDs.Array {
			if _, found := subareaByID[subareaID]; found {
				monsterIDsBySubarea[subareaID] = append(monsterIDsBySubarea[subareaID], monster.ID)
			}
		}
	}
	locationBySubarea := make(map[int]mappedSpawnLocation, len(subareas))
	mappedSubareas := make([]mappedSubarea, 0, len(subareas))
	for _, subarea := range subareas {
		area, areaFound := areaByID[subarea.AreaID]
		if !areaFound {
			return nil, nil, nil, nil, fmt.Errorf("subarea %d references missing area %d", subarea.ID, subarea.AreaID)
		}
		super, superFound := superByID[area.SuperAreaID]
		if !superFound {
			return nil, nil, nil, nil, fmt.Errorf("subarea %d references missing super area %d", subarea.ID, area.SuperAreaID)
		}
		areaRef := locationReference(subarea.AreaID, area.NameID, areaFound, languages)
		superRef := locationReference(area.SuperAreaID, super.NameID, areaFound && superFound, languages)
		mapRefs := make([]mappedMapReference, 0, len(subarea.MapIDs.Array))
		for _, id := range subarea.MapIDs.Array {
			value, found := mapByID[id]
			if !found {
				continue
			}
			mapRefs = append(mapRefs, mappedMapReference{AnkamaID: id, PositionX: value.PosX, PositionY: value.PosY, WorldMap: value.WorldMap})
		}
		monsterRefs := make([]mappedNamedReference, 0, len(monsterIDsBySubarea[subarea.ID]))
		location := mappedSpawnLocation{
			Subarea: locationReference(subarea.ID, subarea.NameID, true, languages), Area: areaRef, SuperArea: superRef, SubareaMaps: mapRefs,
		}
		locationBySubarea[subarea.ID] = location
		for _, id := range monsterIDsBySubarea[subarea.ID] {
			monster := monsterByID[id]
			monsterRefs = append(monsterRefs, namedReference(id, monster.NameID, true, languages))
		}
		mappedSubareas = append(mappedSubareas, mappedSubarea{
			AnkamaID: subarea.ID, Name: localizedText(subarea.NameID, languages), Level: subarea.Level,
			DungeonID: subarea.DungeonID, Area: areaRef, SuperArea: superRef, Maps: mapRefs, Monsters: monsterRefs,
		})
	}
	spawnByMonster := make(map[int][]mappedSpawnLocation)
	for _, monster := range monsters {
		for _, subareaID := range monster.SubareaIDs.Array {
			location, found := locationBySubarea[subareaID]
			if !found {
				continue
			}
			location.Preferred = monster.FavoriteSubareaID == subareaID
			spawnByMonster[monster.ID] = append(spawnByMonster[monster.ID], location)
		}
	}
	return mappedSuperAreas, mappedAreas, mappedSubareas, spawnByMonster, nil
}
