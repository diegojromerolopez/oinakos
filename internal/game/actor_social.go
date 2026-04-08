package game

func (a *Actor) GetRelationshipTier(otherID string) string {
	if a.Relationships == nil { a.Relationships = make(map[string]float64) }
	if a.RomanticInterest == nil { a.RomanticInterest = make(map[string]float64) }

	sentiment := a.Relationships[otherID]
	passion := a.RomanticInterest[otherID]

	// Order of precedence: Strongest feelings first
	if passion > 90 && sentiment > 90 { return "Devoted" }
	if passion > 40 { return "Romantic" }

	if sentiment < -70 { return "Enemy" }
	if sentiment < -30 { return "Antagonistic" }
	if sentiment > 70 { return "Friendly" }
	if sentiment > 30 { return "Acquaintance" }

	return "Neutral"
}

// GetEffectiveSentiment returns the relationship sentiment adjusted for external factors like hygiene.
func (a *Actor) GetEffectiveSentiment(other *Actor) float64 {
	if a.Relationships == nil { a.Relationships = make(map[string]float64) }
	sentiment := a.Relationships[other.Name]
	
	// Hygiene penalty: if hygiene is below 50, it starts affecting how others like you
	if other.State.Hygiene < 50 {
		penalty := (50 - other.State.Hygiene) * 0.5 // Max -25
		sentiment -= penalty
	}
	
	return sentiment
}

// ModifySentiment adjusts the relationship value with another actor.
func (a *Actor) ModifySentiment(otherID string, delta float64) {
	if a.Relationships == nil { a.Relationships = make(map[string]float64) }
	a.Relationships[otherID] += delta
	if a.Relationships[otherID] < -100 { a.Relationships[otherID] = -100 }
	if a.Relationships[otherID] > 100 { a.Relationships[otherID] = 100 }
}

// ModifyRomanticInterest adjusts passion levels.
func (a *Actor) ModifyRomanticInterest(otherID string, delta float64) {
	if a.RomanticInterest == nil { a.RomanticInterest = make(map[string]float64) }
	a.RomanticInterest[otherID] += delta
	if a.RomanticInterest[otherID] < 0 { a.RomanticInterest[otherID] = 0 }
	if a.RomanticInterest[otherID] > 100 { a.RomanticInterest[otherID] = 100 }
}

// ModifySubmission adjusts the submission level toward another actor.
func (a *Actor) ModifySubmission(otherID string, delta float64) {
	if a.Submission == nil { a.Submission = make(map[string]float64) }
	a.Submission[otherID] += delta
	if a.Submission[otherID] < 0 { a.Submission[otherID] = 0 }
	if a.Submission[otherID] > 100 { a.Submission[otherID] = 100 }
}

// GetSubmissionLevel returns the descriptive tier of submission (e.g. Subservient).
func (a *Actor) GetSubmissionLevel(otherID string) string {
	if a.Submission == nil { return "Rebellious" }
	val := a.Submission[otherID]
	if val < 10 { return "Rebellious" }
	if val < 30 { return "Hesitant" }
	if val < 60 { return "Obedient" }
	if val < 90 { return "Subservient" }
	return "Cowering"
}
