use moon_pdk_test_utils::*;
use serial_test::serial;

const SENTINEL: &str = "MOON_NIX_WRAPPED";

/// Reset the env so the next call takes the wrapping path: outside any dev shell
/// and not already wrapped. Tests that need `nix` present rely on running inside
/// the flake (`nix develop --command cargo test`).
fn reset_wrap_env() {
    std::env::remove_var("IN_NIX_SHELL");
    std::env::remove_var(SENTINEL);
}

fn command_input(command: &str, args: &[&str]) -> ExtendTaskCommandInput {
    ExtendTaskCommandInput {
        command: command.into(),
        args: args.iter().map(|a| (*a).to_string()).collect(),
        ..Default::default()
    }
}

/// Extract the flake reference (the `nix develop <flakeref>` argument) from a
/// wrapped command task.
fn flake_ref(output: &ExtendTaskCommandOutput) -> String {
    let Some(Extend::Replace(args)) = &output.args else {
        panic!("expected args to be replaced, got {:?}", output.args);
    };
    assert_eq!(args[0], "develop");
    args[1].clone()
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn wraps_command_task_in_nix_develop() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_command(command_input("clang-format", &["--version"]))
        .await;

    assert_eq!(output.command.as_deref(), Some("nix"));

    let Some(Extend::Replace(args)) = output.args else {
        panic!("expected args to be replaced, got {:?}", output.args);
    };
    assert_eq!(args[0], "develop");
    assert!(!args[1].is_empty(), "workspace root should be resolved");
    assert_eq!(args[2], "--command");
    assert_eq!(args[3], "clang-format");
    assert_eq!(args[4], "--version");

    assert_eq!(output.env.get(SENTINEL).map(String::as_str), Some("1"));
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn no_op_when_already_in_nix_shell() {
    reset_wrap_env();
    std::env::set_var("IN_NIX_SHELL", "impure");

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_command(command_input("echo", &["hi"]))
        .await;

    std::env::remove_var("IN_NIX_SHELL");

    assert_eq!(
        output.command, None,
        "task must run unchanged inside a dev shell"
    );
    assert!(output.args.is_none());
    assert!(output.env.is_empty());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn no_op_when_sentinel_set() {
    reset_wrap_env();
    std::env::set_var(SENTINEL, "1");

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_command(command_input("echo", &["hi"]))
        .await;

    std::env::remove_var(SENTINEL);

    assert_eq!(
        output.command, None,
        "an already-wrapped task must not be re-wrapped"
    );
    assert!(output.args.is_none());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn passthrough_when_nix_absent() {
    use std::os::unix::fs::PermissionsExt;

    reset_wrap_env();

    // Simulate a host without nix: a PATH whose only `which` reports every
    // command as missing (exit 1), mirroring a real `which nix` miss without
    // trapping on a missing `which` itself.
    let bin_dir = std::env::temp_dir().join("moon-nix-toolchain-no-nix");
    std::fs::create_dir_all(&bin_dir).unwrap();
    let which = bin_dir.join("which");
    std::fs::write(&which, "#!/bin/sh\nexit 1\n").unwrap();
    std::fs::set_permissions(&which, std::fs::Permissions::from_mode(0o755)).unwrap();

    let original_path = std::env::var_os("PATH");
    std::env::set_var("PATH", &bin_dir);

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_command(command_input("echo", &["hi"]))
        .await;

    match original_path {
        Some(path) => std::env::set_var("PATH", path),
        None => std::env::remove_var("PATH"),
    }

    assert_eq!(
        output.command, None,
        "must never hard-fail when nix is absent"
    );
    assert!(output.args.is_none());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn wraps_script_task_via_bash() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_script(ExtendTaskScriptInput {
            script: "echo hi && ls".into(),
            ..Default::default()
        })
        .await;

    let script = output.script.expect("script should be wrapped");
    assert!(script.starts_with("nix develop "), "got: {script}");
    assert!(script.contains("--command bash -c "), "got: {script}");
    assert!(script.contains("echo hi && ls"), "got: {script}");
    assert_eq!(output.env.get(SENTINEL).map(String::as_str), Some("1"));
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn wraps_in_named_dev_shell_from_project_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.toolchain_config = serde_json::json!({ "shell": "go" });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#go"),
        "flakeref should select the go devShell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn selects_dev_shell_per_task_from_shell_by_task() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.task.target = serde_json::from_value(serde_json::json!("script-lib-go:go-lint")).unwrap();
    // A per-task `shellByTask` entry wins over the project-wide `shell`.
    input.toolchain_config = serde_json::json!({
        "shellByTask": { "go-lint": "go" },
        "shell": "default"
    });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#go"),
        "shellByTask should select the go devShell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn selects_dev_shell_from_shell_by_toolchain() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.task.toolchains = serde_json::from_value(serde_json::json!(["nix", "go"])).unwrap();
    input.toolchain_config = serde_json::json!({ "shellByToolchain": { "go": "go" } });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#go"),
        "shellByToolchain should select the go devShell from task.toolchains, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn selects_dev_shell_from_shell_by_language() {
    reset_wrap_env();

    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|_id| serde_json::json!({ "language": "bash" }));
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("shellcheck", &["-x"]);
    input.toolchain_config = serde_json::json!({ "shellByLanguage": { "bash": "shell" } });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#shell"),
        "shellByLanguage should select the shell devShell from project.language, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn selects_dev_shell_from_shell_by_tag() {
    reset_wrap_env();

    let mut sandbox = create_empty_moon_sandbox();
    sandbox.host_funcs.mock_load_project(
        |_id| serde_json::json!({ "config": { "tags": ["tenant-shared", "kubernetes"] } }),
    );
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("kustomize", &["build"]);
    input.toolchain_config = serde_json::json!({ "shellByTag": { "kubernetes": "k8s" } });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#k8s"),
        "shellByTag should select the k8s devShell from a project tag, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn shell_by_task_outranks_shell_by_toolchain() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.task.target = serde_json::from_value(serde_json::json!("p:go-lint")).unwrap();
    input.task.toolchains = serde_json::from_value(serde_json::json!(["nix", "go"])).unwrap();
    input.toolchain_config = serde_json::json!({
        "shellByTask": { "go-lint": "go" },
        "shellByToolchain": { "go": "wrong" }
    });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#go"),
        "shellByTask must outrank shellByToolchain, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn shell_by_toolchain_outranks_tag_and_language() {
    reset_wrap_env();

    let mut sandbox = create_empty_moon_sandbox();
    // The project would resolve a different shell by tag/language; the task
    // toolchain must win, and the project must not even need loading.
    sandbox.host_funcs.mock_load_project(
        |_id| serde_json::json!({ "language": "wrong", "config": { "tags": ["wrong"] } }),
    );
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.task.toolchains = serde_json::from_value(serde_json::json!(["nix", "go"])).unwrap();
    input.toolchain_config = serde_json::json!({
        "shellByToolchain": { "go": "go" },
        "shellByTag": { "wrong": "wrong" },
        "shellByLanguage": { "wrong": "wrong" }
    });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#go"),
        "shellByToolchain must outrank shellByTag/shellByLanguage, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn shell_by_tag_outranks_shell_by_language() {
    reset_wrap_env();

    let mut sandbox = create_empty_moon_sandbox();
    sandbox.host_funcs.mock_load_project(
        |_id| serde_json::json!({ "language": "go", "config": { "tags": ["shell"] } }),
    );
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("task", &[]);
    input.toolchain_config = serde_json::json!({
        "shellByTag": { "shell": "shell" },
        "shellByLanguage": { "go": "go" }
    });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#shell"),
        "shellByTag must outrank shellByLanguage, got: {}",
        flake_ref(&output)
    );
}
