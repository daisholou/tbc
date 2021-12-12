import { Encounter } from '/tbc/core/encounter.js';
import { Raid } from '/tbc/core/raid.js';
import { Sim } from '/tbc/core/sim.js';
import { DetailedResults } from '/tbc/core/components/detailed_results.js';
import { LogRunner } from '/tbc/core/components/log_runner.js';
import { SavedDataConfig } from '/tbc/core/components/saved_data_manager.js';
import { SavedDataManager } from '/tbc/core/components/saved_data_manager.js';

import { RaidPicker, PresetSpecSettings } from './raid_picker.js';

declare var tippy: any;

export interface RaidSimConfig {
	knownIssues?: Array<string>,
	presets: Array<PresetSpecSettings<any>>,
}

export class RaidSimUI {
  readonly sim: Sim;

  private readonly config: RaidSimConfig;
  private readonly parentElem: HTMLElement;

  constructor(parentElem: HTMLElement, config: RaidSimConfig) {
    this.sim = new Sim();

    this.config = config;
		this.parentElem = parentElem;
    this.parentElem.innerHTML = layoutHTML;

		const titleElem = this.parentElem.getElementsByClassName('default-title')[0];
		titleElem.textContent = 'TBC Raid Sim';

		const raidPicker = new RaidPicker(this.parentElem.getElementsByClassName('raid-picker')[0] as HTMLElement, this.sim.raid, this.config.presets);
  }
}

const layoutHTML = `
<div class="default-root">
  <section class="default-sidebar">
    <div class="default-title"></div>
    <div class="default-actions"></div>
    <div class="default-results"></div>
    <div class="default-buffs"></div>
  </section>
  <section class="default-main">
    <ul class="nav nav-tabs">
      <li class="active"><a data-toggle="tab" href="#raid-tab">Raid</a></li>
      <li><a data-toggle="tab" href="#detailed-results-tab">Detailed Results</a></li>
      <li><a data-toggle="tab" href="#log-tab">Log</a></li>
      <li class="default-top-bar">
				<div class="known-issues">Known Issues</div>
				<span class="share-link fa fa-link"></span>
			</li>
    </ul>
    <div class="tab-content">
      <div id="raid-tab" class="raid-tab tab-pane fade in active">
				<div class="raid-picker">
				</div>
				<div class="saved-raids-div">
					<div class="saved-raids-manager">
					</div>
				</div>
      </div>
      <div id="detailed-results-tab" class="tab-pane fade">
				<div class="detailed-results">
				</div>
      </div>
      <div id="log-tab" class="tab-pane fade">
				<div class="log-runner">
				</div>
      </div>
    </div>
  </section>
</div>
`;
