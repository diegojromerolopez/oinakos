package game

// RegistryContainer bundles all game registries for easier management and passing.
type RegistryContainer struct {
	Archetypes         *ArchetypeRegistry
	PlayableCharacters *PlayableCharacterRegistry
	Maps               *MapTypeRegistry
	Campaigns          *CampaignRegistry
	Obstacles          *ObstacleRegistry
	NPCs               *NPCRegistry
}

func NewRegistryContainer() *RegistryContainer {
	return &RegistryContainer{
		Archetypes:         NewArchetypeRegistry(),
		PlayableCharacters: NewPlayableCharacterRegistry(),
		Maps:               NewMapTypeRegistry(),
		Campaigns:          NewCampaignRegistry(),
		Obstacles:          NewObstacleRegistry(),
		NPCs:               NewNPCRegistry(),
	}
}
