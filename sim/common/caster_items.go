package common

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/stats"
)

// Keep these (and their functions) in alphabetical order.
func init() {
	// Proc effects. Keep these in order by item ID.
	core.AddItemEffect(28602, ApplyRobeOfTheElderScribes)
}

func ApplyRobeOfTheElderScribes(agent core.Agent) {
	character := agent.GetCharacter()
	character.AddPermanentAura(func(sim *core.Simulation) core.Aura {
		// Gives a chance when your harmful spells land to increase the damage of your spells and effects by up to 130 for 10 sec. (Proc chance: 20%, 50s cooldown)
		icd := core.NewICD()
		const spellBonus = 130.0
		const dur = time.Second * 10
		const icdDur = time.Second * 50
		const proc = 0.2

		return core.Aura{
			ID:      core.MagicIDElderScribe,
			OnSpellHit: func(sim *core.Simulation, cast core.DirectCastAction, result *core.DirectCastDamageResult) {
				if !icd.IsOnCD(sim) && sim.Rando.Float64("unmarked") < proc {
					icd = core.InternalCD(sim.CurrentTime + icdDur)
					core.AddAuraWithTemporaryStats(sim, character, core.MagicIDElderScribeProc, stats.SpellPower, spellBonus, dur)
				}
			},
		}
	})
}
