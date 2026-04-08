package game

import "io/fs"

func (r *CharacterRegistry) CountAssets(assets fs.FS, archs *ArchetypeRegistry, permitList map[string]bool) int {
	if assets == nil { return 0 }
	return len(r.createLoadJobs(assets, archs, permitList))
}
