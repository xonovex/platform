# testing-patterns: Testing Patterns

**Guideline:** Typed assertions, epsilon for floats, parity tests for variants.

**Rationale:** Catches type mismatches. Epsilon handles float imprecision. Parity ensures variants produce identical results.

**Example:**

```c
void test_vec3_dot(void) {
    vec3_t a = {1, 2, 3, 0}, b = {4, 5, 6, 0};
    assert_float_eq(vec3_dot(&a, &b), 32.0f);  // 1*4+2*5+3*6
}

void test_vec3_dot_soa_parity(void) {
    float ax[100], ay[100], az[100];
    float bx[100], by[100], bz[100];
    float scalar_out[100], soa_out[100];

    for (size_t i = 0; i < 100; i++) {
        vec3_t a = {ax[i], ay[i], az[i], 0};
        vec3_t b = {bx[i], by[i], bz[i], 0};
        scalar_out[i] = vec3_dot(&a, &b);
    }

    vec3_soa_t a_soa = {ax, ay, az, 100};
    vec3_dot_soa(soa_out, &a_soa, &(vec3_soa_t){bx, by, bz, 100});

    for (size_t i = 0; i < 100; i++)
        assert_float_near(soa_out[i], scalar_out[i], 1e-5f);
}
```

**Techniques:**
- Typed assertions: Use `assert_int_eq`, `assert_float_eq`, `assert_ptr_not_null`
- Epsilon values: Use 1e-4f general, 1e-5f cumulative, 5e-4f for SIMD
- Scalar tests: Validate against known values for correctness
- Parity tests: Compare variant implementations against scalar reference
- Edge cases: Test zero, negative, denormal, infinity, and boundary conditions
