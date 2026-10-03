use moon_nix_runtime::guard::SENTINEL;
use moon_nix_runtime::installable::{MOON_COMPONENTS_ENV, MOON_FLAKE_ENV, MOON_SHELL_EXPRESSION};
use moon_nix_runtime::serialize::quote_posix;
use moon_pdk_test_utils::*;
use serial_test::serial;

fn reset_wrap_env() {
    std::env::remove_var("IN_NIX_SHELL");
    std::env::remove_var(SENTINEL);
}

/// Simulate a host without `nix`: point PATH at a dir whose only `which` reports
/// every command missing (exit 1). Returns a closure that restores the prior PATH
/// and removes the stub dir, so neither leaks across `#[serial]` tests.
///
/// The directory name carries the pid and a per-call counter because `#[serial]`
/// orders tests within one binary only: two concurrent `cargo test` runs share
/// $TMPDIR, and a fixed name would race on the write and the chmod.
fn stub_missing_nix() -> impl FnOnce() {
    use std::os::unix::fs::PermissionsExt;
    use std::sync::atomic::{AtomicU32, Ordering};

    static CALL: AtomicU32 = AtomicU32::new(0);

    let bin_dir = std::env::temp_dir().join(format!(
        "moon-nix-extension-no-nix-{}-{}",
        std::process::id(),
        CALL.fetch_add(1, Ordering::Relaxed)
    ));
    std::fs::create_dir_all(&bin_dir).unwrap();
    let which = bin_dir.join("which");
    std::fs::write(&which, "#!/bin/sh\nexit 1\n").unwrap();
    std::fs::set_permissions(&which, std::fs::Permissions::from_mode(0o755)).unwrap();

    let original_path = std::env::var_os("PATH");
    std::env::set_var("PATH", &bin_dir);

    move || {
        match original_path {
            Some(path) => std::env::set_var("PATH", path),
            None => std::env::remove_var("PATH"),
        }
        std::fs::remove_dir_all(&bin_dir).unwrap();
    }
}

fn command_input(command: &str, args: &[&str], toolchains: &[&str]) -> ExtendTaskCommandInput {
    ExtendTaskCommandInput {
        command: command.into(),
        args: args.iter().map(|argument| (*argument).to_owned()).collect(),
        project: serde_json::from_value(serde_json::json!({
            "id": "project",
            "source": "project"
        }))
        .unwrap(),
        task: serde_json::from_value(serde_json::json!({
            "target": "project:task",
            "toolchains": toolchains
        }))
        .unwrap(),
        ..Default::default()
    }
}

fn replaced_args(output: &ExtendTaskCommandOutput) -> &[String] {
    let Some(Extend::Replace(args)) = &output.args else {
        panic!("expected args to be replaced, got {:?}", output.args);
    };
    args
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn register_extension_preserves_public_metadata() {
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;

    assert_eq!(plugin.metadata.name, "Nix environment");
    assert_eq!(plugin.metadata.plugin_version, "0.3.0");
    assert_eq!(
        plugin.metadata.description.as_deref(),
        Some("Composes native Moon task toolchains through the workspace Nix flake.")
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn define_extension_config_exposes_camel_case_schema() {
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;

    let output: DefineExtensionConfigOutput = plugin
        .plugin
        .call_func("define_extension_config")
        .await
        .unwrap();
    let rendered = serde_json::to_value(output.schema).unwrap().to_string();

    for key in [
        "baseComponents",
        "environmentByToolchain",
        "environmentByProject",
        "environmentByTask",
        "failClosed",
        "components",
        "mode",
        "installable",
    ] {
        assert!(
            rendered.contains(key),
            "schema should expose `{key}`: {rendered}"
        );
    }
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn detected_node_task_resolves_general_and_node_without_system() {
    reset_wrap_env();
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;

    let output = plugin
        .extend_task_command(command_input(
            "node",
            &["script with spaces.js", "${literal}"],
            &["javascript", "npm", "node"],
        ))
        .await;
    let args = replaced_args(&output);

    assert_eq!(output.command.as_deref(), Some("nix"));
    assert_eq!(args[0], "develop");
    assert_eq!(args[1], "--impure");
    assert_eq!(&args[2..5], ["--option", "eval-cache", "false"]);
    assert_eq!(args[5], "--no-update-lock-file");
    assert_eq!(args[6], "--expr");
    assert_eq!(args[7], MOON_SHELL_EXPRESSION);
    assert_eq!(args[8], "moon");
    assert_eq!(args[9], "--command");
    assert_eq!(args[10], "node");
    assert_eq!(args[11], "script with spaces.js");
    assert_eq!(args[12], "${literal}");
    assert_eq!(output.env.get(SENTINEL).map(String::as_str), Some("1"));
    assert_eq!(
        output.env.get(MOON_COMPONENTS_ENV).map(String::as_str),
        Some("[\"general\",\"node\"]")
    );
    assert!(output
        .env
        .get(MOON_FLAKE_ENV)
        .is_some_and(|flake| flake.starts_with("path:")));
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn unmapped_task_is_unchanged_when_base_components_are_present() {
    reset_wrap_env();
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;

    let mut input = command_input("custom", &[], &["custom"]);
    input.extension_config = serde_json::json!({
        "baseComponents": ["general"],
        "environmentByToolchain": {}
    });
    let output = plugin.extend_task_command(input).await;

    assert_eq!(output.command, None);
    assert!(output.args.is_none());
    assert!(output.env.is_empty());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn script_task_remains_one_opaque_quoted_bash_argument() {
    reset_wrap_env();
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;
    let script = r#"printf '%s\n' "$HOME ${literal} \\ path""#;
    let input = ExtendTaskScriptInput {
        project: serde_json::from_value(serde_json::json!({
            "id": "project",
            "source": "project"
        }))
        .unwrap(),
        task: serde_json::from_value(serde_json::json!({
            "target": "project:task",
            "toolchains": ["go"]
        }))
        .unwrap(),
        script: script.to_owned(),
        ..Default::default()
    };

    let output = plugin.extend_task_script(input).await;
    let wrapped = output.script.unwrap();

    assert!(wrapped.starts_with("'nix' 'develop' '--impure' '--option' 'eval-cache' 'false' '--no-update-lock-file' '--expr' "));
    assert!(wrapped.contains(&quote_posix(MOON_SHELL_EXPRESSION)));
    assert!(wrapped.contains("'moon' '--command' 'bash' '-c'"));
    assert!(wrapped.ends_with(&quote_posix(script)), "got: {wrapped}");
    assert_eq!(output.env.get(SENTINEL).map(String::as_str), Some("1"));
    assert_eq!(
        output.env.get(MOON_COMPONENTS_ENV).map(String::as_str),
        Some("[\"general\",\"go\"]")
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn project_then_task_override_uses_full_target_and_semantic_host_validation() {
    reset_wrap_env();
    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|id| serde_json::json!({ "id": id }));
    sandbox
        .host_funcs
        .mock_load_task(|target| serde_json::json!({ "target": target }));
    let plugin = sandbox.create_extension("nix-environment").await;

    let mut input = command_input("cargo", &["check"], &["custom"]);
    input.extension_config = serde_json::json!({
        "environmentByToolchain": {},
        "environmentByProject": {
            "project": { "components": ["go"], "mode": "append" }
        },
        "environmentByTask": {
            "project:task": { "components": ["rust"], "mode": "replace" }
        }
    });
    let output = plugin.extend_task_command(input).await;
    let args = replaced_args(&output);

    assert_eq!(args[7], MOON_SHELL_EXPRESSION);
    assert_eq!(
        output.env.get(MOON_COMPONENTS_ENV).map(String::as_str),
        Some("[\"rust\"]"),
        "task replace should outrank project append"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn workspace_installable_is_canonicalized_and_bypasses_the_expression() {
    reset_wrap_env();
    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|id| serde_json::json!({ "id": id }));
    let flake_root = sandbox.root.join("project flake");
    std::fs::create_dir_all(&flake_root).unwrap();
    std::fs::write(flake_root.join("flake.nix"), "{}").unwrap();
    std::fs::write(flake_root.join("flake.lock"), "{}").unwrap();
    let plugin = sandbox.create_extension("nix-environment").await;

    let mut input = command_input("go", &["test", "./..."], &["custom"]);
    input.extension_config = serde_json::json!({
        "environmentByToolchain": {},
        "environmentByProject": {
            "project": { "installable": "path:./project flake#moon" }
        }
    });
    let output = plugin.extend_task_command(input).await;
    let args = replaced_args(&output);

    assert_eq!(args[0], "develop");
    assert_eq!(args[1], "--impure");
    assert_eq!(&args[2..5], ["--option", "eval-cache", "false"]);
    assert_eq!(args[5], "--no-update-lock-file");
    assert_eq!(
        args[6],
        format!("path:{}#moon", flake_root.canonicalize().unwrap().display())
    );
    assert_eq!(args[7], "--command");
    assert!(!args.iter().any(|argument| argument == "--expr"));
    assert!(!output.env.contains_key(MOON_FLAKE_ENV));
    assert!(!output.env.contains_key(MOON_COMPONENTS_ENV));
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn workspace_tree_installable_is_a_bare_directory_reference() {
    reset_wrap_env();
    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|id| serde_json::json!({ "id": id }));
    let flake_root = sandbox.root.join("composing-flake");
    std::fs::create_dir_all(&flake_root).unwrap();
    std::fs::write(flake_root.join("flake.nix"), "{}").unwrap();
    std::fs::write(flake_root.join("flake.lock"), "{}").unwrap();
    let plugin = sandbox.create_extension("nix-environment").await;

    let mut input = command_input("go", &["test", "./..."], &["custom"]);
    input.extension_config = serde_json::json!({
        "environmentByToolchain": {},
        "environmentByProject": {
            "project": { "installable": "dir:./composing-flake#default" }
        }
    });
    let output = plugin.extend_task_command(input).await;
    let args = replaced_args(&output);

    assert_eq!(
        args[6],
        format!(
            "{}#default",
            flake_root.canonicalize().unwrap().display()
        ),
        "a tree installable must carry no `path:` prefix so nix resolves it through the enclosing git tree"
    );
    assert_eq!(args[7], "--command");
    assert!(!args.iter().any(|argument| argument == "--expr"));
    assert!(!output.env.contains_key(MOON_FLAKE_ENV));
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn workspace_installable_requires_a_lock_file() {
    reset_wrap_env();
    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|id| serde_json::json!({ "id": id }));
    let flake_root = sandbox.root.join("unlocked-flake");
    std::fs::create_dir_all(&flake_root).unwrap();
    std::fs::write(flake_root.join("flake.nix"), "{}").unwrap();
    let plugin = sandbox.create_extension("nix-environment").await;

    let mut input = command_input("go", &["test"], &["custom"]);
    input.context = plugin.create_context();
    input.extension_config = serde_json::json!({
        "environmentByToolchain": {},
        "environmentByProject": {
            "project": { "installable": "path:./unlocked-flake#moon" }
        }
    });
    let result: Result<ExtendTaskCommandOutput, _> = plugin
        .plugin
        .call_func_with("extend_task_command", input)
        .await;

    let message = format!("{:?}", result.unwrap_err());
    assert!(message.contains("has no flake.lock"), "got: {message}");
}

#[cfg(unix)]
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn workspace_installable_rejects_a_symlink_escape() {
    use std::os::unix::fs::symlink;

    reset_wrap_env();
    let mut sandbox = create_empty_moon_sandbox();
    sandbox
        .host_funcs
        .mock_load_project(|id| serde_json::json!({ "id": id }));
    let outside = sandbox.root.parent().unwrap().join("outside-flake");
    std::fs::create_dir_all(&outside).unwrap();
    std::fs::write(outside.join("flake.nix"), "{}").unwrap();
    symlink(&outside, sandbox.root.join("escape")).unwrap();
    let plugin = sandbox.create_extension("nix-environment").await;

    let mut input = command_input("go", &["test"], &["custom"]);
    input.context = plugin.create_context();
    input.extension_config = serde_json::json!({
        "environmentByToolchain": {},
        "environmentByProject": {
            "project": { "installable": "path:./escape#moon" }
        }
    });
    let result: Result<ExtendTaskCommandOutput, _> = plugin
        .plugin
        .call_func_with("extend_task_command", input)
        .await;

    let message = format!("{:?}", result.unwrap_err());
    assert!(
        message.contains("resolves outside workspace"),
        "got: {message}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn malformed_full_target_key_is_rejected_before_wrapping() {
    reset_wrap_env();
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;
    let mut input = command_input("go", &["test"], &["go"]);
    input.context = plugin.create_context();
    input.extension_config = serde_json::json!({
        "environmentByTask": {
            "task": { "components": ["go"] }
        }
    });

    let result: Result<ExtendTaskCommandOutput, _> = plugin
        .plugin
        .call_func_with("extend_task_command", input)
        .await;
    let message = format!("{:?}", result.unwrap_err());

    assert!(
        message.contains("full `project:task` target"),
        "got: {message}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn missing_nix_fails_only_for_an_active_fail_closed_task() {
    reset_wrap_env();
    let restore = stub_missing_nix();
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;
    let mut input = command_input("node", &["--version"], &["node"]);
    input.context = plugin.create_context();

    let result: Result<ExtendTaskCommandOutput, _> = plugin
        .plugin
        .call_func_with("extend_task_command", input)
        .await;

    restore();
    let message = format!("{:?}", result.unwrap_err());
    assert!(message.contains("nix is required for `project:task`"));
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn missing_nix_leaves_an_active_fail_open_task_unchanged() {
    reset_wrap_env();
    let restore = stub_missing_nix();
    let sandbox = create_empty_moon_sandbox();
    let plugin = sandbox.create_extension("nix-environment").await;
    let mut input = command_input("node", &["--version"], &["node"]);
    input.extension_config = serde_json::json!({ "failClosed": false });

    let output = plugin.extend_task_command(input).await;

    restore();
    assert_eq!(output.command, None);
    assert!(output.args.is_none());
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn outer_shell_and_sentinel_guards_keep_active_tasks_unchanged() {
    for (name, value) in [("IN_NIX_SHELL", "impure"), (SENTINEL, "1")] {
        reset_wrap_env();
        std::env::set_var(name, value);
        let sandbox = create_empty_moon_sandbox();
        let plugin = sandbox.create_extension("nix-environment").await;

        let output = plugin
            .extend_task_command(command_input("node", &["--version"], &["node"]))
            .await;

        std::env::remove_var(name);
        assert_eq!(output.command, None, "guard `{name}` should no-op");
        assert!(output.args.is_none());
        assert!(output.env.is_empty());
    }
}
