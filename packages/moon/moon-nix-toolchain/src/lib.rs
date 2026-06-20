use extism_pdk::*;
use moon_pdk::*;
use moon_pdk_api::*;

/// Env var set on a wrapped task so a task is wrapped at most once, even outside
/// a dev shell (belt-and-suspenders alongside the `IN_NIX_SHELL` guard).
const SENTINEL: &str = "MOON_NIX_WRAPPED";

#[plugin_fn]
pub fn register_toolchain(
    Json(_): Json<RegisterToolchainInput>,
) -> FnResult<Json<RegisterToolchainOutput>> {
    Ok(Json(RegisterToolchainOutput {
        name: "Nix".into(),
        plugin_version: env!("CARGO_PKG_VERSION").into(),
        description: Some("Runs every task inside the workspace nix flake dev shell.".into()),
        ..Default::default()
    }))
}

/// Return the real workspace root to wrap with, or `None` when the task must run
/// unchanged: already inside a dev shell (CI's outer `nix develop`), already
/// wrapped, `nix` is unavailable, or the workspace root cannot be resolved.
fn resolve_wrap_root(context: &MoonContext) -> AnyResult<Option<String>> {
    if !get_host_env_var("IN_NIX_SHELL")?
        .unwrap_or_default()
        .is_empty()
    {
        return Ok(None);
    }

    if get_host_env_var(SENTINEL)?.unwrap_or_default() == "1" {
        return Ok(None);
    }

    if !command_exists(&get_host_environment()?, "nix") {
        return Ok(None);
    }

    Ok(context
        .workspace_root
        .real_path()
        .map(|path| path.to_string_lossy().into_owned()))
}

/// Build the `nix develop` flake reference for a wrapped task: the workspace root,
/// plus a `#<shell>` attribute selecting a named devShell. The shell is chosen from
/// the merged toolchain config — a per-task entry in the `shells` map (keyed by task
/// id) wins, else the project-wide `shell`. An unset, empty, or `default` shell uses
/// the flake's default devShell (the root, no attribute).
fn flake_ref(
    workspace_root: &str,
    task_id: Option<&str>,
    toolchain_config: &serde_json::Value,
) -> String {
    let shell = task_id
        .and_then(|id| {
            toolchain_config
                .get("shells")
                .and_then(|shells| shells.get(id))
                .and_then(|value| value.as_str())
        })
        .or_else(|| {
            toolchain_config
                .get("shell")
                .and_then(|value| value.as_str())
        })
        .map(str::trim)
        .filter(|shell| !shell.is_empty() && *shell != "default");

    match shell {
        Some(shell) => format!("{workspace_root}#{shell}"),
        None => workspace_root.to_string(),
    }
}

/// Quote a string for safe inclusion as a single POSIX shell argument.
fn shell_quote(value: &str) -> String {
    format!("'{}'", value.replace('\'', "'\\''"))
}

#[plugin_fn]
pub fn extend_task_command(
    Json(input): Json<ExtendTaskCommandInput>,
) -> FnResult<Json<ExtendTaskCommandOutput>> {
    let mut output = ExtendTaskCommandOutput::default();

    let Some(workspace_root) = resolve_wrap_root(&input.context)? else {
        return Ok(Json(output));
    };

    let task_id = input.task.target.get_task_id().ok();

    // Rebuild the entire argv: nix develop <flakeref> --command <cmd> <args...>.
    // `--command` must be the last `nix` flag; everything after it is the child
    // argv, passed through verbatim with no shell layer.
    let mut args = vec![
        "develop".to_string(),
        flake_ref(&workspace_root, task_id, &input.toolchain_config),
        "--command".to_string(),
        input.command.clone(),
    ];
    args.extend(input.args.clone());

    output.command = Some("nix".into());
    output.args = Some(Extend::Replace(args));
    output.env.insert(SENTINEL.into(), "1".into());

    Ok(Json(output))
}

#[plugin_fn]
pub fn extend_task_script(
    Json(input): Json<ExtendTaskScriptInput>,
) -> FnResult<Json<ExtendTaskScriptOutput>> {
    let mut output = ExtendTaskScriptOutput::default();

    let Some(workspace_root) = resolve_wrap_root(&input.context)? else {
        return Ok(Json(output));
    };

    let task_id = input.task.target.get_task_id().ok();

    // A script is one opaque string, so it needs a shell layer inside the dev
    // shell: nix develop <flakeref> --command bash -c "<original script>".
    output.script = Some(format!(
        "nix develop {} --command bash -c {}",
        shell_quote(&flake_ref(
            &workspace_root,
            task_id,
            &input.toolchain_config
        )),
        shell_quote(&input.script)
    ));
    output.env.insert(SENTINEL.into(), "1".into());

    Ok(Json(output))
}
