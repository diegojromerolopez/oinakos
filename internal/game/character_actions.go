package game

func (c *Character) GetActor() *Actor { return &c.Actor }
func (c *Character) IsIncapacitated() bool { return c.ActionState == ActorIncapacitated }
