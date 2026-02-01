# new-hooks: React 19 Hooks

**Guideline:** Use React 19's new hooks for forms, optimistic updates, and promise/context reading.

**Rationale:** These hooks replace manual state management patterns, reducing boilerplate and providing built-in pending states, error handling, and progressive enhancement.

**Example:**

```tsx
import {Suspense, use, useActionState, useOptimistic} from "react";
import {useFormStatus} from "react-dom";

// useActionState - form state with async actions
function ContactForm() {
  const [state, formAction, isPending] = useActionState(
    async (prevState, formData) => {
      const result = await submitForm({
        email: formData.get("email"),
        message: formData.get("message"),
      });
      if (result.error) return {error: result.error};
      return {success: true};
    },
    null,
  );

  return (
    <form action={formAction}>
      <input name="email" type="email" disabled={isPending} />
      <textarea name="message" disabled={isPending} />
      <SubmitButton />
      {state?.error && <p className="error">{state.error}</p>}
    </form>
  );
}

// useFormStatus - reads parent form state (must be inside form)
function SubmitButton() {
  const {pending} = useFormStatus();
  return (
    <button type="submit" disabled={pending}>
      {pending ? "Sending..." : "Send"}
    </button>
  );
}

// useOptimistic - instant feedback, reverts on error
function LikeButton({postId, likes}: {postId: string; likes: number}) {
  const [optimisticLikes, addOptimistic] = useOptimistic(
    likes,
    (current, delta: number) => current + delta,
  );

  async function handleLike() {
    addOptimistic(1);
    await likePost(postId); // reverts if throws
  }

  return (
    <form action={handleLike}>
      <button type="submit">{optimisticLikes} likes</button>
    </form>
  );
}

// use() - read promises (suspends until resolved)
function Comments({commentsPromise}: {commentsPromise: Promise<Comment[]>}) {
  const comments = use(commentsPromise);
  return comments.map((c) => <p key={c.id}>{c.text}</p>);
}

// use() - conditional context (not possible with useContext)
function ThemedContent({showTheme}: {showTheme: boolean}) {
  if (!showTheme) return <div>Plain</div>;
  const theme = use(ThemeContext); // can be called conditionally!
  return <div className={theme}>Themed</div>;
}

// Parent creates promise, child reads it
function Page({postId}: {postId: string}) {
  const commentsPromise = fetchComments(postId);
  return (
    <Suspense fallback={<Skeleton />}>
      <Comments commentsPromise={commentsPromise} />
    </Suspense>
  );
}
```

**Techniques:**

- Use `useActionState` for form state with async actions (replaces manual loading/error state)
- Use `useOptimistic` for instant UI feedback that reverts on error
- Use `use()` to read promises (suspends) or context (can be conditional)
- Use `useFormStatus` inside forms to read parent form's pending state
