# string-formatting: Modern String Formatting

**Guideline:** Use f-strings for all string formatting and interpolation.

**Rationale:** F-strings are more readable, faster, and less error-prone than old `%` formatting or `.format()` methods. They support inline expressions, formatting specifiers, and multiline strings. The syntax is concise and makes variable substitution obvious.

**Example:**

```python
# f-strings (preferred)
name = "Alice"
age = 30
message = f"Hello, {name}! You are {age} years old."

# f-strings with expressions
result = f"The answer is {2 + 2}"

# f-strings with formatting
pi = 3.14159
formatted = f"Pi is approximately {pi:.2f}"

# Multiline f-strings
user = User(id="123", email="user@example.com", name="Alice", created_at=datetime.now())
summary = f"""
User Information:
  ID: {user.id}
  Email: {user.email}
  Name: {user.name}
"""

# ❌ Avoid old-style formatting
message_old = "Hello, %s! You are %d years old." % (name, age)
message_format = "Hello, {}! You are {} years old.".format(name, age)
```

**Techniques:**
- Use f-string syntax: `f"text {variable} more text"`
- Embed expressions directly: `f"result: {2 + 2}"`
- Apply format specifiers: `f"{value:.2f}"` for decimals
- Use multiline f-strings with triple quotes for structured text
- Avoid old `%` formatting and `.format()` methods
