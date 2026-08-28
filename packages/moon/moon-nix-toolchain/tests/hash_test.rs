use moon_pdk_test_utils::*;
use serial_test::serial;

/// A `HashTaskContentsInput` with a fixed project id and the given source. The
/// non-empty id keeps the test harness from substituting its own project fragment,
/// so `source` alone controls project-flake-vs-workspace resolution.
fn ws_input(source: &str) -> HashTaskContentsInput {
    HashTaskContentsInput {
        project: serde_json::from_value(serde_json::json!({ "id": "proj", "source": source }))
            .unwrap(),
        ..Default::default()
    }
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn flake_lock_edit_busts_the_hash() {
    let sandbox = create_empty_moon_sandbox();
    std::fs::write(
        sandbox.root.join("flake.lock"),
        r#"{"version":7,"nodes":{}}"#,
    )
    .unwrap();
    let plugin = sandbox.create_toolchain("nix").await;

    let before = plugin.hash_task_contents(ws_input("")).await;

    std::fs::write(
        sandbox.root.join("flake.lock"),
        r#"{"version":7,"nodes":{"a":1}}"#,
    )
    .unwrap();
    let after = plugin.hash_task_contents(ws_input("")).await;

    assert_ne!(
        before.contents, after.contents,
        "editing flake.lock must change the hashed contents"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn flake_source_edit_busts_the_hash() {
    let sandbox = create_empty_moon_sandbox();
    std::fs::write(
        sandbox.root.join("flake.lock"),
        r#"{"version":7,"nodes":{}}"#,
    )
    .unwrap();
    std::fs::write(sandbox.root.join("flake.nix"), "{ outputs = _: {}; }").unwrap();
    std::fs::create_dir_all(sandbox.root.join("nix")).unwrap();
    std::fs::write(
        sandbox.root.join("nix/toolchain.nix"),
        "{ pkgs }: pkgs.hello",
    )
    .unwrap();
    let plugin = sandbox.create_toolchain("nix").await;

    let before = plugin.hash_task_contents(ws_input("")).await;

    std::fs::write(
        sandbox.root.join("nix/toolchain.nix"),
        "{ pkgs }: pkgs.ripgrep",
    )
    .unwrap();
    let after = plugin.hash_task_contents(ws_input("")).await;

    assert_ne!(
        before.contents, after.contents,
        "editing an imported Nix source must change the hashed contents"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn unrelated_edit_keeps_the_hash() {
    let sandbox = create_empty_moon_sandbox();
    std::fs::write(
        sandbox.root.join("flake.lock"),
        r#"{"version":7,"nodes":{}}"#,
    )
    .unwrap();
    let plugin = sandbox.create_toolchain("nix").await;

    let before = plugin.hash_task_contents(ws_input("")).await;

    std::fs::create_dir_all(sandbox.root.join("src")).unwrap();
    std::fs::write(sandbox.root.join("src/main.rs"), "fn main() {}").unwrap();
    let after = plugin.hash_task_contents(ws_input("")).await;

    assert_eq!(
        before.contents, after.contents,
        "an edit that does not touch flake.lock/root/shell must not change the hash"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn project_flake_lock_chosen_over_workspace_lock() {
    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();
    std::fs::write(
        sandbox.root.join("packages/proj/flake.lock"),
        r#"{"marker":"project"}"#,
    )
    .unwrap();
    std::fs::write(sandbox.root.join("flake.lock"), r#"{"marker":"workspace"}"#).unwrap();
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin.hash_task_contents(ws_input("packages/proj")).await;
    let blob = serde_json::to_string(&output.contents).unwrap();

    assert!(
        blob.contains("\\\"marker\\\":\\\"project\\\""),
        "got: {blob}"
    );
    assert!(
        !blob.contains("workspace"),
        "must not embed the workspace lock, got: {blob}"
    );
    assert!(
        blob.contains("/packages/proj"),
        "flakeRoot should be the project flake, got: {blob}"
    );
}

/// The case that had no coverage, and was wrong because of it.
///
/// A project flake composes the workspace's shared devShells through a relative
/// `path:` input. Those carry no `narHash` (Nix 2.26+ resolves them against the
/// parent source tree), so editing `nix/cc.nix` leaves the project's own
/// `flake.nix` AND its `flake.lock` byte-identical while the devShell it
/// resolves to genuinely changes. Hashing only `<projectRoot>/**` therefore
/// served cache hits built with the previous toolchain.
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn workspace_module_edit_busts_a_project_flake_hash() {
    let sandbox = create_empty_moon_sandbox();

    std::fs::write(sandbox.root.join("flake.nix"), "{ outputs = _: {}; }").unwrap();
    std::fs::write(
        sandbox.root.join("flake.lock"),
        r#"{"version":7,"nodes":{}}"#,
    )
    .unwrap();
    std::fs::create_dir_all(sandbox.root.join("nix")).unwrap();
    std::fs::write(sandbox.root.join("nix/cc.nix"), "{ pkgs }: pkgs.gcc").unwrap();

    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(
        sandbox.root.join("packages/proj/flake.nix"),
        r#"{ inputs.nixShells.url = "path:../../nix"; outputs = _: {}; }"#,
    )
    .unwrap();
    // Note the absence of a narHash on the path input — that is what the real
    // lock files look like, and the whole reason this test is necessary.
    std::fs::write(
        sandbox.root.join("packages/proj/flake.lock"),
        r#"{"version":7,"nodes":{"nixShells":{"locked":{"path":"../../nix","type":"path"}}}}"#,
    )
    .unwrap();

    let plugin = sandbox.create_toolchain("nix").await;

    let before = plugin.hash_task_contents(ws_input("packages/proj")).await;

    std::fs::write(sandbox.root.join("nix/cc.nix"), "{ pkgs }: pkgs.clang").unwrap();
    let after = plugin.hash_task_contents(ws_input("packages/proj")).await;

    assert_ne!(
        before.contents, after.contents,
        "editing a shared workspace module must bust the hash of a project that \
         composes it through a relative path: input"
    );
}

/// The shared modules must be labelled relative to the workspace root. An
/// absolute path would differ between CI and a developer's machine, so the key
/// would never match across them.
#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn workspace_modules_are_labelled_relatively() {
    let sandbox = create_empty_moon_sandbox();
    std::fs::create_dir_all(sandbox.root.join("nix")).unwrap();
    std::fs::write(sandbox.root.join("nix/cc.nix"), "{ pkgs }: pkgs.gcc").unwrap();
    std::fs::create_dir_all(sandbox.root.join("packages/proj")).unwrap();
    std::fs::write(sandbox.root.join("packages/proj/flake.nix"), "{}").unwrap();

    let plugin = sandbox.create_toolchain("nix").await;
    let output = plugin.hash_task_contents(ws_input("packages/proj")).await;
    let blob = serde_json::to_string(&output.contents).unwrap();

    assert!(
        blob.contains("nix/cc.nix"),
        "the shared module must be hashed, got: {blob}"
    );
    assert!(
        !blob.contains(&format!("{}/nix/cc.nix", sandbox.root.to_string_lossy())),
        "the label must be relative, not absolute, got: {blob}"
    );
}

#[tokio::test(flavor = "multi_thread", worker_threads = 1)]
#[serial]
async fn setup_environment_pre_builds_devshell_when_nix_present() {
    let sandbox = create_empty_moon_sandbox();
    let expected_root = format!("{}/", sandbox.root.to_string_lossy());
    let plugin = sandbox.create_toolchain("nix").await;

    let output = plugin
        .setup_environment(SetupEnvironmentInput::default())
        .await;

    assert_eq!(
        output.commands.len(),
        1,
        "exactly one pre-build command should be emitted, got: {:?}",
        output.commands
    );
    let command = &output.commands[0];
    assert!(
        command.allow_failure,
        "the pre-build must be non-blocking (allow_failure)"
    );
    assert_eq!(command.command.command, "nix");
    let args = &command.command.args;
    assert_eq!(args.len(), 8, "got: {args:?}");
    assert_eq!(args[0], "develop");
    assert_eq!(&args[1..4], ["--option", "eval-cache", "false"]);
    assert_eq!(args[4], "--no-update-lock-file");
    assert_eq!(args[5], expected_root);
    assert_eq!(args[6], "--command");
    assert_eq!(args[7], "true");
}
