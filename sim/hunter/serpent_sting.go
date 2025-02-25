package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

var SerpentStingDebuffID = core.NewDebuffID()
var SerpentStingActionID = core.ActionID{SpellID: 27016}

// Serpent sting uses the melee hit table for checking hit, but otherwise acts like
// a spell. So we have to wrap the Dot spell within a melee ability.
// TODO: Figure out a way to simplify this, and remove the metrics hack in core/metrics_aggregator.go.
func (hunter *Hunter) newSerpentStingDotTemplate(sim *core.Simulation) core.SimpleSpellTemplate {
	dotSpell := core.SimpleSpell{
		SpellCast: core.SpellCast{
			Cast: core.Cast{
				ActionID:       SerpentStingActionID,
				SpellSchool:    stats.NatureSpellPower,
				Character:      &hunter.Character,
				IgnoreManaCost: true,
			},
		},
		Effect: core.SpellHitEffect{
			SpellEffect: core.SpellEffect{
				DamageMultiplier:       1,
				StaticDamageMultiplier: 1,
				ThreatMultiplier:       1,
				IgnoreHitCheck:         true,
			},
			DotInput: core.DotDamageInput{
				NumberOfTicks:  5,
				TickLength:     time.Second * 3,
				TickBaseDamage: 0, // Calculated on application
				DebuffID:       SerpentStingDebuffID,
			},
		},
	}
	dotSpell.Effect.StaticDamageMultiplier *= 1 + 0.06*float64(hunter.Talents.ImprovedStings)
	return core.NewSimpleSpellTemplate(dotSpell)
}

func (hunter *Hunter) newSerpentStingTemplate(sim *core.Simulation) core.MeleeAbilityTemplate {
	ama := core.ActiveMeleeAbility{
		MeleeAbility: core.MeleeAbility{
			ActionID:    SerpentStingActionID,
			Character:   &hunter.Character,
			SpellSchool: stats.NatureSpellPower,
			GCD:         core.GCDDefault,
			Cost: core.ResourceCost{
				Type:  stats.Mana,
				Value: 275,
			},
		},
		Effect: core.AbilityHitEffect{
			WeaponInput: core.WeaponDamageInput{
				IsRanged: true,
			},
		},
		OnMeleeAttack: func(sim *core.Simulation, ability *core.ActiveMeleeAbility, hitEffect *core.AbilityHitEffect) {
			if !hitEffect.Landed() {
				return
			}

			dot := &hunter.serpentStingDot
			hunter.serpentStingDotTemplate.Apply(dot)

			// Set dynamic fields, i.e. the stuff we couldn't precompute.
			dot.Effect.Target = hitEffect.Target
			// TODO: This should probably include AP from mark of the champion / elixir of demonslaying / target debuffs
			dot.Effect.DotInput.TickBaseDamage = 132 + hunter.GetStat(stats.RangedAttackPower)*0.02

			dot.Init(sim)
			dot.Cast(sim)
		},
	}

	ama.Cost.Value *= 1 - 0.02*float64(hunter.Talents.Efficiency)
	ama.Effect.BonusCritRating = -100 * core.MeleeCritRatingPerCritChance // Prevent crits

	return core.NewMeleeAbilityTemplate(ama)
}

func (hunter *Hunter) NewSerpentSting(sim *core.Simulation, target *core.Target) *core.ActiveMeleeAbility {
	ss := &hunter.serpentSting
	hunter.serpentStingTemplate.Apply(ss)

	// Set dynamic fields, i.e. the stuff we couldn't precompute.
	ss.Effect.Target = target

	return ss
}
