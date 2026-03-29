package game

import "io/fs"

func (r *CharacterRegistry) CountAssets(assets fs.FS, archs *ArchetypeRegistry, permitList map[string]bool) int {
	return len(r.createLoadJobs(assets, archs, permitList))
}
