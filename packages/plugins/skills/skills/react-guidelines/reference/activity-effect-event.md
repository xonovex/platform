# activity-effect-event: Activity Component & useEffectEvent

**Guideline:** Use `<Activity mode="hidden">` to preserve state; use `useEffectEvent` for Effect logic that reads latest values without re-triggering.

**Rationale:** Activity preserves DOM state during hide (tabs, accordions); useEffectEvent eliminates false re-runs caused by dependency changes.

**Example:**
```tsx
// Activity - preserve state when hidden
function TabContainer({ tabs, activeTab }) {
  return tabs.map(tab => (
    <Activity
      key={tab.id}
      mode={tab.id === activeTab ? 'visible' : 'hidden'}
    >
      <TabContent tab={tab} /> {/* State preserved even when hidden */}
    </Activity>
  ));
}

// useEffectEvent - read theme without reconnecting
function ChatRoom({ roomId, theme }) {
  const onConnected = useEffectEvent(() => {
    showNotification(`Connected to ${roomId}`, theme);
  });
  useEffect(() => {
    const connection = createConnection(roomId);
    connection.on('connected', onConnected);
    connection.connect();
    return () => connection.disconnect();
  }, [roomId]); // theme NOT here - doesn't reconnect
}
```

**Techniques:**
- Activity visible/hidden: Preserves state of hidden components (vs conditional render which destroys state)
- Activity pre-rendering: Load content in background for faster transition when shown
- useEffectEvent: Read latest prop/state without adding to dependencies; reads current value
- Solves false re-runs: Chat reconnect on theme change; analytics re-fires on items change
- Media handling: Activity doesn't pause video/audio automatically; use useLayoutEffect to pause
- Deprecating useLayoutEffect: useEffectEvent replaces some useLayoutEffect patterns in React 19

