use moon_nix_runtime::guard::{decide_wrap, WrapFacts, WrapGuard};

#[test]
fn nix_shell_reentry_keeps_the_task_unchanged() {
    let facts = WrapFacts {
        in_nix_shell: true,
        already_wrapped: false,
        nix_available: true,
    };

    let decision = decide_wrap(facts);

    assert_eq!(decision, WrapGuard::Unchanged);
}

#[test]
fn sentinel_reentry_keeps_the_task_unchanged() {
    let facts = WrapFacts {
        in_nix_shell: false,
        already_wrapped: true,
        nix_available: true,
    };

    let decision = decide_wrap(facts);

    assert_eq!(decision, WrapGuard::Unchanged);
}

#[test]
fn absent_nix_defers_to_the_adapter_policy() {
    let facts = WrapFacts {
        in_nix_shell: false,
        already_wrapped: false,
        nix_available: false,
    };

    let decision = decide_wrap(facts);

    assert_eq!(decision, WrapGuard::MissingNix);
}

#[test]
fn available_nix_allows_target_resolution() {
    let facts = WrapFacts {
        in_nix_shell: false,
        already_wrapped: false,
        nix_available: true,
    };

    let decision = decide_wrap(facts);

    assert_eq!(decision, WrapGuard::Ready);
}
