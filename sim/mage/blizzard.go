package mage

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

const SpellIDBlizzard int32 = 27085

func (mage *Mage) newBlizzardTemplate(sim *core.Simulation) core.SimpleSpellTemplate {
	spell := core.SimpleSpell{
		SpellCast: core.SpellCast{
			Cast: core.Cast{
				SpellSchool:  stats.FrostSpellPower,
				Character:    &mage.Character,
				BaseManaCost: 1645,
				ManaCost:     1645,
				GCD:          core.GCDDefault,
				ActionID: core.ActionID{
					SpellID: SpellIDBlizzard,
				},
			},
		},
		IsChannel: true,
		AOECap:    3620,
	}

	baseEffect := core.SpellHitEffect{
		SpellEffect: core.SpellEffect{
			DamageMultiplier:       1,
			StaticDamageMultiplier: mage.spellDamageMultiplier,
			ThreatMultiplier:       1 - (0.1/3)*float64(mage.Talents.FrostChanneling),
			IgnoreHitCheck:         true,
		},
		DotInput: core.DotDamageInput{
			NumberOfTicks:        8,
			TickLength:           time.Second * 1,
			TickBaseDamage:       184,
			TickSpellCoefficient: 0.119,
			AffectedByCastSpeed:  true,
		},
	}

	spell.ManaCost -= spell.BaseManaCost * float64(mage.Talents.FrostChanneling) * 0.05
	spell.ManaCost *= 1 - float64(mage.Talents.ElementalPrecision)*0.01
	baseEffect.BonusSpellHitRating += float64(mage.Talents.ElementalPrecision) * 1 * core.SpellHitRatingPerHitChance
	baseEffect.StaticDamageMultiplier *= 1 + 0.02*float64(mage.Talents.PiercingIce)
	baseEffect.StaticDamageMultiplier *= 1 + 0.01*float64(mage.Talents.ArcticWinds)

	numHits := sim.GetNumTargets()
	effects := make([]core.SpellHitEffect, 0, numHits)
	for i := int32(0); i < numHits; i++ {
		effects = append(effects, baseEffect)
		effects[i].Target = sim.GetTarget(i)
	}
	spell.Effects = effects

	return core.NewSimpleSpellTemplate(spell)
}

func (mage *Mage) NewBlizzard(sim *core.Simulation) *core.SimpleSpell {
	// Initialize cast from precomputed template.
	blizzard := &mage.blizzardSpell
	mage.blizzardCastTemplate.Apply(blizzard)

	// Set dynamic fields, i.e. the stuff we couldn't precompute.
	blizzard.Init(sim)

	return blizzard
}
