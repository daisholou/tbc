package hunter

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

var ArcaneShotCooldownID = core.NewCooldownID()
var ArcaneShotActionID = core.ActionID{SpellID: 27019, CooldownID: ArcaneShotCooldownID}

func (hunter *Hunter) newArcaneShotTemplate(sim *core.Simulation) core.MeleeAbilityTemplate {
	ama := core.ActiveMeleeAbility{
		MeleeAbility: core.MeleeAbility{
			ActionID:    ArcaneShotActionID,
			Character:   &hunter.Character,
			SpellSchool: stats.ArcaneSpellPower,
			GCD:         core.GCDDefault,
			Cooldown:    time.Second * 6,
			Cost: core.ResourceCost{
				Type:  stats.Mana,
				Value: 230,
			},
			CritMultiplier: hunter.critMultiplier(true, sim.GetPrimaryTarget()),
		},
		Effect: core.AbilityHitEffect{
			AbilityEffect: core.AbilityEffect{
				DamageMultiplier:       1,
				StaticDamageMultiplier: 1,
				ThreatMultiplier:       1,
				IgnoreArmor:            true,
			},
			WeaponInput: core.WeaponDamageInput{
				IsRanged: true,
				CalculateDamage: func(attackPower float64, bonusWeaponDamage float64) float64 {
					return attackPower*0.15 + 273
				},
			},
		},
	}

	ama.Cost.Value *= 1 - 0.02*float64(hunter.Talents.Efficiency)
	ama.Cooldown -= time.Millisecond * 200 * time.Duration(hunter.Talents.ImprovedArcaneShot)

	return core.NewMeleeAbilityTemplate(ama)
}

func (hunter *Hunter) NewArcaneShot(sim *core.Simulation, target *core.Target) *core.ActiveMeleeAbility {
	as := &hunter.arcaneShot
	hunter.arcaneShotTemplate.Apply(as)

	// Set dynamic fields, i.e. the stuff we couldn't precompute.
	as.Effect.Target = target

	// Arcane shot is super weird, because its a melee ability but it uses arcane
	// modifiers instead of physical. Luckily, CoE and Misery are the only modifiers
	// for arcane in the game so we can hardcode them here.
	if target.HasAura(core.MiseryDebuffID) {
		as.Effect.DamageMultiplier *= 1.05
	}
	if target.HasAura(core.CurseOfElementsDebuffID) {
		level := target.NumStacks(core.CurseOfElementsDebuffID)
		as.Effect.DamageMultiplier *= 1.1 + 0.01*float64(level)
	}

	return as
}
