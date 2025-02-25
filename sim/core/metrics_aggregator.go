package core

import (
	"math"
	"time"

	"github.com/wowsims/tbc/sim/core/proto"
)

// A unique number based on an ActionID.
// This works by making item IDs negative to avoid collisions, and assumes
// there are no collisions with OtherID. Tag adds decimals.
// Actual key values dont matter, just need something unique and fast to compute.
type ActionKey float64

func NewActionKey(actionID ActionID) ActionKey {
	return ActionKey(float64((int32(actionID.OtherID) + actionID.SpellID - actionID.ItemID)) + (float64(actionID.Tag) / 256))
}

type DistributionMetrics struct {
	// Values for the current iteration. These are cleared after each iteration.
	Total float64

	// Aggregate values. These are updated after each iteration.
	sum        float64
	sumSquared float64
	max        float64
	hist       map[int32]int32 // rounded DPS to count
}

func (distMetrics *DistributionMetrics) reset() {
	distMetrics.Total = 0
}

// This should be called when a Sim iteration is complete.
func (distMetrics *DistributionMetrics) doneIteration(encounterDurationSeconds float64) {
	dps := distMetrics.Total / encounterDurationSeconds

	distMetrics.sum += dps
	distMetrics.sumSquared += dps * dps
	distMetrics.max = MaxFloat(distMetrics.max, dps)

	dpsRounded := int32(math.Round(dps/10) * 10)
	distMetrics.hist[dpsRounded]++
}

func (distMetrics *DistributionMetrics) ToProto(numIterations int32) *proto.DistributionMetrics {
	dpsAvg := distMetrics.sum / float64(numIterations)

	return &proto.DistributionMetrics{
		Avg:   dpsAvg,
		Stdev: math.Sqrt((distMetrics.sumSquared / float64(numIterations)) - (dpsAvg * dpsAvg)),
		Max:   distMetrics.max,
		Hist:  distMetrics.hist,
	}
}

func NewDistributionMetrics() DistributionMetrics {
	return DistributionMetrics{
		hist: make(map[int32]int32),
	}
}

type CharacterMetrics struct {
	dps    DistributionMetrics
	threat DistributionMetrics

	CharacterIterationMetrics

	// Aggregate values. These are updated after each iteration.
	oomTimeSum float64
	actions    map[ActionKey]ActionMetrics
}

// Metrics for the current iteration, for 1 agent. Keep this as a separate
// struct so its easy to clear.
type CharacterIterationMetrics struct {
	WentOOM bool // Whether the agent has hit OOM at least once in this iteration.

	ManaSpent       float64
	ManaGained      float64
	BonusManaGained float64 // Only includes amount from mana pots / runes / innervates.

	OOMTime time.Duration // time spent not casting and waiting for regen.
}

type ActionMetrics struct {
	ActionID ActionID
	IsMelee  bool // True if melee action, false if spell action.

	Casts  int32
	Hits   int32
	Crits  int32
	Misses int32

	// These will be 0 for spell actions.
	Dodges  int32
	Parries int32
	Blocks  int32
	Glances int32

	Damage float64
}

func (actionMetrics *ActionMetrics) ToProto() *proto.ActionMetrics {
	// Hack because serpent sting is super weird
	casts := actionMetrics.Casts
	hits := actionMetrics.Hits
	if actionMetrics.ActionID.SpellID == 27016 {
		extras := hits / 2
		hits -= extras
		casts -= extras
	}

	return &proto.ActionMetrics{
		Id:      actionMetrics.ActionID.ToProto(),
		IsMelee: actionMetrics.IsMelee,

		Casts:   casts,
		Hits:    hits,
		Crits:   actionMetrics.Crits,
		Misses:  actionMetrics.Misses,
		Dodges:  actionMetrics.Dodges,
		Parries: actionMetrics.Parries,
		Blocks:  actionMetrics.Blocks,
		Glances: actionMetrics.Glances,
		Damage:  actionMetrics.Damage,
	}
}

func NewCharacterMetrics() CharacterMetrics {
	return CharacterMetrics{
		dps:     NewDistributionMetrics(),
		threat:  NewDistributionMetrics(),
		actions: make(map[ActionKey]ActionMetrics),
	}
}

func (characterMetrics *CharacterMetrics) addCastInternal(actionID ActionID) {
	actionKey := NewActionKey(actionID)
	actionMetrics, ok := characterMetrics.actions[actionKey]

	if !ok {
		actionMetrics.ActionID = actionID
	}

	actionMetrics.Casts++

	characterMetrics.actions[actionKey] = actionMetrics
}

func (characterMetrics *CharacterMetrics) AddInstantCast(actionID ActionID) {
	characterMetrics.addCastInternal(actionID)
}

// Adds the results of a cast to the aggregated metrics.
func (characterMetrics *CharacterMetrics) AddCast(cast *Cast) {
	characterMetrics.addCastInternal(cast.ActionID)
}

// Adds the results of an action to the aggregated metrics.
func (characterMetrics *CharacterMetrics) AddSpellCast(spellCast *SpellCast) {
	actionID := spellCast.ActionID
	actionKey := NewActionKey(actionID)
	actionMetrics, ok := characterMetrics.actions[actionKey]

	if !ok {
		actionMetrics.ActionID = actionID
	}

	actionMetrics.Casts++
	actionMetrics.Hits += spellCast.Hits
	actionMetrics.Misses += spellCast.Misses
	actionMetrics.Crits += spellCast.Crits
	actionMetrics.Damage += spellCast.TotalDamage
	characterMetrics.dps.Total += spellCast.TotalDamage
	characterMetrics.threat.Total += spellCast.TotalThreat

	characterMetrics.actions[actionKey] = actionMetrics
}

// Adds the results of a melee action to the aggregated metrics.
func (characterMetrics *CharacterMetrics) AddMeleeAbility(ability *ActiveMeleeAbility) {
	actionID := ability.ActionID
	actionKey := NewActionKey(actionID)
	actionMetrics, ok := characterMetrics.actions[actionKey]

	if !ok {
		actionMetrics.ActionID = actionID
		actionMetrics.IsMelee = true
	}

	actionMetrics.Casts++
	actionMetrics.Hits += ability.Hits
	actionMetrics.Misses += ability.Misses
	actionMetrics.Crits += ability.Crits
	actionMetrics.Dodges += ability.Dodges
	actionMetrics.Parries += ability.Parries
	actionMetrics.Blocks += ability.Blocks
	actionMetrics.Glances += ability.Glances
	actionMetrics.Damage += ability.TotalDamage
	characterMetrics.dps.Total += ability.TotalDamage
	characterMetrics.threat.Total += ability.TotalThreat

	characterMetrics.actions[actionKey] = actionMetrics
}

// This should be called at the end of each iteration, to include metrics from Pets in
// those of their owner.
// Assumes that doneIteration() has already been called on the pet metrics.
func (characterMetrics *CharacterMetrics) AddFinalPetMetrics(petMetrics *CharacterMetrics) {
	characterMetrics.dps.Total += petMetrics.dps.Total
}

func (characterMetrics *CharacterMetrics) MarkOOM(character *Character, dur time.Duration) {
	characterMetrics.CharacterIterationMetrics.OOMTime += dur
	characterMetrics.CharacterIterationMetrics.WentOOM = true
}

func (characterMetrics *CharacterMetrics) reset() {
	characterMetrics.dps.reset()
	characterMetrics.threat.reset()
	characterMetrics.CharacterIterationMetrics = CharacterIterationMetrics{}
}

// This should be called when a Sim iteration is complete.
func (characterMetrics *CharacterMetrics) doneIteration(encounterDurationSeconds float64) {
	characterMetrics.dps.doneIteration(encounterDurationSeconds)
	characterMetrics.threat.doneIteration(encounterDurationSeconds)
	characterMetrics.oomTimeSum += float64(characterMetrics.OOMTime.Seconds())
}

func (characterMetrics *CharacterMetrics) ToProto(numIterations int32) *proto.PlayerMetrics {
	protoMetrics := &proto.PlayerMetrics{
		Dps:           characterMetrics.dps.ToProto(numIterations),
		Threat:        characterMetrics.threat.ToProto(numIterations),
		SecondsOomAvg: characterMetrics.oomTimeSum / float64(numIterations),
	}

	for _, action := range characterMetrics.actions {
		protoMetrics.Actions = append(protoMetrics.Actions, action.ToProto())
	}

	return protoMetrics
}

type AuraMetrics struct {
	ID ActionID

	// Metrics for the current iteration.
	Uptime time.Duration

	// Aggregate values. These are updated after each iteration.
	uptimeSum        time.Duration
	uptimeSumSquared time.Duration
}

func (auraMetrics *AuraMetrics) reset() {
	auraMetrics.Uptime = 0
}

// This should be called when a Sim iteration is complete.
func (auraMetrics *AuraMetrics) doneIteration() {
	auraMetrics.uptimeSum += auraMetrics.Uptime
	auraMetrics.uptimeSumSquared += auraMetrics.Uptime * auraMetrics.Uptime
}

func (auraMetrics *AuraMetrics) ToProto(numIterations int32) *proto.AuraMetrics {
	uptimeAvg := auraMetrics.uptimeSum.Seconds() / float64(numIterations)

	return &proto.AuraMetrics{
		Id: auraMetrics.ID.ToProto(),

		UptimeSecondsAvg:   uptimeAvg,
		UptimeSecondsStdev: math.Sqrt((auraMetrics.uptimeSumSquared.Seconds() / float64(numIterations)) - (uptimeAvg * uptimeAvg)),
	}
}

// Calculates DPS for an action.
func GetActionDPS(playerMetrics proto.PlayerMetrics, iterations int32, duration time.Duration, actionID ActionID, ignoreTag bool) float64 {
	totalDPS := 0.0
	for _, action := range playerMetrics.Actions {
		metricsActionID := ProtoToActionID(*action.Id)
		if actionID.SameAction(metricsActionID) || (ignoreTag && actionID.SameActionIgnoreTag(metricsActionID)) {
			totalDPS += action.Damage / float64(iterations) / duration.Seconds()
		}
	}
	return totalDPS
}

// Calculates average cast damage for an action.
func GetActionAvgCast(playerMetrics proto.PlayerMetrics, actionID ActionID) float64 {
	for _, action := range playerMetrics.Actions {
		if actionID.SameAction(ProtoToActionID(*action.Id)) {
			return action.Damage / float64(action.Casts)
		}
	}
	return 0
}
