# activity-effect-event: Activity Component & useEffectEvent

Use `<Activity mode="hidden">` to preserve DOM/state when hidden; use `useEffectEvent` for Effect logic that reads latest values without adding them to deps.

## Example

```tsx
// Activity - preserve state when hidden
function TabContainer({tabs, activeTab}) {
  return tabs.map((tab) => (
    <Activity key={tab.id} mode={tab.id === activeTab ? "visible" : "hidden"}>
      <TabContent tab={tab} /> {/* State preserved even when hidden */}
    </Activity>
  ));
}

// useEffectEvent - read theme without reconnecting
function ChatRoom({roomId, theme}) {
  const onConnected = useEffectEvent(() => {
    showNotification(`Connected to ${roomId}`, theme);
  });
  useEffect(() => {
    const connection = createConnection(roomId);
    connection.on("connected", onConnected);
    connection.connect();
    return () => connection.disconnect();
  }, [roomId]); // theme NOT here - doesn't reconnect
}
```

## Techniques

- Activity visible/hidden: preserves state of hidden components (vs conditional render which destroys it)
- Activity pre-rendering: loads hidden content in background for faster transition when shown
- useEffectEvent: reads current prop/state without adding to deps; solves false re-runs (reconnect on theme change, analytics re-fire on items change)
- Activity does NOT pause video/audio automatically; use useLayoutEffect to pause
