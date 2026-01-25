# dom-measurement: Measuring DOM Elements

**Guideline:** Divide `getBoundingClientRect()` values by `useCurrentScale()` to account for Remotion's scale transforms.

**Rationale:** Remotion applies scale transforms to elements; measurements need correction to reflect actual dimensions at scale 1.

**Example:**

```tsx
// Without scale correction (wrong)
const rect = ref.current.getBoundingClientRect();
const width = rect.width;  // Scaled value

// With scale correction (correct)
const scale = useCurrentScale();
const width = rect.width / scale;  // Actual unscaled width

// Full example
const scale = useCurrentScale();
const {width: scaled_width, height: scaled_height} = ref.current.getBoundingClientRect();
return <div>Actual size: {scaled_width / scale} x {scaled_height / scale}</div>;
```

**Correct Measurement:**

```tsx
import {useCurrentScale} from "remotion";
import {useEffect, useRef, useState} from "react";

export function MeasuredComponent() {
  const ref = useRef<HTMLDivElement>(null);
  const scale = useCurrentScale();
  const [dimensions, setDimensions] = useState({width: 0, height: 0});

  useEffect(() => {
    if (!ref.current) return;

    const rect = ref.current.getBoundingClientRect();
    setDimensions({
      width: rect.width / scale,
      height: rect.height / scale,
    });
  }, [scale]);

  return (
    <div ref={ref}>
      Content to measure ({dimensions.width}x{dimensions.height})
    </div>
  );
}
```

**Outline vs Border:** Outline doesn't affect layout; border does at different scales.

```tsx
function PositionedElement() {
  const ref = useRef<HTMLDivElement>(null);
  const scale = useCurrentScale();
  const [position, setPosition] = useState({x: 0, y: 0});

  useEffect(() => {
    if (!ref.current) return;

    const rect = ref.current.getBoundingClientRect();
    setPosition({
      x: rect.left / scale,
      y: rect.top / scale,
    });
  }, [scale]);

  return <div ref={ref}>Element at ({position.x}, {position.y})</div>;
}
```

**Techniques:**
- useCurrentScale(): Get scale factor applied by Remotion
- Divide measurements: width/scale, height/scale for correct values
- useEffect hook: Measure elements after mount with scale dependency
- outline vs border: Use outline to avoid layout shifts at different scales
- getBoundingClientRect(): Get left, top, width, height of element
- useRef: Store reference to DOM element for measurement
- Sync with scale: Update measurements when scale changes
