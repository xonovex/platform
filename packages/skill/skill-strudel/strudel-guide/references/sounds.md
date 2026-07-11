# sounds: Available Sounds and Instruments

Not every sample name loads in the web version; use these verified names or fall back to synthesis (always available).

- **Verified drums**: `bd` `sd` `hh` `oh` `cp` `rim` `lt` `mt` `ht`
- **Drum banks** (via `.bank()`): `RolandTR909` `RolandTR808` `RolandTR707` `AkaiLinn` `ViscoSpaceDrum`
- **Waveforms** (always available): `sine` `sawtooth` `square` `triangle`
- **Verified melodic samples**: `piano` `epiano` `casio` `gm_acoustic_bass` `gm_synth_bass_1`
- **Fallback**: replace an unreliable sample with a waveform shaped by `attack`/`sustain`/`release`/`lpf`/`room`.

```javascript
s("bd sd hh").bank("RolandTR808");
note("c3").sound("sawtooth").lpf(400); // bass fallback
note("c3").sound("sine").attack(0.5).sustain(2).release(1).room(0.8).gain(0.4); // pad fallback
```
