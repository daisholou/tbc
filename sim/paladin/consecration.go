package paladin

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

const SpellIDConsecration int32 = 27173

func (paladin *Paladin) newConsecrationTemplate(sim *core.Simulation) core.SimpleSpellTemplate {

	spell := core.SimpleSpell{
		SpellCast: core.SpellCast{
			Cast: core.Cast{
				ActionID: core.ActionID{
					SpellID: SpellIDConsecration,
				},
				Character:    &paladin.Character,
				SpellSchool:  stats.HolySpellPower,
				BaseManaCost: 660,
				ManaCost:     660,
				GCD:          core.GCDDefault,
			},
		},
	}

	effect := core.SpellHitEffect{
		SpellEffect: core.SpellEffect{
			DamageMultiplier:       1,
			StaticDamageMultiplier: 1,
			IgnoreHitCheck:         true,
		},
		DotInput: core.DotDamageInput{
			NumberOfTicks:        8,
			TickLength:           time.Second,
			TickBaseDamage:       64,
			TickSpellCoefficient: 0.119,
		},
	}

	// TODO: consecration talents here

	numHits := sim.GetNumTargets()
	effects := make([]core.SpellHitEffect, 0, numHits)
	for i := int32(0); i < numHits; i++ {
		effects = append(effects, effect)
		effects[i].Target = sim.GetTarget(i)
	}
	spell.Effects = effects

	return core.NewSimpleSpellTemplate(spell)
}

func (paladin *Paladin) NewConsecration(sim *core.Simulation) *core.SimpleSpell {
	// Cancel the current flamestrike dot.
	paladin.ConsecrationSpell.Cancel(sim)

	consecration := &paladin.ConsecrationSpell
	paladin.consecrationTemplate.Apply(consecration)

	// Set dynamic fields, i.e. the stuff we couldn't precompute.
	consecration.Init(sim)

	return consecration
}
