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

    // Rebuild the entire argv: nix develop <root> --command <cmd> <args...>.
    // `--command` must be the last `nix` flag; everything after it is the child
    // argv, passed through verbatim with no shell layer.
    let mut args = vec![
        "develop".to_string(),
        workspace_root,
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

    // A script is one opaque string, so it needs a shell layer inside the dev
    // shell: nix develop <root> --command bash -c "<original script>".
    output.script = Some(format!(
        "nix develop {} --command bash -c {}",
        shell_quote(&workspace_root),
        shell_quote(&input.script)
    ));
    output.env.insert(SENTINEL.into(), "1".into());

    Ok(Json(output))
}
