package priest

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

const SpellIDStarshards int32 = 25446

var SSCooldownID = core.NewCooldownID()

func (priest *Priest) newStarshardsTemplate(sim *core.Simulation) core.SimpleSpellTemplate {
	baseCast := core.Cast{
		ActionID: core.ActionID{
			SpellID:    SpellIDStarshards,
			CooldownID: SSCooldownID,
		},
		Character:    &priest.Character,
		SpellSchool:  stats.ArcaneSpellPower,
		BaseManaCost: 0,
		ManaCost:     0,
		CastTime:     0,
		GCD:          core.GCDDefault,
		Cooldown:     time.Second * 30,
	}

	effect := core.SpellHitEffect{
		SpellEffect: core.SpellEffect{
			DamageMultiplier:       1,
			StaticDamageMultiplier: 1,
			ThreatMultiplier:       1,
		},
		DotInput: core.DotDamageInput{
			NumberOfTicks:        5,
			TickLength:           time.Second * 3,
			TickBaseDamage:       785 / 5,
			TickSpellCoefficient: 0.167,
		},
	}

	return core.NewSimpleSpellTemplate(core.SimpleSpell{
		SpellCast: core.SpellCast{
			Cast: baseCast,
		},
		Effect: effect,
	})
}

func (priest *Priest) NewStarshards(sim *core.Simulation, target *core.Target) *core.SimpleSpell {
	// Initialize cast from precomputed template.
	mf := &priest.StarshardsSpell

	priest.starshardsTemplate.Apply(mf)

	// Set dynamic fields, i.e. the stuff we couldn't precompute.
	mf.Effect.Target = target
	mf.Init(sim)

	return mf
}
