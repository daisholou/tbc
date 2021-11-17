import type { BinaryWriteOptions } from "@protobuf-ts/runtime";
import type { IBinaryWriter } from "@protobuf-ts/runtime";
import type { BinaryReadOptions } from "@protobuf-ts/runtime";
import type { IBinaryReader } from "@protobuf-ts/runtime";
import type { PartialMessage } from "@protobuf-ts/runtime";
import { MessageType } from "@protobuf-ts/runtime";
/**
 * @generated from protobuf message proto.ShamanTalents
 */
export interface ShamanTalents {
    /**
     * Elemental
     *
     * @generated from protobuf field: int32 convection = 1;
     */
    convection: number;
    /**
     * @generated from protobuf field: int32 concussion = 2;
     */
    concussion: number;
    /**
     * @generated from protobuf field: int32 call_of_flame = 3;
     */
    callOfFlame: number;
    /**
     * @generated from protobuf field: bool elemental_focus = 4;
     */
    elementalFocus: boolean;
    /**
     * @generated from protobuf field: int32 reverberation = 5;
     */
    reverberation: number;
    /**
     * @generated from protobuf field: int32 call_of_thunder = 6;
     */
    callOfThunder: number;
    /**
     * @generated from protobuf field: int32 improved_fire_totems = 7;
     */
    improvedFireTotems: number;
    /**
     * @generated from protobuf field: int32 elemental_devastation = 8;
     */
    elementalDevastation: number;
    /**
     * @generated from protobuf field: bool elemental_fury = 9;
     */
    elementalFury: boolean;
    /**
     * @generated from protobuf field: int32 unrelenting_storm = 10;
     */
    unrelentingStorm: number;
    /**
     * @generated from protobuf field: int32 elemental_precision = 11;
     */
    elementalPrecision: number;
    /**
     * @generated from protobuf field: int32 lightning_mastery = 12;
     */
    lightningMastery: number;
    /**
     * @generated from protobuf field: bool elemental_mastery = 13;
     */
    elementalMastery: boolean;
    /**
     * @generated from protobuf field: int32 lightning_overload = 14;
     */
    lightningOverload: number;
    /**
     * @generated from protobuf field: bool totemOfWrath = 33;
     */
    totemOfWrath: boolean;
    /**
     * Enhancement
     *
     * @generated from protobuf field: int32 ancestral_knowledge = 15;
     */
    ancestralKnowledge: number;
    /**
     * @generated from protobuf field: int32 thundering_strikes = 16;
     */
    thunderingStrikes: number;
    /**
     * @generated from protobuf field: int32 enhancing_totems = 17;
     */
    enhancingTotems: number;
    /**
     * @generated from protobuf field: bool shamanistic_focus = 18;
     */
    shamanisticFocus: boolean;
    /**
     * @generated from protobuf field: int32 flurry = 19;
     */
    flurry: number;
    /**
     * @generated from protobuf field: int32 improved_weapon_totems = 20;
     */
    improvedWeaponTotems: number;
    /**
     * @generated from protobuf field: int32 elemental_weapons = 21;
     */
    elementalWeapons: number;
    /**
     * @generated from protobuf field: int32 mental_quickness = 22;
     */
    mentalQuickness: number;
    /**
     * @generated from protobuf field: int32 weapon_mastery = 23;
     */
    weaponMastery: number;
    /**
     * @generated from protobuf field: int32 dual_wield_specialization = 24;
     */
    dualWieldSpecialization: number;
    /**
     * @generated from protobuf field: int32 unleashed_rage = 25;
     */
    unleashedRage: number;
    /**
     * Restoration
     *
     * @generated from protobuf field: int32 totemic_focus = 26;
     */
    totemicFocus: number;
    /**
     * @generated from protobuf field: int32 natures_guidance = 27;
     */
    naturesGuidance: number;
    /**
     * @generated from protobuf field: int32 restorative_totems = 28;
     */
    restorativeTotems: number;
    /**
     * @generated from protobuf field: int32 tidal_mastery = 29;
     */
    tidalMastery: number;
    /**
     * @generated from protobuf field: bool natures_swiftness = 30;
     */
    naturesSwiftness: boolean;
    /**
     * @generated from protobuf field: bool mana_tide_totem = 31;
     */
    manaTideTotem: boolean;
    /**
     * @generated from protobuf field: int32 natures_blessing = 32;
     */
    naturesBlessing: number;
}
/**
 * @generated from protobuf message proto.ElementalShaman
 */
export interface ElementalShaman {
    /**
     * @generated from protobuf field: proto.ElementalShaman.Rotation rotation = 1;
     */
    rotation?: ElementalShaman_Rotation;
    /**
     * @generated from protobuf field: proto.ShamanTalents talents = 2;
     */
    talents?: ShamanTalents;
    /**
     * @generated from protobuf field: proto.ElementalShaman.Options options = 3;
     */
    options?: ElementalShaman_Options;
}
/**
 * @generated from protobuf message proto.ElementalShaman.Rotation
 */
export interface ElementalShaman_Rotation {
    /**
     * @generated from protobuf field: proto.ElementalShaman.Rotation.RotationType type = 1;
     */
    type: ElementalShaman_Rotation_RotationType;
    /**
     * Only used if type == FixedLBCL
     *
     * @generated from protobuf field: int32 lbs_per_cl = 2;
     */
    lbsPerCl: number;
}
/**
 * @generated from protobuf enum proto.ElementalShaman.Rotation.RotationType
 */
export declare enum ElementalShaman_Rotation_RotationType {
    /**
     * @generated from protobuf enum value: Unknown = 0;
     */
    Unknown = 0,
    /**
     * @generated from protobuf enum value: Adaptive = 1;
     */
    Adaptive = 1,
    /**
     * @generated from protobuf enum value: CLOnClearcast = 2;
     */
    CLOnClearcast = 2,
    /**
     * @generated from protobuf enum value: CLOnCD = 3;
     */
    CLOnCD = 3,
    /**
     * @generated from protobuf enum value: FixedLBCL = 4;
     */
    FixedLBCL = 4,
    /**
     * @generated from protobuf enum value: LBOnly = 5;
     */
    LBOnly = 5
}
/**
 * @generated from protobuf message proto.ElementalShaman.Options
 */
export interface ElementalShaman_Options {
    /**
     * @generated from protobuf field: bool water_shield = 1;
     */
    waterShield: boolean;
    /**
     * @generated from protobuf field: bool bloodlust = 2;
     */
    bloodlust: boolean;
    /**
     * @generated from protobuf field: bool mana_spring_totem = 3;
     */
    manaSpringTotem: boolean;
    /**
     * @generated from protobuf field: bool totem_of_wrath = 4;
     */
    totemOfWrath: boolean;
    /**
     * @generated from protobuf field: bool wrath_of_air_totem = 5;
     */
    wrathOfAirTotem: boolean;
}
declare class ShamanTalents$Type extends MessageType<ShamanTalents> {
    constructor();
    create(value?: PartialMessage<ShamanTalents>): ShamanTalents;
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: ShamanTalents): ShamanTalents;
    internalBinaryWrite(message: ShamanTalents, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter;
}
/**
 * @generated MessageType for protobuf message proto.ShamanTalents
 */
export declare const ShamanTalents: ShamanTalents$Type;
declare class ElementalShaman$Type extends MessageType<ElementalShaman> {
    constructor();
    create(value?: PartialMessage<ElementalShaman>): ElementalShaman;
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: ElementalShaman): ElementalShaman;
    internalBinaryWrite(message: ElementalShaman, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter;
}
/**
 * @generated MessageType for protobuf message proto.ElementalShaman
 */
export declare const ElementalShaman: ElementalShaman$Type;
declare class ElementalShaman_Rotation$Type extends MessageType<ElementalShaman_Rotation> {
    constructor();
    create(value?: PartialMessage<ElementalShaman_Rotation>): ElementalShaman_Rotation;
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: ElementalShaman_Rotation): ElementalShaman_Rotation;
    internalBinaryWrite(message: ElementalShaman_Rotation, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter;
}
/**
 * @generated MessageType for protobuf message proto.ElementalShaman.Rotation
 */
export declare const ElementalShaman_Rotation: ElementalShaman_Rotation$Type;
declare class ElementalShaman_Options$Type extends MessageType<ElementalShaman_Options> {
    constructor();
    create(value?: PartialMessage<ElementalShaman_Options>): ElementalShaman_Options;
    internalBinaryRead(reader: IBinaryReader, length: number, options: BinaryReadOptions, target?: ElementalShaman_Options): ElementalShaman_Options;
    internalBinaryWrite(message: ElementalShaman_Options, writer: IBinaryWriter, options: BinaryWriteOptions): IBinaryWriter;
}
/**
 * @generated MessageType for protobuf message proto.ElementalShaman.Options
 */
export declare const ElementalShaman_Options: ElementalShaman_Options$Type;
export {};
