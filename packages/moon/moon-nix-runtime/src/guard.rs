/// Environment variable set on a wrapped task so it is wrapped at most once.
pub const SENTINEL: &str = "MOON_NIX_WRAPPED";

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

pub const fn decide_wrap(facts: WrapFacts) -> WrapGuard {
    if facts.in_nix_shell || facts.already_wrapped {
        WrapGuard::Unchanged
    } else if facts.nix_available {
        WrapGuard::Ready
    } else {
        WrapGuard::MissingNix
    }
}
