# accessibility: Accessibility Best Practices

**accessibility-guide** owns applicable criteria, assurance evidence, exceptions, and conformance claims. This reference owns only the React implementation delta: semantic JSX, component labels, and hook-based focus management.

## Example

```tsx
function Modal({isOpen, onClose, title, children}) {
  const closeButtonRef = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (isOpen) closeButtonRef.current?.focus();
  }, [isOpen]);
  if (!isOpen) return null;
  return (
    <div role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <h2 id="modal-title">{title}</h2>
      {children}
      <button ref={closeButtonRef} onClick={onClose} aria-label="Close modal">
        Close
      </button>
    </div>
  );
}
```

## Techniques

- Semantic HTML: `<button>`, `<nav>`, `<main>`, `<article>` instead of divs
- ARIA: aria-label, aria-labelledby, aria-describedby, role="dialog", role="list"
- Focus management: useRef + useEffect to move focus to interactive elements
- Visually hidden text: `sr-only` Tailwind utility for screen readers
- Icon buttons: always include aria-label; never icon-only without text
- Keyboard navigation: all interactive elements focusable via Tab
