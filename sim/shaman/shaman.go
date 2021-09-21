package shaman

import (
	"time"

	"github.com/wowsims/tbc/sim/core"
)

type AgentType int

const (
	AgentTypeUnknown       = 0
	AgentTypeFixedLBCL     = 1
	AgentTypeCLOnClearcast = 2
	AgentTypeAdaptive      = 3
	AgentTypeEnhancer      = 4
)

func NewShaman(player *core.Player, party *core.Party, talents Talents, totems Totems, agentID int) *Shaman {
	agent := NewAdaptiveAgent(nil)

	// if WaterShield {
	player.InitialStats[core.StatMP5] += 50
	// }

	for _, pl := range party.Players {
		pl.InitialStats = totems.AddStats(pl.InitialStats)
		pl.Stats = pl.InitialStats
	}

	player.InitialStats[core.StatMP5] += player.InitialStats[core.StatIntellect] * (0.02 * float64(talents.UnrelentingStorm))
	player.Stats = player.InitialStats

	return &Shaman{
		agent:   agent,
		Player:  player,
		Talents: talents,

		convectionBonus: 1 - 0.02*float64(talents.Convection),
		concussionBonus: 1 + 0.01*float64(talents.Concussion),
	}
}

// Shaman represents a shaman player.
type Shaman struct {
	agent        shamanAgent // Controller logic
	Talents      Talents     // Shaman Talents
	*core.Player             // State of player

	// cache
	convectionBonus float64
	concussionBonus float64
}

// BuffUp lets you buff up all players in sim (and yourself)
func (s *Shaman) BuffUp(sim *core.Simulation, party *core.Party) {
	if s.Talents.LightningOverload > 0 {
		s.AddAura(sim, AuraLightningOverload(s.Talents.LightningOverload))
	}
}
func (s *Shaman) ChooseAction(sim *core.Simulation, party *core.Party) core.AgentAction {
	return s.agent.ChooseAction(s, party, sim)
}
func (s *Shaman) OnActionAccepted(sim *core.Simulation, action core.AgentAction) {
	s.agent.OnActionAccepted(s, sim, action)
}
func (s *Shaman) Reset(newsim *core.Simulation) {
	// Do we need to reset anything special?
}
func (s *Shaman) OnSpellHit(sim *core.Simulation, _ *core.Player, cast *core.Cast) {
	cast.DidDmg *= s.concussionBonus // add convection

	if cast.DidCrit && cast.Spell.ID != core.MagicIDTLCLB { // TLC does not proc focus.
		a := core.Aura{
			ID:             core.MagicIDEleFocus,
			Expires:        sim.CurrentTime + time.Second*15,
			Stacks:         2,
			OnCast:         elementalFocusOnCast,
			OnCastComplete: elementalFocusOnCastComplete,
		}
		s.AddAura(sim, a)
	}
}

func elementalFocusOnCast(sim *core.Simulation, player *core.Player, c *core.Cast) {
	c.ManaCost *= .6 // reduced by 40%
}

func elementalFocusOnCastComplete(sim *core.Simulation, player *core.Player, c *core.Cast) {
	if c.ManaCost <= 0 {
		return // Don't consume charges from free spells.
	}

	player.Auras[core.MagicIDEleFocus].Stacks--
	if player.Auras[core.MagicIDEleFocus].Stacks == 0 {
		player.RemoveAura(sim, player, core.MagicIDEleFocus)
	}
}

// Agent is shaman specific agent for behavior.
type shamanAgent interface {
	// Returns the action this Agent would like to take next.
	ChooseAction(*Shaman, *core.Party, *core.Simulation) core.AgentAction

	// This will be invoked if the chosen action is actually executed, so the Agent can update its state.
	OnActionAccepted(*Shaman, *core.Simulation, core.AgentAction)

	// Returns this Agent to its initial state.
	Reset(*core.Simulation)
}

type Totems struct {
	TotemOfWrath int
	WrathOfAir   bool
	ManaStream   bool
}

func (tt Totems) AddStats(s core.Stats) core.Stats {
	s[core.StatSpellCrit] += 66.24 * float64(tt.TotemOfWrath)
	s[core.StatSpellHit] += 37.8 * float64(tt.TotemOfWrath)
	if tt.WrathOfAir {
		s[core.StatSpellPower] += 101
	}
	if tt.ManaStream {
		s[core.StatMP5] += 50
	}
	return s
}

type Talents struct {
	LightningOverload  int
	ElementalPrecision int
	NaturesGuidance    int
	TidalMastery       int
	ElementalMastery   bool
	UnrelentingStorm   int
	CallOfThunder      int
	Convection         int
	Concussion         int
}

func TryActivateBloodlust(sim *core.Simulation, party *core.Party, player *core.Player) {
	if player.IsOnCD(core.MagicIDBloodlust, sim.CurrentTime) {
		return
	}

	dur := time.Second * 40 // assumes that multiple BLs are different shaman.
	player.SetCD(core.MagicIDBloodlust, time.Minute*10+sim.CurrentTime)

	for _, p := range party.Players {
		p.AddAura(sim, core.Aura{
			ID:      core.MagicIDBloodlust,
			Expires: sim.CurrentTime + dur,
			OnCast: func(sim *core.Simulation, p *core.Player, c *core.Cast) {
				c.CastTime = (c.CastTime * 10) / 13 // 30% faster
			},
		})
	}
}

// NewCastAction is how a shaman creates a new spell
//  TODO: Decide if we need separate functions for elemental and enhancement?
func NewCastAction(sim *core.Simulation, player *Shaman, sp *core.Spell) core.AgentAction {
	cast := core.NewCast(sim, sp)

	itsElectric := sp.ID == core.MagicIDCL6 || sp.ID == core.MagicIDLB12

	if player.Talents.ElementalPrecision > 0 {
		// FUTURE: This only impacts "frost fire and nature" spells.
		//  We know it doesnt impact TLC.
		//  Are there any other spells that a shaman can cast?
		cast.Hit += float64(player.Talents.ElementalPrecision) * 0.02
	}
	if player.Talents.NaturesGuidance > 0 {
		cast.Hit += float64(player.Talents.NaturesGuidance) * 0.01
	}
	if player.Talents.TidalMastery > 0 {
		cast.Crit += float64(player.Talents.TidalMastery) * 0.01
	}

	if itsElectric {
		// TODO: Should we change these to be full auras?
		//   Doesnt seem needed since they can only be used by shaman right here.
		if player.Equip[core.ItemSlotRanged].ID == 28248 {
			cast.Dmg += 55
		} else if player.Equip[core.ItemSlotRanged].ID == 23199 {
			cast.Dmg += 33
		} else if player.Equip[core.ItemSlotRanged].ID == 32330 {
			cast.Dmg += 85
		}
		if player.Talents.CallOfThunder > 0 { // only applies to CL and LB
			cast.Crit += float64(player.Talents.CallOfThunder) * 0.01
		}
		if sp.ID == core.MagicIDCL6 && sim.Options.Encounter.NumTargets > 1 {
			cast.DoItNow = ChainCast
		}
		cast.ManaCost *= player.convectionBonus

		// TODO: Add LightningMastery to talent list
		cast.CastTime -= time.Millisecond * 500 // Talent Lightning Mastery
	}
	cast.CastTime = time.Duration(float64(cast.CastTime) / player.HasteBonus())

	// Apply any on cast effects.
	for _, id := range player.ActiveAuraIDs {
		if player.Auras[id].OnCast != nil {
			player.Auras[id].OnCast(sim, player.Player, cast)
		}
	}
	if itsElectric { // TODO: Add ElementalFury talent
		// This is written this way to deal with making CSD dmg increase correct.
		cast.CritBonus *= 2 // This handles the 'Elemental Fury' talent which increases the crit bonus.
		cast.CritBonus -= 1 // reduce to multiplier instead of percent.
	}

	return core.AgentAction{
		Wait: 0,
		Cast: cast,
	}
}
