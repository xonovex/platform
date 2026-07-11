# testing: Supertest API Testing (Express glue)

Drive the app with `request(app)` — no network port needed — to exercise the full middleware chain end to end. Runner setup, `describe`/`it`, assertions, and mocking belong to **vitest-guide**; this file covers only the supertest glue.

```typescript
import request from "supertest";
import {app} from "../src/app";

describe("POST /api/users", () => {
  it("creates a user with valid data", async () => {
    const res = await request(app).post("/api/users").send({
      email: "test@example.com",
      name: "Test User",
      password: "securepassword",
    });
    expect(res.status).toBe(201);
    expect(res.body.data).toHaveProperty("id");
  });

  it("returns 400 for invalid email", async () => {
    const res = await request(app).post("/api/users").send({email: "invalid"});
    expect(res.status).toBe(400);
    expect(res.body).toHaveProperty("error");
  });
});
```
