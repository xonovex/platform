# testing: API Testing Pattern

**Guideline:** Test API endpoints using supertest with realistic request/response scenarios.

**Rationale:** Integration tests validate the full middleware chain and response format.

**Example:**

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

**Techniques:**
- supertest: Make HTTP requests to Express app without network
- request(app).get/post/patch/delete(): Chain HTTP verb with URL path
- .send(): Pass request body as JSON
- expect(res.status).toBe(201): Assert HTTP status code
- expect(res.body.data): Assert response structure
- Integration testing: Validate full middleware chain, validation, handlers
- Realistic scenarios: Test valid/invalid inputs, edge cases, error responses
