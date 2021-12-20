import { Class } from '/tbc/core/proto/common.js';
import { RaidTarget } from '/tbc/core/proto/common.js';
import { Raid as RaidProto } from '/tbc/core/proto/api.js';
import { RaidStats as RaidStatsProto } from '/tbc/core/proto/api.js';
import { RaidBuffs } from '/tbc/core/proto/common.js';
import { Spec } from '/tbc/core/proto/common.js';
import { NO_TARGET } from '/tbc/core/proto_utils/utils.js';

import { Party, MAX_PARTY_SIZE } from './party.js';
import { Player } from './player.js';
import { TypedEvent } from './typed_event.js';
import { Sim } from './sim.js';
import { sum } from './utils.js';

export const MAX_NUM_PARTIES = 5;

// Manages all the settings for a single Raid.
export class Raid {
  private buffs: RaidBuffs = RaidBuffs.create();

	// Emits when a raid member is added/removed/moved.
  readonly compChangeEmitter = new TypedEvent<void>();

  readonly buffsChangeEmitter = new TypedEvent<void>();

  // Emits when anything in the raid changes.
  readonly changeEmitter = new TypedEvent<void>();

	// Should always hold exactly MAX_NUM_PARTIES elements.
	private parties: Array<Party>;

	private modifyRaidProto: ((raidProto: RaidProto) => void) = () => {};

	readonly sim: Sim;

  constructor(sim: Sim) {
		this.sim = sim;

		this.parties = [...Array(MAX_NUM_PARTIES).keys()].map(i => {
			const newParty = new Party(this, sim);
			newParty.compChangeEmitter.on(() => this.compChangeEmitter.emit());
			newParty.changeEmitter.on(() => this.changeEmitter.emit());
			return newParty;
		});

    [
      this.compChangeEmitter,
      this.buffsChangeEmitter,
    ].forEach(emitter => emitter.on(() => this.changeEmitter.emit()));
  }

	size(): number {
		return sum(this.parties.map(party => party.size()));
	}

	isEmpty(): boolean {
		return this.size() == 0;
	}

	getParties(): Array<Party> {
		// Make defensive copy.
		return this.parties.slice();
	}

	getParty(index: number): Party {
		return this.parties[index];
	}

	getPlayers(): Array<Player<any> | null> {
		return this.parties.map(party => party.getPlayers()).flat();
	}

	getPlayer(index: number): Player<any> | null {
		const party = this.parties[Math.floor(index / MAX_PARTY_SIZE)];
		return party.getPlayer(index % MAX_PARTY_SIZE);
	}

	getPlayerFromRaidTarget(raidTarget: RaidTarget): Player<any> | null {
		if (raidTarget.targetIndex == NO_TARGET) {
			return null;
		} else {
			return this.getPlayer(raidTarget.targetIndex);
		}
	}

	setPlayer(index: number, newPlayer: Player<any> | null) {
		const party = this.parties[Math.floor(index / MAX_PARTY_SIZE)];
		party.setPlayer(index % MAX_PARTY_SIZE, newPlayer);
	}

	getClassCount(playerClass: Class) {
		return this.getPlayers().filter(player => player != null && player.getClass() == playerClass).length;
	}

  getBuffs(): RaidBuffs {
    // Make a defensive copy
    return RaidBuffs.clone(this.buffs);
  }

  setBuffs(newBuffs: RaidBuffs) {
    if (RaidBuffs.equals(this.buffs, newBuffs))
      return;

    // Make a defensive copy
    this.buffs = RaidBuffs.clone(newBuffs);
    this.buffsChangeEmitter.emit();
  }

	setModifyRaidProto(newModFn: (raidProto: RaidProto) => void) {
		this.modifyRaidProto = newModFn;
	}

	toProto(): RaidProto {
		const raidProto = RaidProto.create({
			parties: this.parties.map(party => party.toProto()),
			buffs: this.buffs,
		});

		this.modifyRaidProto(raidProto);

		return raidProto;
	}

	fromProto(proto: RaidProto) {
		this.setBuffs(proto.buffs || RaidBuffs.create());

		for (let i = 0; i < MAX_NUM_PARTIES; i++) {
			if (proto.parties[i]) {
				this.parties[i].fromProto(proto.parties[i]);
			} else {
				this.parties[i].clear();
			}
		}
	}

  // Returns JSON representing all the current values.
  toJson(): Object {
		return {
			'parties': this.parties.map(party => party.toJson()),
      'buffs': RaidBuffs.toJson(this.buffs),
		};
  }

  // Set all the current values, assumes obj is the same type returned by toJson().
  fromJson(obj: any) {
		try {
			this.setBuffs(RaidBuffs.fromJson(obj['buffs']));
		} catch (e) {
			console.warn('Failed to parse raid buffs: ' + e);
		}

		if (obj['parties']) {
			for (let i = 0; i < MAX_NUM_PARTIES; i++) {
				const partyObj = obj['parties'][i];
				if (!partyObj) {
					this.parties[i].clear();
					continue;
				}

				this.parties[i].fromJson(partyObj);
			}
		}
  }
}
