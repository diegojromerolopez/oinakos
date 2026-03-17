package game

// RegistryContainer bundles all game registries for easier management and passing.
type RegistryContainer struct {
	Archetypes         *ArchetypeRegistry
	Characters         *CharacterRegistry
	Maps               *MapTypeRegistry
	Campaigns          *CampaignRegistry
	Obstacles          *ObstacleRegistry
	Objects            *ObjectRegistry
}

func NewRegistryContainer() *RegistryContainer {
	return &RegistryContainer{
		Archetypes: NewArchetypeRegistry(),
		Characters: NewCharacterRegistry(),
		Maps:       NewMapTypeRegistry(),
		Campaigns:  NewCampaignRegistry(),
		Obstacles:  NewObstacleRegistry(),
		Objects:    NewObjectRegistry(),
	}
}
