# server-components: React Server Components & Actions

**Guideline:** Use Server Components by default for data fetching and static UI; add `'use client'` only for interactivity.

**Rationale:** Server Components send zero JS to the client, access databases/APIs directly, and eliminate loading state boilerplate. Client boundaries should be minimal.

**Example:**

```tsx
import {useActionState} from "react";
import {createPost} from "./actions";

// Server Component (default) - no directive needed
async function ProductPage({params}: {params: {id: string}}) {
  const product = await db.products.find(params.id);
  const reviews = await db.reviews.findMany({productId: params.id});

  return (
    <main>
      <h1>{product.name}</h1>
      <p>{product.description}</p>
      <AddToCartButton productId={product.id} /> {/* Client island */}
      <ReviewList reviews={reviews} />
    </main>
  );
}

// Client Component - needs directive for interactivity
("use client");

function AddToCartButton({productId}: {productId: string}) {
  const [isPending, startTransition] = useTransition();

  return (
    <button
      onClick={() => startTransition(() => addToCart(productId))}
      disabled={isPending}>
      {isPending ? "Adding..." : "Add to Cart"}
    </button>
  );
}

// Server Action - separate file recommended
// actions.ts
("use server");

export async function createPost(prevState: any, formData: FormData) {
  const title = formData.get("title") as string;
  if (!title) return {error: "Title required"};

  await db.posts.create({title});
  revalidatePath("/posts");
  return {success: true};
}

export async function deletePost(id: string) {
  await db.posts.delete(id);
  revalidatePath("/posts");
}

// Using Server Actions in Client Component
("use client");

function NewPostForm() {
  const [state, formAction, isPending] = useActionState(createPost, null);

  return (
    <form action={formAction}>
      <input name="title" disabled={isPending} />
      <button disabled={isPending}>Create</button>
      {state?.error && <span>{state.error}</span>}
    </form>
  );
}

// Inline Server Action in Server Component
async function QuickForm() {
  async function handleSubmit(formData: FormData) {
    "use server";
    await db.items.create({name: formData.get("name")});
    revalidatePath("/items");
  }

  return (
    <form action={handleSubmit}>
      <input name="name" />
      <button>Add</button>
    </form>
  );
}
```

**Techniques:**

- No directive = Server Component (default); can be `async`, access db/fs directly
- Add `'use client'` only for useState, useEffect, onClick, browser APIs
- Use `'use server'` for Server Actions (mutations callable from client)
- Keep client boundaries small - wrap only interactive parts
- Pass Server Components as children to Client Components when needed
