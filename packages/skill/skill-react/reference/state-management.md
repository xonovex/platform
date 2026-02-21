# state-management: State Management Best Practices

**Guideline:** Use local state with `useState`, lift state only when multiple components need it, and derive computed values.

**Rationale:** Local state keeps components independent and easier to reason about. Lifting state only when necessary prevents prop drilling and over-sharing. Derived state eliminates synchronization bugs.

**Example:**

```tsx
// ✅ Local state
function Counter() {
  const [count, setCount] = useState(0);
  return <button onClick={() => setCount(count + 1)}>{count}</button>;
}

// ✅ Lifted state (when siblings need to share)
function Parent() {
  const [count, setCount] = useState(0);
  return (
    <>
      <Counter count={count} onIncrement={() => setCount(count + 1)} />
      <Display count={count} />
    </>
  );
}

// ✅ useReducer for complex state
type State = {count: number; step: number};
type Action =
  | {type: "increment"}
  | {type: "decrement"}
  | {type: "setStep"; step: number};

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "increment":
      return {...state, count: state.count + state.step};
    case "decrement":
      return {...state, count: state.count - state.step};
    case "setStep":
      return {...state, step: action.step};
    default:
      return state;
  }
}

function ComplexCounter() {
  const [state, dispatch] = useReducer(reducer, {count: 0, step: 1});

  return (
    <div>
      <p>Count: {state.count}</p>
      <button onClick={() => dispatch({type: "increment"})}>+</button>
      <button onClick={() => dispatch({type: "decrement"})}>-</button>
      <input
        type="number"
        value={state.step}
        onChange={(e) =>
          dispatch({type: "setStep", step: Number(e.target.value)})
        }
      />
    </div>
  );
}
```

**Techniques:**

- Start with local `useState` in the component that owns the data
- If multiple sibling components need the state, lift to their parent
- For complex state logic, use `useReducer` instead of multiple `useState`
- Use `useMemo` to derive values from existing state
- Consider Context API for deeply nested prop passing
