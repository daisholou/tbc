import { Party as PartyProto } from '/tbc/core/proto/api.js';
import { PartyBuffs } from '/tbc/core/proto/common.js';
import { Player } from './player.js';
import { TypedEvent } from './typed_event.js';
export const MAX_PARTY_SIZE = 5;
// Manages all the settings for a single Party.
export class Party {
    constructor(raid, sim) {
        this.buffs = PartyBuffs.create();
        // Emits when a party member is added/removed/moved.
        this.compChangeEmitter = new TypedEvent();
        this.buffsChangeEmitter = new TypedEvent();
        // Emits when anything in the party changes.
        this.changeEmitter = new TypedEvent();
        this.sim = sim;
        this.raid = raid;
        this.players = [...Array(MAX_PARTY_SIZE).keys()].map(i => null);
        this.playerChangeListener = () => this.changeEmitter.emit();
        [
            this.compChangeEmitter,
            this.buffsChangeEmitter,
        ].forEach(emitter => emitter.on(() => this.changeEmitter.emit()));
    }
    size() {
        return this.players.filter(player => player != null).length;
    }
    isEmpty() {
        return this.size() == 0;
    }
    clear() {
        this.setBuffs(PartyBuffs.create());
        for (let i = 0; i < MAX_PARTY_SIZE; i++) {
            this.setPlayer(i, null);
        }
    }
    // Returns this party's index within the raid [0-4].
    getIndex() {
        return this.raid.getParties().indexOf(this);
    }
    getPlayers() {
        // Make defensive copy.
        return this.players.slice();
    }
    getPlayer(playerIndex) {
        return this.players[playerIndex];
    }
    setPlayer(playerIndex, newPlayer) {
        if (playerIndex < 0 || playerIndex >= MAX_PARTY_SIZE) {
            throw new Error('Invalid player index: ' + playerIndex);
        }
        if (newPlayer == this.players[playerIndex]) {
            return;
        }
        if (this.players[playerIndex] != null) {
            this.players[playerIndex].changeEmitter.off(this.playerChangeListener);
        }
        if (newPlayer != null) {
            newPlayer.changeEmitter.on(this.playerChangeListener);
            newPlayer.setParty(this);
        }
        this.players[playerIndex] = newPlayer;
        this.compChangeEmitter.emit();
    }
    getBuffs() {
        // Make a defensive copy
        return PartyBuffs.clone(this.buffs);
    }
    setBuffs(newBuffs) {
        if (PartyBuffs.equals(this.buffs, newBuffs))
            return;
        // Make a defensive copy
        this.buffs = PartyBuffs.clone(newBuffs);
        this.buffsChangeEmitter.emit();
    }
    toProto() {
        return PartyProto.create({
            players: this.players.filter(player => player != null).map(player => player.toProto()),
            buffs: this.buffs,
        });
    }
    // Returns JSON representing all the current values.
    toJson() {
        return {
            'players': this.players.map(player => {
                if (player == null) {
                    return null;
                }
                else {
                    return {
                        'spec': player.spec,
                        'player': player.toJson(),
                    };
                }
            }),
            'buffs': PartyBuffs.toJson(this.getBuffs()),
        };
    }
    // Set all the current values, assumes obj is the same type returned by toJson().
    fromJson(obj) {
        try {
            this.setBuffs(PartyBuffs.fromJson(obj['buffs']));
        }
        catch (e) {
            console.warn('Failed to parse party buffs: ' + e);
        }
        if (obj['players']) {
            for (let i = 0; i < MAX_PARTY_SIZE; i++) {
                const playerObj = obj['players'][i];
                if (!playerObj) {
                    this.setPlayer(i, null);
                    continue;
                }
                const newSpec = playerObj['spec'];
                if (this.players[i] != null && this.players[i].spec == newSpec) {
                    this.players[i].fromJson(playerObj['player']);
                }
                else {
                    const newPlayer = new Player(playerObj['spec'], this.sim);
                    newPlayer.fromJson(playerObj['player']);
                    this.setPlayer(i, newPlayer);
                }
            }
        }
    }
}
