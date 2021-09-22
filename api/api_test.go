package api

import (
	"log"
	"testing"
)

var basicSpec = &PlayerOptions_ElementalShaman{
	ElementalShaman: &ElementalShaman{
		Agent: &ElementalShaman_ElementalShamanAgent{
			Type: ElementalShaman_ElementalShamanAgent_Adaptive,
		},
		Talents: &ShamanTalents{
			// ElementalDevastation
			// ElementalFury
			Convection:         5,
			Concussion:         5,
			ElementalFocus:     true,
			CallOfThunder:      5,
			UnrelentingStorm:   3,
			ElementalPrecision: 3,
			LightningMastery:   5,
			ElementalMastery:   true,
			LightningOverload:  5,
		},
		Options: &ElementalShaman_ElementalShamanOptions{
			WaterShield: true,
		},
	},
}

var basicConsumes = &Consumes{
	FlaskOfBlindingLight: true,
	BlackenedBasilisk:    true,
	BrilliantWizardOil:   true,
	SuperManaPotion:      true,
	DarkRune:             true,
}

var basicBuffs = &Buffs{
	ArcaneBrilliance: true,
	BlessingOfKings:  true,
	Bloodlust:        1,
	MoonkinAura:      TristateEffect_TristateEffectRegular,
	ManaSpringTotem:  TristateEffect_TristateEffectRegular,
	TotemOfWrath:     1,
	WrathOfAirTotem:  TristateEffect_TristateEffectRegular,
}

var p1Equip = &EquipmentSpec{
	Items: []*ItemSpec{
		{Id: 29035, Gems: []int32{34220, 24059}, Enchant: 29191},
		{Id: 28762},
		{Id: 29037, Gems: []int32{24059, 24059}, Enchant: 28909},
		{Id: 28766},
		{Id: 29519},
		{Id: 29521},
		{Id: 28780},
		{Id: 29520},
		{Id: 30541},
		{Id: 28810},
		{Id: 30667},
		{Id: 28753},
		{Id: 28785},
		{Id: 29370},
		{Id: 28248},
		{Id: 28770, Enchant: 22555},
		{Id: 29268},
	},
}

// TestIndividualSim is designed to test the conversion of proto objects to
//  internal objects. This should not be a comprehensive test of the internals of the simulator.
//  It might be worth adding more features to ensure they all convert properly though!
//  Perhaps instead of running a real sim we just test that the output objects from the conversion functions work properly.
func TestIndividualSim(t *testing.T) {
	req := &IndividualSimRequest{
		Player: &Player{
			Options: &PlayerOptions{
				Race:     Race_RaceTroll10,
				Spec:     basicSpec,
				Consumes: basicConsumes,
			},
			Equipment: p1Equip,
		},
		Buffs: basicBuffs,
		Encounter: &Encounter{
			Duration:   120,
			NumTargets: 1,
		},
		Iterations: 5000,
		RandomSeed: 1,
		Debug:      false,
	}

	res := RunSimulation(req)

	log.Printf("LOGS:\n%s\n", res.Logs)

	if res.DpsAvg != 100 {
		t.Fatalf("DPS: %0.1f", res.DpsAvg)
	}
}
