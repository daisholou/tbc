import { Enchant } from '/tbc/core/proto/common.js';
import { Encounter as EncounterProto } from '/tbc/core/proto/common.js';
import { EquipmentSpec } from '/tbc/core/proto/common.js';
import { Gem } from '/tbc/core/proto/common.js';
import { GemColor } from '/tbc/core/proto/common.js';
import { ItemSlot } from '/tbc/core/proto/common.js';
import { ItemSpec } from '/tbc/core/proto/common.js';
import { Item } from '/tbc/core/proto/common.js';
import { Spec } from '/tbc/core/proto/common.js';
import { Stat } from '/tbc/core/proto/common.js';
import { Raid as RaidProto } from '/tbc/core/proto/api.js';
import { RaidSimRequest, RaidSimResult } from '/tbc/core/proto/api.js';
import { StatWeightsRequest, StatWeightsResult } from '/tbc/core/proto/api.js';
import { EquippedItem } from '/tbc/core/proto_utils/equipped_item.js';
import { Gear } from '/tbc/core/proto_utils/gear.js';
import { SimResult } from '/tbc/core/proto_utils/sim_result.js';
import { Encounter } from './encounter.js';
import { Player } from './player.js';
import { Raid } from './raid.js';
import { EventID, TypedEvent } from './typed_event.js';
export declare type RaidSimData = {
    request: RaidSimRequest;
    result: RaidSimResult;
};
export declare type StatWeightsData = {
    request: StatWeightsRequest;
    result: StatWeightsResult;
};
export declare class Sim {
    private readonly workerPool;
    private iterations;
    private phase;
    readonly raid: Raid;
    readonly encounter: Encounter;
    private items;
    private enchants;
    private gems;
    readonly iterationsChangeEmitter: TypedEvent<void>;
    readonly phaseChangeEmitter: TypedEvent<void>;
    readonly changeEmitter: TypedEvent<void>;
    readonly simResultEmitter: TypedEvent<SimResult>;
    private readonly _initPromise;
    private modifyRaidProto;
    private modifyEncounterProto;
    constructor();
    waitForInit(): Promise<void>;
    setModifyRaidProto(newModFn: (raidProto: RaidProto) => void): void;
    setModifyEncounterProto(newModFn: (encounterProto: EncounterProto) => void): void;
    private makeRaidSimRequest;
    runRaidSim(eventID: EventID): Promise<SimResult>;
    runRaidSimWithLogs(eventID: EventID): Promise<SimResult>;
    private updateCharacterStats;
    statWeights(player: Player<any>, epStats: Array<Stat>, epReferenceStat: Stat): Promise<StatWeightsResult>;
    getItems(slot: ItemSlot | undefined): Array<Item>;
    getEnchants(slot: ItemSlot | undefined): Array<Enchant>;
    getGems(socketColor: GemColor | undefined): Array<Gem>;
    getMatchingGems(socketColor: GemColor): Array<Gem>;
    getPhase(): number;
    setPhase(eventID: EventID, newPhase: number): void;
    getIterations(): number;
    setIterations(eventID: EventID, newIterations: number): void;
    lookupItemSpec(itemSpec: ItemSpec): EquippedItem | null;
    lookupEquipmentSpec(equipSpec: EquipmentSpec): Gear;
    toJson(): Object;
    fromJson(eventID: EventID, obj: any, spec?: Spec): void;
}
