/// Environment variable set on a wrapped task so it is wrapped at most once.
pub const SENTINEL: &str = "MOON_NIX_WRAPPED";
/// Environment variable that identifies the exact Nix environment selected for a task.
pub const SHELL_IDENTITY: &str = "MOON_NIX_SHELL_ID";

/// Host observations that determine whether an adapter should wrap a task.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct WrapFacts {
    pub in_nix_shell: bool,
    pub already_wrapped: bool,
    pub nix_available: bool,
}

/// The adapter action implied by the current host observations.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub enum WrapGuard {
    Unchanged,
    MissingNix,
    Ready,
}

/// Host observations for adapters that can identify their required environment exactly.
#[derive(Clone, Copy, Debug, Eq, PartialEq)]
pub struct IdentityWrapFacts<'a> {
    pub current_identity: Option<&'a str>,
    pub required_identity: &'a str,
    pub nix_available: bool,
}

pub const fn decide_wrap(facts: WrapFacts) -> WrapGuard {
    if facts.in_nix_shell || facts.already_wrapped {
        WrapGuard::Unchanged
    } else if facts.nix_available {
        WrapGuard::Ready
    } else {
        WrapGuard::MissingNix
    }
}

pub fn decide_identity_wrap(facts: IdentityWrapFacts<'_>) -> WrapGuard {
    if facts.current_identity == Some(facts.required_identity) {
        WrapGuard::Unchanged
    } else if facts.nix_available {
        WrapGuard::Ready
    } else {
        WrapGuard::MissingNix
    }
}
