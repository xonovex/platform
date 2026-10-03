use moon_nix_runtime::target::{resolve_flake_target, FlakeTarget};

#[test]
fn project_flake_outranks_the_workspace_flake() {
    let target = resolve_flake_target(
        Some("/workspace".to_owned()),
        Some("/workspace/project".to_owned()),
    );

    assert_eq!(
        target,
        Some(FlakeTarget {
            root: "/workspace/project".to_owned(),
            is_project_flake: true,
        })
    );
}

#[test]
fn workspace_flake_is_the_fallback_target() {
    let target = resolve_flake_target(Some("/workspace".to_owned()), None);

    assert_eq!(
        target,
        Some(FlakeTarget {
            root: "/workspace".to_owned(),
            is_project_flake: false,
        })
    );
}

#[test]
fn unresolved_roots_leave_the_task_unchanged() {
    let target = resolve_flake_target(None, None);

    assert_eq!(target, None);
}
