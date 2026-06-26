use extism_pdk::*;
use moon_pdk::*;
use moon_pdk_api::*;
use schematic::{Config, SchemaBuilder};
use std::collections::HashMap;

/// Env var set on a wrapped task so a task is wrapped at most once, even outside
/// a dev shell (belt-and-suspenders alongside the `IN_NIX_SHELL` guard).
const SENTINEL: &str = "MOON_NIX_WRAPPED";

/// Typed `nix` toolchain configuration, validated against the schema returned by
/// `define_toolchain_config`. The devShell selectors are resolved most-specific
/// first: `shell_by_task` > `shell_by_toolchain` > `shell_by_tag` >
/// `shell_by_language` > `shell` (a project-wide default).
#[derive(Clone, Config, Debug)]
#[config(rename_all = "camelCase")]
pub struct NixToolchainConfig {
    /// Project-wide default devShell name. Empty or `default` selects the flake's
    /// default devShell.
    pub shell: Option<String>,

    /// devShell name keyed by task id.
    pub shell_by_task: HashMap<String, String>,

    /// devShell name keyed by a task toolchain id.
    pub shell_by_toolchain: HashMap<String, String>,

    /// devShell name keyed by a project tag.
    pub shell_by_tag: HashMap<String, String>,

    /// devShell name keyed by the project language.
    pub shell_by_language: HashMap<String, String>,

    /// Project tags whose tasks MUST run inside nix. When `nix` is unavailable for a
    /// task in a project carrying one of these tags, the plugin errors instead of
    /// silently falling back to host tools. Empty (the default) = no enforcement.
    pub fail_closed_by_tag: Vec<String>,

    /// Project languages whose tasks MUST run inside nix — same fail-closed contract
    /// as `fail_closed_by_tag`, keyed on the project's moon `language`.
    pub fail_closed_by_language: Vec<String>,
}

#[host_fn]
extern "ExtismHost" {
    fn load_project_by_id(id: String) -> Json<serde_json::Value>;
}

/// Load a project's fragment over the host. The plugin sandbox cannot read the
/// workspace directly, so tags/language come from moon via `load_project_by_id`.
fn load_project(project_id: &str) -> AnyResult<serde_json::Value> {
    Ok(unsafe { load_project_by_id(project_id.to_owned())? }.0)
}

/// The project's moon tags (`config.tags`), empty when unset.
fn project_tags(project: &serde_json::Value) -> impl Iterator<Item = &str> {
    project
        .get("config")
        .and_then(|config| config.get("tags"))
        .and_then(|tags| tags.as_array())
        .into_iter()
        .flatten()
        .filter_map(|tag| tag.as_str())
}

/// The project's moon `language`, when present.
fn project_language(project: &serde_json::Value) -> Option<&str> {
    project
        .get("language")
        .and_then(|language| language.as_str())
}

/// Whether the task's project opted into fail-closed nix: one of its tags is in
/// `failClosedByTag`, or its language is in `failClosedByLanguage`. Returns
/// `false` without a host round-trip when both allowlists are empty.
fn fail_closed_opted_in(project_id: &str, config: &NixToolchainConfig) -> AnyResult<bool> {
    if config.fail_closed_by_tag.is_empty() && config.fail_closed_by_language.is_empty() {
        return Ok(false);
    }

    let project = load_project(project_id)?;

    let tag_opt_in = project_tags(&project)
        .any(|tag| config.fail_closed_by_tag.iter().any(|allow| allow == tag));
    let language_opt_in = project_language(&project).is_some_and(|language| {
        config
            .fail_closed_by_language
            .iter()
            .any(|allow| allow == language)
    });

    Ok(tag_opt_in || language_opt_in)
}

#[plugin_fn]
pub fn register_toolchain(
    Json(_): Json<RegisterToolchainInput>,
) -> FnResult<Json<RegisterToolchainOutput>> {
    Ok(Json(RegisterToolchainOutput {
        name: "Nix".into(),
        plugin_version: env!("CARGO_PKG_VERSION").into(),
        description: Some(
            "Runs every task inside the project's or workspace's nix flake dev shell.".into(),
        ),
        ..Default::default()
    }))
}

/// Register the JSON schema for the typed toolchain config, so moon validates the
/// `shell`/`shellBy*` keys (unknown key, wrong type) at config-load time instead of
/// silently ignoring them.
#[plugin_fn]
pub fn define_toolchain_config() -> FnResult<Json<DefineToolchainConfigOutput>> {
    Ok(Json(DefineToolchainConfigOutput {
        schema: SchemaBuilder::build_root::<NixToolchainConfig>(),
    }))
}

/// The nix flake a task is wrapped with: the real path passed to `nix develop`. A
/// project-local `flake.nix` wins over the workspace flake; either way the resolved
/// devShell selector (`resolve_shell`) is applied, so a project flake routes a task to
/// `{root}#<shell>` when a selector matches. For a project flake a selected name the
/// flake does not expose is dropped (`effective_shell`), falling back to its `default`.
struct WrapTarget {
    root: String,
    /// True when `root` is a project-local flake (`<project>/flake.nix`), false for the
    /// workspace flake. Only project-flake selectors are existence-checked before
    /// routing; the workspace flake is curated to expose its configured devShells.
    is_project_flake: bool,
}

/// Return the flake to wrap the task with, or `None` when the task must run
/// unchanged: already inside a dev shell (CI's outer `nix develop`), already
/// wrapped, `nix` is unavailable for a non-opted project, or no real path resolves.
/// Returns `Err` when `nix` is unavailable but the project opted into fail-closed
/// nix (see `fail_closed_opted_in`). When the task's project has its own `flake.nix`,
/// that project flake wins over the workspace flake.
fn resolve_wrap_target(
    context: &MoonContext,
    project: &ProjectFragment,
    target_id: &str,
    config: &NixToolchainConfig,
) -> AnyResult<Option<WrapTarget>> {
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
        // Fail closed for opted-in projects: a task that must run inside nix must
        // not silently fall back to host tools when `nix` is absent.
        if fail_closed_opted_in(project.id.as_str(), config)? {
            return Err(anyhow!(
                "nix is required for `{target_id}` but `nix` was not found on PATH; \
                 this project opted into fail-closed nix \
                 (failClosedByTag / failClosedByLanguage)"
            ));
        }
        return Ok(None);
    }

    resolve_flake_target(context, project.source.as_str())
}

/// Resolve the flake that wraps a task purely from paths, with no runtime guards:
/// the project flake when `<project>/flake.nix` exists, else the workspace flake.
/// Shared by `resolve_wrap_target` (after its guards) and `hash_task_contents`,
/// whose cache key must not depend on transient env (`IN_NIX_SHELL`/`MOON_NIX_WRAPPED`)
/// or on whether `nix` is installed on the hashing host.
fn resolve_flake_target(
    context: &MoonContext,
    project_source: &str,
) -> AnyResult<Option<WrapTarget>> {
    if !project_source.is_empty() {
        if let Some(project_root) = context.workspace_root.join(project_source).real_path() {
            // Detect the project flake over the host: the plugin's sandbox has no
            // direct read access to the workspace, so `VirtualPath::is_file` cannot
            // see it. `test -f` runs on the host against the real project path.
            let flake = project_root.join("flake.nix");
            let flake_path = flake.to_string_lossy();

            if exec_captured("test", ["-f", flake_path.as_ref()])
                .is_ok_and(|result| result.exit_code == 0)
            {
                return Ok(Some(WrapTarget {
                    root: project_root.to_string_lossy().into_owned(),
                    is_project_flake: true,
                }));
            }
        }
    }

    Ok(context.workspace_root.real_path().map(|path| WrapTarget {
        root: path.to_string_lossy().into_owned(),
        is_project_flake: false,
    }))
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
    config: &NixToolchainConfig,
) -> AnyResult<Option<String>> {
    if let Some(value) = task_id.and_then(|id| config.shell_by_task.get(id)) {
        return Ok(normalize_shell(value));
    }

    for toolchain in task_toolchains {
        if let Some(value) = config.shell_by_toolchain.get(toolchain.as_str()) {
            return Ok(normalize_shell(value));
        }
    }

    if !config.shell_by_tag.is_empty() || !config.shell_by_language.is_empty() {
        let project = load_project(project_id)?;

        if !config.shell_by_tag.is_empty() {
            for tag in project_tags(&project) {
                if let Some(value) = config.shell_by_tag.get(tag) {
                    return Ok(normalize_shell(value));
                }
            }
        }

        if let Some(value) =
            project_language(&project).and_then(|language| config.shell_by_language.get(language))
        {
            return Ok(normalize_shell(value));
        }
    }

    Ok(config.shell.as_deref().and_then(normalize_shell))
}

/// Build the `nix develop` flake reference: the flake root, plus a `#<shell>`
/// attribute when a named devShell is selected, else the root's default devShell.
fn flake_ref(root: &str, shell: Option<&str>) -> String {
    match shell {
        Some(shell) => format!("{root}#{shell}"),
        None => root.to_string(),
    }
}

/// Whether the flake at `root` exposes a devShell named `shell` for the current
/// system. Evaluates `<root>#devShells` only (attribute names, never building the
/// shell) and never writes a lock file, so it does not mutate the project. `nix` is
/// guaranteed present past the wrap guards; a probe failure (eval error, no
/// `devShells` output, missing system) reports `false` so the wrap degrades to the
/// flake's `default` rather than emitting a `#<shell>` that `nix develop` rejects.
fn flake_exposes_shell(root: &str, shell: &str) -> bool {
    // Escape the name for a `"..."` Nix string literal: backslash, double-quote, and
    // the `${` interpolation opener (`\${` is a literal `${`).
    let escaped = shell
        .replace('\\', "\\\\")
        .replace('"', "\\\"")
        .replace("${", "\\${");
    let reference = format!("{root}#devShells");
    let apply =
        format!("sets: builtins.hasAttr \"{escaped}\" (sets.${{builtins.currentSystem}} or {{}})");
    exec_captured(
        "nix",
        [
            "eval",
            "--impure",
            "--no-write-lock-file",
            "--json",
            reference.as_str(),
            "--apply",
            apply.as_str(),
        ],
    )
    .is_ok_and(|result| result.exit_code == 0 && result.stdout.trim() == "true")
}

/// The devShell selector actually used to wrap a task. A project-flake selector the
/// flake does not expose is dropped (falling back to the flake's `default`), so a
/// configured `#<shell>` can never hard-fail `nix develop`. Workspace-flake selectors
/// and the no-selector case are returned unchanged — the existence probe runs only
/// for a project flake with a resolved name.
fn effective_shell(target: &WrapTarget, shell: Option<String>) -> Option<String> {
    match shell {
        Some(name) if target.is_project_flake && !flake_exposes_shell(&target.root, &name) => None,
        other => other,
    }
}

/// Read a flake's `flake.lock` over the host so its pinned inputs fold into the
/// task hash. Returns an empty string when the lock is absent (a flake with no
/// lock, or a non-flake workspace root) — an absent lock is a stable value, so it
/// never thrashes the cache.
fn flake_lock_contents(root: &str) -> String {
    let lock_path = format!("{root}/flake.lock");
    exec_captured("cat", [lock_path.as_str()])
        .ok()
        .filter(|result| result.exit_code == 0)
        .map(|result| result.stdout)
        .unwrap_or_default()
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

    let config: NixToolchainConfig = parse_toolchain_config_schema(input.toolchain_config.clone())?;

    let Some(target) = resolve_wrap_target(
        &input.context,
        &input.project,
        input.task.target.id.as_str(),
        &config,
    )?
    else {
        return Ok(Json(output));
    };

    // Resolve the devShell selector for every flake, project-local or workspace, then
    // drop a project-flake name the flake does not expose so the wrap falls back to its
    // default devShell instead of a `#<shell>` that `nix develop` would reject.
    let shell = effective_shell(
        &target,
        resolve_shell(
            input.task.target.get_task_id().ok(),
            &input.task.toolchains,
            input.project.id.as_str(),
            &config,
        )?,
    );

    // Rebuild the entire argv: nix develop <flakeref> --command <cmd> <args...>.
    // `--command` must be the last `nix` flag; everything after it is the child
    // argv, passed through verbatim with no shell layer.
    let mut args = vec![
        "develop".to_string(),
        flake_ref(&target.root, shell.as_deref()),
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

    let config: NixToolchainConfig = parse_toolchain_config_schema(input.toolchain_config.clone())?;

    let Some(target) = resolve_wrap_target(
        &input.context,
        &input.project,
        input.task.target.id.as_str(),
        &config,
    )?
    else {
        return Ok(Json(output));
    };

    let shell = effective_shell(
        &target,
        resolve_shell(
            input.task.target.get_task_id().ok(),
            &input.task.toolchains,
            input.project.id.as_str(),
            &config,
        )?,
    );

    // A script is one opaque string, so it needs a shell layer inside the dev
    // shell: nix develop <flakeref> --command bash -c "<original script>".
    output.script = Some(format!(
        "nix develop {} --command bash -c {}",
        shell_quote(&flake_ref(&target.root, shell.as_deref())),
        shell_quote(&input.script)
    ));
    output.env.insert(SENTINEL.into(), "1".into());

    Ok(Json(output))
}

#[plugin_fn]
pub fn hash_task_contents(
    Json(input): Json<HashTaskContentsInput>,
) -> FnResult<Json<HashTaskContentsOutput>> {
    let mut contents = Vec::new();

    let config: NixToolchainConfig = parse_toolchain_config_schema(input.toolchain_config.clone())?;

    if let Some(target) = resolve_flake_target(&input.context, input.project.source.as_str())? {
        // Mirror the wrap hooks' selector resolution, but track the *configured* shell
        // rather than `effective_shell`'s fallback: the cache key must stay independent
        // of `nix` (no devShell-existence probe on the hashing host), and the configured
        // selector is a stable proxy — it changes exactly when the config changes.
        let shell = resolve_shell(
            input.task.target.get_task_id().ok(),
            &input.task.toolchains,
            input.project.id.as_str(),
            &config,
        )?;

        // Fold the resolved flake root, the selected shell, and the lock's pinned
        // inputs into the cache key: editing flake.lock or switching the shell
        // changes `contents`; an unrelated edit leaves it byte-identical.
        contents.push(serde_json::json!({
            "flakeRoot": target.root,
            "shell": shell,
            "flakeLock": flake_lock_contents(&target.root),
        }));
    }

    Ok(Json(HashTaskContentsOutput { contents }))
}

#[plugin_fn]
pub fn setup_environment(
    Json(input): Json<SetupEnvironmentInput>,
) -> FnResult<Json<SetupEnvironmentOutput>> {
    let mut output = SetupEnvironmentOutput::default();

    if !command_exists(&get_host_environment()?, "nix") {
        return Ok(Json(output));
    }

    let project_source = input
        .project
        .as_ref()
        .map_or("", |project| project.source.as_str());
    if let Some(target) = resolve_flake_target(&input.context, project_source)? {
        // Realise (and cache) the devShell closure before the first task runs, so the
        // first wrapped task is not a cold `nix develop`. `allow_failure` keeps setup
        // non-blocking; the closure lands in the nix store without a GC root, so a
        // `nix store gc` before the task can still evict it.
        let reference = flake_ref(&target.root, None);
        output.commands.push(
            ExecCommand::new(ExecCommandInput::new(
                "nix",
                ["develop", reference.as_str(), "--command", "true"],
            ))
            .allow_failure()
            .label(format!("Pre-building nix devShell {reference}")),
        );
    }

    Ok(Json(output))
}
