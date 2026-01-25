# sounds: Available Sounds and Instruments

**Guideline:** Use verified drum samples and synthesis waveforms; fall back to synthesis for unreliable samples.

**Rationale:** Not all sample names work in the web version. Verified sounds prevent silent patterns.

**Example:**

```javascript
// Use verified drum sounds
s("bd sd hh ~ hh")  // Always works in web version

// Switch drum banks
s("bd sd hh").bank("RolandTR808")  // TR-808 drums

// Synthesis waveforms (always available)
note("c3").sound("sine")      // Pure tone
note("c3").sound("sawtooth")  // Bright sound

// Combination: Drums + melody
s("bd*4, hh*8").gain(0.7)
note("c3 e3 g3 c4").sound("sine").sustain(0.5)

// If melodic sample fails, use synthesis
note("c2").sound("sawtooth").lpf(400)  // Fallback bass
```

## Verified Drum Samples

```javascript
s("bd")    // Kick drum
s("sd")    // Snare
s("hh")    // Hi-hat (closed)
s("oh")    // Hi-hat (open)
s("cp")    // Clap
s("rim")   // Rimshot
s("lt")    // Low tom
s("mt")    // Mid tom
s("ht")    // High tom
```

## Drum Banks

```javascript
s("bd sd hh").bank("RolandTR909")   // TR-909
s("bd sd hh").bank("RolandTR808")   // TR-808
// Also: RolandTR707, AkaiLinn, ViscoSpaceDrum
```

## Synthesis Waveforms (Always Available)

```javascript
note("c3").sound("sine")      // Pure, smooth
note("c3").sound("sawtooth")  // Bright, rich
note("c3").sound("square")    // Hollow, vintage
note("c3").sound("triangle")  // Warm, soft
```

## Verified Melodic Samples

```javascript
note("c3").sound("piano")
note("c3").sound("epiano")
note("c3").sound("casio")
note("c2").sound("gm_acoustic_bass")
note("c2").sound("gm_synth_bass_1")
```

## Synthesis Alternative (for unreliable samples)

```javascript
// Instead of pad sample:
note("c3").sound("sine")
  .attack(0.5).sustain(2).release(1)
  .room(0.8).gain(0.4)

// Instead of bass sample:
note("c2").sound("sawtooth").lpf(400).gain(0.6)
```

**Techniques:**
- Verified drums: bd, sd, hh, oh, cp, rim, lt, mt, ht always work
- Drum banks: RolandTR909, RolandTR808, RolandTR707, AkaiLinn, ViscoSpaceDrum
- Synthesis waveforms: sine, sawtooth, square, triangle always available
- Melodic samples: piano, epiano, casio, gm_acoustic_bass, gm_synth_bass_1
- Fallback strategy: Use synthesis (sine/sawtooth) if samples fail
- Envelope control: attack(), sustain(), release() for shaping
- Effects: room(), lpf() (low-pass filter), gain() for tone shaping
- Test in browser: Verify sounds work in web version before performance
