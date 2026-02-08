# accessibility: Accessibility Best Practices

**Guideline:** Use semantic HTML, ARIA labels/roles, proper focus management, and keyboard navigation for inclusive components.

**Rationale:** Accessible design works for all users (screen readers, keyboards, assistive tech); legal requirement in many jurisdictions.

**Example:**

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

**Techniques:**

- Semantic HTML: `<button>`, `<nav>`, `<main>`, `<article>` instead of divs
- ARIA attributes: aria-label, aria-labelledby, aria-describedby, role="dialog", role="list"
- Focus management: useRef + useEffect to move focus to interactive elements
- Visually hidden text: sr-only class for screen readers ("sr-only" Tailwind utility)
- Icon buttons: Always include aria-label; never icon-only without text
- Keyboard navigation: All interactive elements must be focusable via Tab
