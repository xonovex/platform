# app-setup: Basic Application Setup

Register middleware in this order: security (`helmet`, `cors`) → logging (`morgan`) → parsing (`express.json`/`urlencoded`) → routes → 404 catch-all → error handler. The error handler must be last, and the 404 must precede it.

```typescript
const app = express();

app.use(helmet());
app.use(
  cors({origin: process.env.ALLOWED_ORIGINS?.split(","), credentials: true}),
);
app.use(morgan("combined"));
app.use(express.json());
app.use(express.urlencoded({extended: true}));

app.use("/api/users", userRoutes);
app.use("/api/auth", authRoutes);

app.use((req, res) => res.status(404).json({error: "Not found"})); // 404 catch-all
app.use(errorHandler); // last; must take (err, req, res, next)
```
