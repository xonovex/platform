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

/// Simulate a host without `nix`: point PATH at a dir whose only `which` reports
/// every command missing (exit 1), mirroring a real `which nix` miss without
/// trapping on a missing `which` itself. Returns a closure that restores the prior
/// PATH; call it before asserting so PATH never leaks across `#[serial]` tests.
fn stub_missing_nix() -> impl FnOnce() {
    use std::os::unix::fs::PermissionsExt;

    let bin_dir = std::env::temp_dir().join("moon-nix-toolchain-no-nix");
    std::fs::create_dir_all(&bin_dir).unwrap();
    let which = bin_dir.join("which");
    std::fs::write(&which, "#!/bin/sh\nexit 1\n").unwrap();
    std::fs::set_permissions(&which, std::fs::Permissions::from_mode(0o755)).unwrap();

    let original_path = std::env::var_os("PATH");
    std::env::set_var("PATH", &bin_dir);

    move || match original_path {
        Some(path) => std::env::set_var("PATH", path),
        None => std::env::remove_var("PATH"),
    }
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
    reset_wrap_env();
    // Canonical not-opted case: no `failClosed*` config, so a nix-absent host stays
    // a silent no-op rather than failing closed.
    let restore = stub_missing_nix();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_command(command_input("echo", &["hi"]))
        .await;

    restore();

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

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn routes_project_flake_to_named_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    // A project that ships its own flake.nix exposing a named `go` devShell.
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(
        sandbox.root.join("packages/proj/flake.nix"),
        "{ outputs = _: { devShells.x86_64-linux = { default = {}; go = {}; }; }; }",
    )
    .unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();
    // A selector now routes a project flake to one of ITS OWN named devShells.
    input.toolchain_config = serde_json::json!({ "shell": "go" });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("/packages/proj#go"),
        "a project flake with a selector should route to its named devShell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn project_flake_without_selector_uses_default_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();

    let output = plugin.extend_task_command(input).await;

    let flake_ref = flake_ref(&output);
    assert!(
        flake_ref.ends_with("/packages/proj"),
        "project flake root should be used, got: {flake_ref}"
    );
    assert!(
        !flake_ref.contains('#'),
        "no selector should keep the project flake's default devShell, got: {flake_ref}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn project_flake_falls_back_to_default_when_named_shell_absent() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    // A project flake that exposes only `default`, not the selected `go`.
    std::fs::write(
        sandbox.root.join("packages/proj/flake.nix"),
        "{ outputs = _: { devShells.x86_64-linux.default = {}; }; }",
    )
    .unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();
    // The selector resolves `go`, but the project flake does not expose it, so the wrap
    // must fall back to the flake's default devShell instead of a `#go` nix would reject.
    input.toolchain_config = serde_json::json!({ "shell": "go" });

    let output = plugin.extend_task_command(input).await;

    let flake_ref = flake_ref(&output);
    assert!(
        flake_ref.ends_with("/packages/proj"),
        "an unexposed named shell must fall back to the project flake root, got: {flake_ref}"
    );
    assert!(
        !flake_ref.contains('#'),
        "fallback must drop the `#<shell>` suffix, got: {flake_ref}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn routes_project_flake_per_task_from_shell_by_task() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(
        sandbox.root.join("packages/proj/flake.nix"),
        "{ outputs = _: { devShells.x86_64-linux = { default = {}; go = {}; }; }; }",
    )
    .unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();
    input.task.target = serde_json::from_value(serde_json::json!("proj:go-lint")).unwrap();
    // shellByTask outranks the project-wide `shell`, even for a project flake.
    input.toolchain_config = serde_json::json!({
        "shellByTask": { "go-lint": "go" },
        "shell": "default"
    });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("/packages/proj#go"),
        "shellByTask should route a project flake to its named devShell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn define_toolchain_config_exposes_camel_case_schema() {
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    let output: DefineToolchainConfigOutput = plugin
        .plugin
        .call_func("define_toolchain_config")
        .await
        .unwrap();

    let rendered = serde_json::to_value(&output.schema).unwrap().to_string();
    for key in [
        "shell",
        "shellByTask",
        "shellByToolchain",
        "shellByTag",
        "shellByLanguage",
    ] {
        assert!(
            rendered.contains(key),
            "schema should expose `{key}`: {rendered}"
        );
    }
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn typed_config_defaults_to_flake_default_shell() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    // No toolchain config: parse yields struct defaults, so no `#shell` suffix.
    let output = plugin
        .extend_task_command(command_input("cargo", &["build"]))
        .await;

    assert!(
        !flake_ref(&output).contains('#'),
        "an unset config must select the default devShell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn uses_workspace_flake_when_project_has_no_flake() {
    reset_wrap_env();

    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_toolchain("nix").await;

    // A project source without a flake.nix falls back to the workspace flake, and
    // the shell selector still applies.
    let mut input = command_input("golangci-lint", &["run"]);
    input.project =
        serde_json::from_value(serde_json::json!({ "id": "proj", "source": "packages/proj" }))
            .unwrap();
    input.toolchain_config = serde_json::json!({ "shell": "go" });

    let output = plugin.extend_task_command(input).await;

    assert!(
        flake_ref(&output).ends_with("#go"),
        "no project flake should fall back to the workspace flake + shell, got: {}",
        flake_ref(&output)
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn fails_closed_when_opted_in_and_nix_absent() {
    reset_wrap_env();
    let restore = stub_missing_nix();

    let mut sandbox = create_empty_moon_sandbox();
    // A `cmake`-tagged project opted into fail-closed nix via `failClosedByTag`.
    sandbox
        .host_funcs
        .mock_load_project(|_id| serde_json::json!({ "config": { "tags": ["cmake"] } }));
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("cmake", &["--build", "."]);
    input.context = plugin.create_context();
    input.project = serde_json::from_value(
        serde_json::json!({ "id": "game-worldgen", "source": "packages/game/game-worldgen" }),
    )
    .unwrap();
    input.task.target = serde_json::from_value(serde_json::json!("game-worldgen:build")).unwrap();
    input.toolchain_config = serde_json::json!({ "failClosedByTag": ["cmake"] });

    // Call the plugin container directly: the wrapper `.unwrap()`s the plugin `Err`,
    // so this returns a `Result` and PATH is restored before asserting.
    let result: Result<ExtendTaskCommandOutput, _> = plugin
        .plugin
        .call_func_with("extend_task_command", input)
        .await;

    restore();
    let err = result.expect_err("opted-in task must fail when nix is absent");
    let message = format!("{err:?}");
    assert!(message.contains("nix is required"), "got: {message}");
    assert!(
        message.contains("game-worldgen:build"),
        "error must name <project>:<task>, got: {message}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn no_op_when_not_opted_in_and_nix_absent() {
    reset_wrap_env();
    let restore = stub_missing_nix();

    let mut sandbox = create_empty_moon_sandbox();
    // Project is tagged `go`, but only `cmake` is fail-closed -> stay a no-op.
    sandbox
        .host_funcs
        .mock_load_project(|_id| serde_json::json!({ "config": { "tags": ["go"] } }));
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("go", &["build"]);
    input.toolchain_config = serde_json::json!({ "failClosedByTag": ["cmake"] });
    let output = plugin.extend_task_command(input).await;

    restore();
    assert_eq!(
        output.command, None,
        "non-opted task must keep the host-tool no-op"
    );
    assert!(output.args.is_none());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn in_nix_shell_outranks_fail_closed() {
    reset_wrap_env();
    // Already inside a dev shell: the `IN_NIX_SHELL` guard wins above the nix probe,
    // so an opted-in, nix-absent task is a no-op rather than a fail-closed error.
    std::env::set_var("IN_NIX_SHELL", "impure");
    let restore = stub_missing_nix();

    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|_id| serde_json::json!({ "config": { "tags": ["cmake"] } }));
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("cmake", &["--build", "."]);
    input.toolchain_config = serde_json::json!({ "failClosedByTag": ["cmake"] });
    let output = plugin.extend_task_command(input).await;

    restore();
    std::env::remove_var("IN_NIX_SHELL");

    assert_eq!(
        output.command, None,
        "IN_NIX_SHELL must win over fail-closed: no wrap and no error"
    );
    assert!(output.args.is_none());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn fails_closed_when_opted_in_by_language_and_nix_absent() {
    reset_wrap_env();
    let restore = stub_missing_nix();

    let mut sandbox = create_empty_moon_sandbox();
    // A C project opted into fail-closed nix via `failClosedByLanguage` (the other
    // opt-in half from `failClosedByTag`).
    sandbox
        .host_funcs
        .mock_load_project(|_id| serde_json::json!({ "language": "c" }));
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("cmake", &["--build", "."]);
    input.context = plugin.create_context();
    input.project = serde_json::from_value(
        serde_json::json!({ "id": "game-worldgen", "source": "packages/game/game-worldgen" }),
    )
    .unwrap();
    input.task.target = serde_json::from_value(serde_json::json!("game-worldgen:build")).unwrap();
    input.toolchain_config = serde_json::json!({ "failClosedByLanguage": ["c"] });

    let result: Result<ExtendTaskCommandOutput, _> = plugin
        .plugin
        .call_func_with("extend_task_command", input)
        .await;

    restore();
    let err = result.expect_err("language-opted-in task must fail when nix is absent");
    let message = format!("{err:?}");
    assert!(message.contains("nix is required"), "got: {message}");
    assert!(
        message.contains("game-worldgen:build"),
        "error must name <project>:<task>, got: {message}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn sentinel_outranks_fail_closed() {
    reset_wrap_env();
    // Already wrapped: the MOON_NIX_WRAPPED guard wins above the nix probe, so an
    // opted-in, nix-absent task is a no-op rather than a fail-closed error.
    std::env::set_var(SENTINEL, "1");
    let restore = stub_missing_nix();

    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|_id| serde_json::json!({ "config": { "tags": ["cmake"] } }));
    let plugin = sandbox.create_toolchain("nix").await;

    let mut input = command_input("cmake", &["--build", "."]);
    input.toolchain_config = serde_json::json!({ "failClosedByTag": ["cmake"] });
    let output = plugin.extend_task_command(input).await;

    restore();
    std::env::remove_var(SENTINEL);

    assert_eq!(
        output.command, None,
        "MOON_NIX_WRAPPED must win over fail-closed: no wrap and no error"
    );
    assert!(output.args.is_none());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn no_host_load_when_allowlists_empty() {
    reset_wrap_env();
    let restore = stub_missing_nix();

    let mut sandbox = create_empty_moon_sandbox();
    // With no `failClosed*` configured (the common case for every current consumer),
    // the fail-closed check must short-circuit before any host project load — a load
    // here panics the host mock, failing the test.
    sandbox
        .host_funcs
        .mock_load_project(|_id| panic!("must not load project when both allowlists are empty"));
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .extend_task_command(command_input("echo", &["hi"]))
        .await;

    restore();
    assert_eq!(
        output.command, None,
        "non-opted nix-absent task must stay a no-op without loading the project"
    );
    assert!(output.args.is_none());
}
