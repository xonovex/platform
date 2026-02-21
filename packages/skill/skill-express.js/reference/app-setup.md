# app-setup: Basic Application Setup

**Guideline:** Configure Express app with security, logging, body parsing, routes, and error handling in correct order.

**Rationale:** Middleware order matters - security first, error handler last ensures proper request processing.

**Example:**

```typescript
const app = express();

// Security middleware (first)
app.use(helmet());
app.use(
  cors({origin: process.env.ALLOWED_ORIGINS?.split(","), credentials: true}),
);

// Logging and parsing
app.use(morgan("combined"));
app.use(express.json());
app.use(express.urlencoded({extended: true}));

// Routes
app.use("/api/users", userRoutes);
app.use("/api/auth", authRoutes);

// 404 handler
app.use((req, res) => res.status(404).json({error: "Not found"}));

// Error handler (last)
app.use(errorHandler);
```

**Techniques:**

- Middleware order: Security → logging → parsing → routes → 404 → error handler
- helmet(): Add security headers (CSP, HSTS, X-Frame-Options, etc)
- cors(): Configure allowed origins and credentials for cross-origin requests
- morgan(): Log HTTP requests in combined format
- express.json()/urlencoded(): Parse request bodies
- 404 handler: Catch unmatched routes before error handler
- Error handler: Must be last, receives 4 parameters (err, req, res, next)
