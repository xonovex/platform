# state-management: State Management Best Practices

Start with local `useState`; lift state only when siblings must share it; use `useReducer` for complex state logic; derive computed values instead of storing them.

## Example

```tsx
// useReducer for complex state - typed discriminated-union actions
type State = {count: number; step: number};
type Action =
  {type: "increment"} | {type: "decrement"} | {type: "setStep"; step: number};

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

## Techniques

- Start with local `useState` in the component that owns the data
- Lift to the parent when sibling components need the state
- Use `useReducer` for complex state logic instead of multiple `useState`
- Use `useMemo` to derive values from existing state
- Consider Context for deeply nested prop passing
