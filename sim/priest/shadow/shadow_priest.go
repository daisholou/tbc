package shadow

import (
	"fmt"
	"time"

	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
	"github.com/wowsims/tbc/sim/priest"
)

func RegisterShadowPriest() {
	core.RegisterAgentFactory(
		proto.Player_ShadowPriest{},
		func(character core.Character, options proto.Player) core.Agent {
			return NewShadowPriest(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_ShadowPriest)
			if !ok {
				panic("Invalid spec value for Shadow Priest!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewShadowPriest(character core.Character, options proto.Player) *ShadowPriest {
	shadowOptions := options.GetShadowPriest()

	// Only undead can do Dev Plague
	if shadowOptions.Rotation.UseDevPlague && options.Race != proto.Race_RaceUndead {
		shadowOptions.Rotation.UseDevPlague = false
	}
	// Only nelf can do starshards
	if shadowOptions.Rotation.UseStarshards && options.Race != proto.Race_RaceNightElf {
		shadowOptions.Rotation.UseStarshards = false
	}

	selfBuffs := priest.SelfBuffs{
		UseShadowfiend: shadowOptions.Options.UseShadowfiend,
	}

	basePriest := priest.New(character, selfBuffs, *shadowOptions.Talents)
	spriest := &ShadowPriest{
		Priest:   basePriest,
		rotation: *shadowOptions.Rotation,
	}

	spriest.ApplyShadowOnHitEffects()

	return spriest
}

type ShadowPriest struct {
	*priest.Priest

	rotation proto.ShadowPriest_Rotation
}

func (spriest *ShadowPriest) GetPriest() *priest.Priest {
	return spriest.Priest
}

func (spriest *ShadowPriest) Reset(sim *core.Simulation) {
	spriest.Priest.Reset(sim)
}

// TODO: probably do something different instead of making it global?
const (
	mbidx int = iota
	swdidx
	vtidx
	swpidx
)

func (spriest *ShadowPriest) OnGCDReady(sim *core.Simulation) {
	spriest.tryUseGCD(sim)
}

func (spriest *ShadowPriest) OnManaTick(sim *core.Simulation) {
	if spriest.FinishedWaitingForManaAndGCDReady(sim) {
		spriest.tryUseGCD(sim)
	}
}

func (spriest *ShadowPriest) tryUseGCD(sim *core.Simulation) {
	if spriest.rotation.PrecastVt && sim.CurrentTime == 0 {
		spell := spriest.NewVampiricTouch(sim, sim.GetPrimaryTarget())
		spell.CastTime = 0
		spell.GCD = 0
		spell.Cast(sim)
	}

	// This if block is to handle being able to cast a VT while having another one ticking.
	//  This will swap the casting spell if it is ticking, so the newly cast spell is now the ticking spell.
	//  spriest.NewVampiricTouch() will always use the priest.VTSpellCasting as the target.
	if spriest.VTSpellCasting.Effect.DotInput.IsTicking(sim) {
		if spriest.VTSpell.Effect.DotInput.TimeRemaining(sim) > 0 {
			// If we still have VT ticking that isn't allowed... crash immediately so we can fix the logic.
			panic(fmt.Sprintf("Two VT copies ticking at: %s, time remaining: %s, %s",
				sim.CurrentTime,
				spriest.VTSpellCasting.Effect.DotInput.TimeRemaining(sim),
				spriest.VTSpell.Effect.DotInput.TimeRemaining(sim)))
		}
		oldVT := spriest.VTSpell
		spriest.VTSpell = spriest.VTSpellCasting
		spriest.VTSpellCasting = oldVT // will probably have one more tick
	}

	// Activate shared behaviors
	target := sim.GetPrimaryTarget()
	var spell *core.SimpleSpell
	var wait1 time.Duration
	var wait2 time.Duration
	var wait time.Duration

	// calculate how much time a VT cast would take so we can possibly start casting right before the dot is up.
	castSpeed := spriest.CastSpeed()
	vtCastTime := time.Duration(float64(time.Millisecond*1500) / castSpeed)

	// timeForDots := sim.Duration-sim.CurrentTime > time.Second*12
	// TODO: stop casting dots near the end?

	if spriest.Talents.VampiricTouch && spriest.VTSpell.Effect.DotInput.TimeRemaining(sim) <= vtCastTime {
		spell = spriest.NewVampiricTouch(sim, target)
	} else if !spriest.SWPSpell.Effect.DotInput.IsTicking(sim) {
		spell = spriest.NewShadowWordPain(sim, target)
	} else if spriest.rotation.UseStarshards && spriest.GetRemainingCD(priest.SSCooldownID, sim.CurrentTime) == 0 {
		spell = spriest.NewStarshards(sim, target)
	} else if spriest.rotation.UseDevPlague && spriest.GetRemainingCD(priest.DevouringPlagueCooldownID, sim.CurrentTime) == 0 {
		spell = spriest.NewDevouringPlague(sim, target)
	} else if spriest.Talents.MindFlay {

		allCDs := []time.Duration{
			mbidx:  spriest.Character.GetRemainingCD(priest.MBCooldownID, sim.CurrentTime),
			swdidx: spriest.Character.GetRemainingCD(priest.SWDCooldownID, sim.CurrentTime),
			vtidx:  spriest.VTSpell.Effect.DotInput.TimeRemaining(sim) - vtCastTime,
			swpidx: spriest.SWPSpell.Effect.DotInput.TimeRemaining(sim),
		}

		if allCDs[mbidx] == 0 {
			if spriest.Talents.InnerFocus && spriest.GetRemainingCD(priest.InnerFocusCooldownID, sim.CurrentTime) == 0 {
				priest.ApplyInnerFocus(sim, spriest.Priest)
			}
			spell = spriest.NewMindBlast(sim, target)
		} else if allCDs[swdidx] == 0 {
			spell = spriest.NewShadowWordDeath(sim, target)
		} else {
			gcd := core.MinDuration(core.GCDMin, time.Duration(float64(core.GCDDefault)/castSpeed))
			tickLength := time.Duration(float64(time.Second) / castSpeed)

			var numTicks int
			switch spriest.rotation.RotationType {
			case proto.ShadowPriest_Rotation_Basic:
				numTicks = spriest.BasicMindflayRotation(sim, allCDs, gcd, tickLength)
			case proto.ShadowPriest_Rotation_Clipping:
				numTicks = spriest.ClippingMindflayRotation(sim, allCDs, gcd, tickLength)
			case proto.ShadowPriest_Rotation_Ideal:
				numTicks = spriest.IdealMindflayRotation(sim, allCDs, gcd, tickLength)
			}

			if numTicks == 0 {
				// Means we'd rather wait for next CD (swp, vt, etc) than start a MF cast.
				nextCD := core.NeverExpires
				for _, v := range allCDs {
					if v < nextCD {
						nextCD = v
					}
				}
				spriest.WaitUntil(sim, sim.CurrentTime+nextCD)
				return
			}

			spell = spriest.NewMindFlay(sim, target, numTicks)

			// if our channel is longer than GCD it will have human latency to end it beause you can't queue the next spell.
			if wait > gcd && spriest.rotation.Latency > 0 {
				base := spriest.rotation.Latency * 0.66
				variation := base + sim.RandomFloat("spriest latency")*base // should vary from 0.66 - 1.33 of given latency
				variation = core.MaxFloat(variation, 10)                    // no player can go under XXXms response time
				spell.AfterCastDelay += time.Duration(variation) * time.Millisecond
			}
		}
	} else {
		// what do you even do... i guess just sit around
		mbcd := spriest.Character.GetRemainingCD(priest.MBCooldownID, sim.CurrentTime)
		swdcd := spriest.Character.GetRemainingCD(priest.SWDCooldownID, sim.CurrentTime)
		vtidx := spriest.VTSpell.Effect.DotInput.TimeRemaining(sim) - vtCastTime
		swpidx := spriest.SWPSpell.Effect.DotInput.TimeRemaining(sim)
		wait1 = core.MinDuration(mbcd, swdcd)
		wait2 = core.MinDuration(vtidx, swpidx)
		wait = core.MinDuration(wait1, wait2)
		spriest.WaitUntil(sim, sim.CurrentTime+wait)
		return
	}

	if success := spell.Cast(sim); !success {
		spriest.WaitForMana(sim, spell.GetManaCost())
	}
}

// Returns the number of MF ticks to use, or 0 to wait for next CD.
func (spriest *ShadowPriest) BasicMindflayRotation(sim *core.Simulation, allCDs []time.Duration, gcd time.Duration, tickLength time.Duration) int {
	// just do MF3, never clipping
	nextCD := core.NeverExpires
	for _, v := range allCDs {
		if v < nextCD {
			nextCD = v
		}
	}
	// But don't start a MF if we can't get a single tick off.
	if nextCD < gcd {
		return 0
	} else {
		return 3
	}
}

// Returns the number of MF ticks to use, or 0 to wait for next CD.
func (spriest *ShadowPriest) IdealMindflayRotation(sim *core.Simulation, allCDs []time.Duration, gcd time.Duration, tickLength time.Duration) int {
	nextCD := core.NeverExpires
	nextIdx := -1
	for i, v := range allCDs {
		if v < nextCD {
			nextCD = v
			nextIdx = i
		}
	}

	var numTicks int
	if nextCD < gcd {
		numTicks = 0
	} else {
		numTicks = int(nextCD / tickLength)
	}

	critChance := (spriest.GetStat(stats.SpellCrit) / (core.SpellCritRatingPerCritChance * 100)) + (float64(spriest.Talents.ShadowPower) * 0.03)
	averageCritMultiplier := 1 + 0.5*critChance
	mfDamage := (528 + 0.57*(spriest.GetStat(stats.SpellPower))) * 0.3333

	if numTicks == 0 {
		//  calculate the dps gain from casting vs waiting.
		var Major_dmg float64
		if nextIdx == 0 {
			Major_dmg = (731.5 + spriest.GetStat(stats.SpellPower)*0.429) / (gcd + nextCD).Seconds() * averageCritMultiplier
		} else if nextIdx == 1 {
			Major_dmg = (618 + spriest.GetStat(stats.SpellPower)*0.429) / (gcd + nextCD).Seconds() * averageCritMultiplier
		} else if nextIdx == 2 {
			Major_dmg = (spriest.VTSpell.Effect.DotInput.DamagePerTick() * float64(spriest.VTSpell.Effect.DotInput.NumberOfTicks)) / (gcd + nextCD).Seconds()
		} else if nextIdx == 3 {
			Major_dmg = (spriest.SWPSpell.Effect.DotInput.DamagePerTick() * float64(spriest.SWPSpell.Effect.DotInput.NumberOfTicks)) / (gcd + nextCD).Seconds()
		}

		dpsPossibleshort := []float64{
			(Major_dmg * float64(nextCD+gcd)) / float64(gcd+nextCD),                          // dps with no tick and just wait
			(Major_dmg*(nextCD+gcd).Seconds() + mfDamage) / (gcd + gcd).Seconds(),            // new damage for 1 extra tick
			(Major_dmg*(nextCD+gcd).Seconds() + 2*mfDamage) / (2*tickLength + gcd).Seconds(), // new damage for 2 extra tick
			(Major_dmg*(nextCD+gcd).Seconds() + 3*mfDamage) / (3*tickLength + gcd).Seconds(), // new damage for 3 extra tick
		}

		// Find the highest possible dps and its index
		highestPossibleIdx := 0
		highestPossibleDmg := 0.0
		if highestPossibleIdx == 0 {
			for i, v := range dpsPossibleshort {
				if v >= highestPossibleDmg {
					highestPossibleIdx = i
					highestPossibleDmg = v
				}
			}
		}
		if highestPossibleIdx == 0 {
			return 0
		}
		numTicks = highestPossibleIdx

		// Now that the number of optimal ticks has been determined to optimize dps
		// Now optimize mf2s and mf3s
		if numTicks == 1 {
			return 1
		} else if numTicks == 2 || numTicks == 4 {
			return 2
		} else {
			return 3
		}
	}

	// TODO: Should spriest latency be added to the second option here?
	mfTime := core.MaxDuration(gcd, time.Duration(numTicks)*tickLength)

	// Amount of gap time after casting mind flay, but before each CD is available.
	cdDiffs := []time.Duration{
		allCDs[0] - mfTime,
		allCDs[1] - mfTime,
		allCDs[2] - mfTime,
		allCDs[3] - mfTime,
	}

	spellDamages := []float64{
		mbidx:  (731.5 + spriest.GetStat(stats.SpellPower)*0.429) / (gcd + cdDiffs[mbidx]).Seconds() * averageCritMultiplier,
		swdidx: (618 + spriest.GetStat(stats.SpellPower)*0.429) / (gcd + cdDiffs[swdidx]).Seconds() * averageCritMultiplier,
		vtidx:  (spriest.VTSpell.Effect.DotInput.DamagePerTick() * float64(spriest.VTSpell.Effect.DotInput.NumberOfTicks)) / (gcd + cdDiffs[vtidx]).Seconds(),
		swpidx: (spriest.SWPSpell.Effect.DotInput.DamagePerTick() * float64(spriest.SWPSpell.Effect.DotInput.NumberOfTicks)) / (gcd + cdDiffs[swpidx]).Seconds(),
	}

	bestIdx := 0
	bestDmg := 0.0
	for i, v := range spellDamages {
		if sim.Log != nil {
			//spriest.Log(sim, "\tSpellDamages[%d]: %01.f", i, v)
			//spriest.Log(sim, "\tcdDiffs[%d]: %0.1f", i, cdDiffs[i].Seconds())
		}
		if v > bestDmg {
			bestIdx = i
			bestDmg = v
		}
	}

	if nextIdx != bestIdx && cdDiffs[bestIdx] < time.Millisecond*1490 {
		numTicks = int(allCDs[bestIdx] / tickLength)
	}

	chosenWait := cdDiffs[bestIdx]
	if chosenWait > cdDiffs[nextIdx] && cdDiffs[nextIdx] < time.Millisecond*100 {
		chosenWait = cdDiffs[nextIdx]
	}

	finalMFStart := numTicks // Base ticks before adding additional

	//spriest.Log(sim, "CW %d", chosenWait)
	dpsPossible := []float64{
		bestDmg, // dps with no tick and just wait
		0,
		0,
		0,
	}
	dpsDuration := float64((chosenWait + gcd).Seconds())

	highestPossibleIdx := 0
	// TODO: Modified this slightly to expand time window, but it still doesn't change dps for any tests.
	// Probably can remove this entirely (and then also the if highestPossibleIdx == 0 right after)
	if (finalMFStart == 2) && (chosenWait <= tickLength && chosenWait > (tickLength-time.Millisecond*15)) {
		highestPossibleIdx = 1 // if the wait time is equal to an extra mf tick, and there are already 2 ticks, then just add 1
	}

	if highestPossibleIdx == 0 {
		switch finalMFStart {
		case 0:
			// this means that the extra ticks will be relative to starting a new mf cast entirely
			dpsPossible[1] = (bestDmg*dpsDuration + mfDamage) / float64(gcd+gcd)            // new damage for 1 extra tick
			dpsPossible[2] = (bestDmg*dpsDuration + 2*mfDamage) / float64(2*tickLength+gcd) // new damage for 2 extra tick
			dpsPossible[3] = (bestDmg*dpsDuration + 3*mfDamage) / float64(3*tickLength+gcd) // new damage for 3 extra tick
		case 1:
			total_check_time := 2 * tickLength

			if total_check_time < gcd {
				newDuration := float64((gcd + gcd).Seconds())
				dpsPossible[1] = (bestDmg*dpsDuration + (mfDamage * float64(finalMFStart+1))) / newDuration
			} else {
				newDuration := float64(((total_check_time - gcd) + gcd).Seconds())
				dpsPossible[1] = (bestDmg*dpsDuration + (mfDamage * float64(finalMFStart+1))) / newDuration
			}
			// % check add 2
			total_check_time2 := 2 * tickLength
			if total_check_time2 < gcd {
				dpsPossible[2] = (bestDmg*dpsDuration + (mfDamage * float64(finalMFStart+2))) / float64(gcd+gcd)
			} else {
				dpsPossible[2] = (bestDmg*dpsDuration + (mfDamage * float64(finalMFStart+2))) / float64(total_check_time2+gcd)
			}
		case 2:
			// % check add 1
			total_check_time := tickLength
			newDuration := float64((total_check_time + gcd).Seconds())
			dpsPossible[1] = (bestDmg*dpsDuration + mfDamage) / newDuration

		default:
			dpsPossible[1] = (bestDmg*dpsDuration + mfDamage) / float64(gcd+gcd)
			if tickLength*2 > gcd {
				dpsPossible[2] = (bestDmg*dpsDuration + 2*mfDamage) / float64(2*tickLength+gcd)
			} else {
				dpsPossible[2] = (bestDmg*dpsDuration + 2*mfDamage) / float64(gcd+gcd)
			}
			dpsPossible[3] = (bestDmg*dpsDuration + 3*mfDamage) / float64(3*tickLength+gcd)
		}
	}

	// Find the highest possible dps and its index
	// highestPossibleIdx := 0
	highestPossibleDmg := 0.0
	if highestPossibleIdx == 0 {
		for i, v := range dpsPossible {
			if sim.Log != nil {
				//spriest.Log(sim, "\tdpsPossible[%d]: %01.f", i, v)
			}
			if v >= highestPossibleDmg {
				highestPossibleIdx = i
				highestPossibleDmg = v
			}
		}
	}

	numTicks += highestPossibleIdx

	//  Now that the number of optimal ticks has been determined to optimize dps
	//  Now optimize mf2s and mf3s
	if numTicks == 1 {
		return 1
	} else if numTicks == 2 || numTicks == 4 {
		return 2
	} else {
		return 3
	}

	//  ONE BIG CAVEAT THAT STILL NEEDS WORK.. THIS NEEDS TO BE UPDATED TO INCLUDE HASTE PROCS THAT CAN OCCUR/DROP OFF MID MF SEQUENCE
}

// ClippingMindflayRotation is to be a 'sweaty but not perfect' rotation.
//  It will prioritize casting MB / SWD by clipping.
//  If there is 4s until the next CD it will use a 2xMF2 instead of 3+1.
//  Returns the number of MF ticks to use, or 0 to wait for next CD.
func (spriest *ShadowPriest) ClippingMindflayRotation(sim *core.Simulation, allCDs []time.Duration, gcd time.Duration, tickLength time.Duration) int {
	nextCD := core.NeverExpires
	for _, v := range allCDs[:2] {
		if v < nextCD {
			nextCD = v
		}
	}

	if sim.Log != nil {
		spriest.Log(sim, "<spriest> NextCD: %0.2f", nextCD.Seconds())
	}
	// This means a CD is coming up before we could cast a single MF
	if nextCD < gcd {
		return 0
	}

	// How many ticks we have time for.
	numTicks := int((nextCD - time.Duration(spriest.rotation.Latency)) / tickLength)

	if numTicks == 1 {
		return 1
	} else if numTicks == 2 || numTicks == 4 {
		return 2
	} else {
		return 3
	}
}
