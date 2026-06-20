use extism_pdk::*;
use moon_pdk::*;
use moon_pdk_api::*;

/// Env var set on a wrapped task so a task is wrapped at most once, even outside
/// a dev shell (belt-and-suspenders alongside the `IN_NIX_SHELL` guard).
const SENTINEL: &str = "MOON_NIX_WRAPPED";

#[host_fn]
extern "ExtismHost" {
    fn load_project_by_id(id: String) -> Json<serde_json::Value>;
}

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

/// Trim a configured devShell name, treating empty or `default` as no selection
/// (the flake's default devShell). A matched-but-`default` entry still resolves the
/// search — a more specific selector wins even when it points at the default shell.
fn normalize_shell(value: &str) -> Option<String> {
    let trimmed = value.trim();
    (!trimmed.is_empty() && trimmed != "default").then(|| trimmed.to_owned())
}

/// Select the flake devShell name for a task from the merged toolchain config, in
/// precedence order (most specific first): `shellByTask` keyed by task id,
/// `shellByToolchain` keyed by a task toolchain id, `shellByTag` keyed by a project
/// tag, `shellByLanguage` keyed by the project language, then the project-wide `shell`.
/// The first selector with a matching key resolves the search; its value is returned,
/// or `None` when that value is empty or `default`. Returns `None` when nothing matches
/// (the flake's default devShell). The project is loaded over the host only when a tag-
/// or language-based selector is configured and no more specific selector matched.
fn resolve_shell(
    task_id: Option<&str>,
    task_toolchains: &[Id],
    project_id: &str,
    config: &serde_json::Value,
) -> AnyResult<Option<String>> {
    if let Some(value) = task_id
        .and_then(|id| config.get("shellByTask").and_then(|map| map.get(id)))
        .and_then(|value| value.as_str())
    {
        return Ok(normalize_shell(value));
    }

    if let Some(map) = config.get("shellByToolchain") {
        for toolchain in task_toolchains {
            if let Some(value) = map.get(toolchain.as_str()).and_then(|value| value.as_str()) {
                return Ok(normalize_shell(value));
            }
        }
    }

    if config.get("shellByTag").is_some() || config.get("shellByLanguage").is_some() {
        let project = unsafe { load_project_by_id(project_id.to_owned())? }.0;

        if let Some(map) = config.get("shellByTag") {
            let tags = project
                .get("config")
                .and_then(|config| config.get("tags"))
                .and_then(|tags| tags.as_array());

            if let Some(tags) = tags {
                for tag in tags.iter().filter_map(|tag| tag.as_str()) {
                    if let Some(value) = map.get(tag).and_then(|value| value.as_str()) {
                        return Ok(normalize_shell(value));
                    }
                }
            }
        }

        if let Some(value) = config
            .get("shellByLanguage")
            .zip(
                project
                    .get("language")
                    .and_then(|language| language.as_str()),
            )
            .and_then(|(map, language)| map.get(language))
            .and_then(|value| value.as_str())
        {
            return Ok(normalize_shell(value));
        }
    }

    Ok(config
        .get("shell")
        .and_then(|value| value.as_str())
        .and_then(normalize_shell))
}

/// Build the `nix develop` flake reference: the workspace root, plus a `#<shell>`
/// attribute when a named devShell is selected, else the root's default devShell.
fn flake_ref(workspace_root: &str, shell: Option<&str>) -> String {
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

    let shell = resolve_shell(
        input.task.target.get_task_id().ok(),
        &input.task.toolchains,
        input.project.id.as_str(),
        &input.toolchain_config,
    )?;

    // Rebuild the entire argv: nix develop <flakeref> --command <cmd> <args...>.
    // `--command` must be the last `nix` flag; everything after it is the child
    // argv, passed through verbatim with no shell layer.
    let mut args = vec![
        "develop".to_string(),
        flake_ref(&workspace_root, shell.as_deref()),
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

    let shell = resolve_shell(
        input.task.target.get_task_id().ok(),
        &input.task.toolchains,
        input.project.id.as_str(),
        &input.toolchain_config,
    )?;

    // A script is one opaque string, so it needs a shell layer inside the dev
    // shell: nix develop <flakeref> --command bash -c "<original script>".
    output.script = Some(format!(
        "nix develop {} --command bash -c {}",
        shell_quote(&flake_ref(&workspace_root, shell.as_deref())),
        shell_quote(&input.script)
    ));
    output.env.insert(SENTINEL.into(), "1".into());

    Ok(Json(output))
}
