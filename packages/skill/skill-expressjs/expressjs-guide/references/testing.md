# testing: Supertest API Testing (Express glue)

## Guideline

Test API endpoints using supertest with realistic request/response scenarios.

## Rationale

Integration tests validate the full middleware chain and response format.

Test runner setup, `describe`/`it` organization, assertions, mocking, and HTTP status conventions belong to **vitest-guide**. This file covers only the Express glue: driving the app with supertest.

## Example

```typescript
import request from "supertest";
import {app} from "../src/app";

describe("POST /api/users", () => {
  it("should create user with valid data", async () => {
    const res = await request(app).post("/api/users").send({
      email: "test@example.com",
      name: "Test User",
      password: "securepassword",
    });
    expect(res.status).toBe(201);
    expect(res.body.data).toHaveProperty("id");
  });

  it("should return 400 for invalid email", async () => {
    const res = await request(app).post("/api/users").send({email: "invalid"});
    expect(res.status).toBe(400);
    expect(res.body).toHaveProperty("error");
  });
});
```

## Techniques (Express glue — runner/assertions live in vitest-guide)

- supertest: Make HTTP requests to the Express app without opening a network port
- request(app).get/post/patch/delete(): Chain HTTP verb with URL path
- .send(): Pass request body as JSON
- Integration testing: Drive the full middleware chain, validation, and handlers end to end
- Realistic scenarios: Exercise valid/invalid inputs, edge cases, and error responses
