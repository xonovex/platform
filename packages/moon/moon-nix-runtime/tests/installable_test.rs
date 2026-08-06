use moon_nix_runtime::installable::{
    canonical_installable, moon_shell_environment, parse_workspace_installable, InstallableScheme,
    MOON_COMPONENTS_ENV, MOON_FLAKE_ENV, MOON_SHELL_EXPRESSION,
};
use std::path::{Path, PathBuf};

#[test]
fn workspace_installable_requires_a_relative_path_and_named_shell() {
    let parsed = parse_workspace_installable("path:./packages/demo#go").unwrap();

    assert_eq!(parsed.scheme, InstallableScheme::Copy);
    assert_eq!(parsed.relative_flake, PathBuf::from("./packages/demo"));
    assert_eq!(parsed.dev_shell, "go");
}

#[test]
fn workspace_installable_accepts_a_tree_reference() {
    let parsed = parse_workspace_installable("dir:./packages/demo#default").unwrap();

    assert_eq!(parsed.scheme, InstallableScheme::Tree);
    assert_eq!(parsed.relative_flake, PathBuf::from("./packages/demo"));
    assert_eq!(parsed.dev_shell, "default");
}

#[test]
fn workspace_installable_rejects_remote_absolute_and_missing_shell_forms() {
    for value in [
        "github:owner/repo#go",
        "path:/outside#go",
        "path:./project",
        "path:./project#",
        "path:./project#go#extra",
        "dir:/outside#go",
        "dir:./project",
        "dir:./project#",
    ] {
        assert!(
            parse_workspace_installable(value).is_err(),
            "`{value}` should be rejected"
        );
    }
}

#[test]
fn canonical_installable_keeps_workspace_paths_with_spaces() {
    let installable = canonical_installable(
        Path::new("/workspace with spaces"),
        Path::new("/workspace with spaces/packages/demo"),
        "node",
        InstallableScheme::Copy,
    )
    .unwrap();

    assert_eq!(
        installable,
        "path:/workspace with spaces/packages/demo#node"
    );
}

#[test]
fn canonical_tree_installable_is_a_bare_directory_reference() {
    let installable = canonical_installable(
        Path::new("/workspace"),
        Path::new("/workspace/packages/demo"),
        "default",
        InstallableScheme::Tree,
    )
    .unwrap();

    assert_eq!(installable, "/workspace/packages/demo#default");
}

#[test]
fn canonical_installable_rejects_a_symlink_escape_fact() {
    let error = canonical_installable(
        Path::new("/workspace"),
        Path::new("/outside/project"),
        "go",
        InstallableScheme::Copy,
    )
    .unwrap_err();

    assert_eq!(
        error.to_string(),
        "environment flake `/outside/project` resolves outside workspace `/workspace`"
    );
}

#[test]
fn moon_shell_environment_serializes_validated_flake_and_component_values() {
    let environment = moon_shell_environment(
        Path::new("/workspace with spaces"),
        &["general".to_owned(), "node".to_owned()],
    )
    .unwrap();

    assert_eq!(
        environment.get(MOON_FLAKE_ENV).map(String::as_str),
        Some("path:/workspace with spaces")
    );
    assert_eq!(
        environment.get(MOON_COMPONENTS_ENV).map(String::as_str),
        Some("[\"general\",\"node\"]")
    );
    assert!(!MOON_SHELL_EXPRESSION.contains("/workspace"));
    assert!(!MOON_SHELL_EXPRESSION.contains("general"));
}

#[test]
fn moon_shell_environment_allows_an_explicit_empty_component_set() {
    let environment = moon_shell_environment(Path::new("/workspace"), &[]).unwrap();

    assert_eq!(
        environment.get(MOON_COMPONENTS_ENV).map(String::as_str),
        Some("[]")
    );
}
