use extism_pdk::*;
use moon_nix_runtime::guard::{
    decide_identity_wrap, IdentityWrapFacts, WrapGuard, SENTINEL, SHELL_IDENTITY,
};
use moon_nix_runtime::serialize::escape_nix_string;
use moon_nix_runtime::target::{resolve_flake_target as resolve_runtime_flake_target, FlakeTarget};
use moon_nix_runtime::wrap::{
    effective_shell as resolve_effective_shell, flake_ref, plan_command, plan_script, WrapDecision,
};
use moon_pdk::*;
use moon_pdk_api::*;
use schematic::{Config, SchemaBuilder};
use std::collections::HashMap;
use std::path::PathBuf;

const SHELL_IDENTITY_PREFIX: &str = "moon_nix_toolchain:v1:";

struct ResolvedWrapTarget {
    target: FlakeTarget,
    shell: Option<String>,
    identity: String,
}

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

fn enforce_missing_nix(
    project_id: &str,
    target_id: &str,
    config: &NixToolchainConfig,
) -> AnyResult<()> {
    if fail_closed_opted_in(project_id, config)? {
        return Err(anyhow!(
            "nix is required for `{target_id}` but `nix` was not found on PATH; \
             this project opted into fail-closed nix \
             (failClosedByTag / failClosedByLanguage)"
        ));
    }

    Ok(())
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

/// Return the exact flake and devShell to wrap the task with, or `None` when the task
/// already carries that environment identity, `nix` is unavailable for a non-opted
/// project, or no real path resolves.
/// Returns `Err` when `nix` is unavailable but the project opted into fail-closed
/// nix (see `fail_closed_opted_in`). When the task's project has its own `flake.nix`,
/// that project flake wins over the workspace flake.
fn resolve_wrap_target(
    context: &MoonContext,
    project: &ProjectFragment,
    task: &TaskFragment,
    config: &NixToolchainConfig,
) -> AnyResult<Option<ResolvedWrapTarget>> {
    let current_identity = get_host_env_var(SHELL_IDENTITY)?.unwrap_or_default();
    let host_environment = get_host_environment()?;
    let nix_available = command_exists(host_environment, "nix");

    if current_identity.is_empty() && !nix_available {
        enforce_missing_nix(project.id.as_str(), task.target.as_str(), config)?;
        return Ok(None);
    }

    let Some(target) = resolve_flake_target(context, project.source.as_str())? else {
        return Ok(None);
    };

    let configured_shell = resolve_shell(
        task.target.get_task_id().ok(),
        &task.toolchains,
        project.id.as_str(),
        config,
    )?;
    let configured_identity = shell_identity(&target, configured_shell.as_deref());
    let facts = IdentityWrapFacts {
        current_identity: (!current_identity.is_empty()).then_some(current_identity.as_str()),
        required_identity: &configured_identity,
        nix_available,
    };

    match decide_identity_wrap(facts) {
        WrapGuard::Unchanged => Ok(None),
        WrapGuard::Ready => {
            let shell = effective_shell(&target, configured_shell)?;
            let identity = shell_identity(&target, shell.as_deref());

            if current_identity == identity {
                return Ok(None);
            }

            Ok(Some(ResolvedWrapTarget {
                target,
                shell,
                identity,
            }))
        }
        WrapGuard::MissingNix => {
            enforce_missing_nix(project.id.as_str(), task.target.as_str(), config)?;
            Ok(None)
        }
    }
}

fn shell_identity(target: &FlakeTarget, shell: Option<&str>) -> String {
    format!("{SHELL_IDENTITY_PREFIX}{}", flake_ref(&target.root, shell))
}

/// Resolve the flake that wraps a task purely from paths, with no runtime guards:
/// the project flake when `<project>/flake.nix` exists, else the workspace flake.
/// Shared by `resolve_wrap_target` (after its guards) and `hash_task_contents`,
/// whose cache key must not depend on transient env (`IN_NIX_SHELL`,
/// `MOON_NIX_WRAPPED`, or `MOON_NIX_SHELL_ID`) or on whether `nix` is installed
/// on the hashing host.
fn resolve_flake_target(
    context: &MoonContext,
    project_source: &str,
) -> AnyResult<Option<FlakeTarget>> {
    let workspace_root = canonical_workspace_root(context)?;
    let project_flake_root = if !project_source.is_empty() {
        let project_root = workspace_root.join(project_source);
        let flake = project_root.join("flake.nix");
        let flake_path = flake.to_string_lossy();

        if exec_captured("test", ["-f", flake_path.as_ref()])
            .is_ok_and(|result| result.exit_code == 0)
        {
            Some(project_root.to_string_lossy().into_owned())
        } else {
            None
        }
    } else {
        None
    };

    Ok(resolve_runtime_flake_target(
        Some(workspace_root.to_string_lossy().into_owned()),
        project_flake_root,
    ))
}

fn canonical_workspace_root(context: &MoonContext) -> AnyResult<PathBuf> {
    context
        .workspace_root
        .to_real_path()?
        .map(RealPath::into_inner)
        .ok_or_else(|| anyhow!("Moon workspace root has no real filesystem path"))
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

/// Whether the flake at `root` exposes a devShell named `shell` for the current
/// system. Evaluates `<root>#devShells` only (attribute names, never building the
/// shell) and never writes a lock file, so it does not mutate the project. `nix` is
/// guaranteed present past the wrap guards. A valid `false` result selects the
/// default shell; command and evaluation failures stop wrapping.
fn flake_exposes_shell(root: &str, shell: &str) -> AnyResult<bool> {
    let escaped = escape_nix_string(shell);
    let reference = format!("{root}#devShells");
    let apply =
        format!("sets: builtins.hasAttr \"{escaped}\" (sets.${{builtins.currentSystem}} or {{}})");
    let result = exec_captured(
        "nix",
        [
            "eval",
            "--impure",
            "--option",
            "eval-cache",
            "false",
            "--no-update-lock-file",
            "--json",
            reference.as_str(),
            "--apply",
            apply.as_str(),
        ],
    )?;
    if result.exit_code != 0 {
        return Err(anyhow!(
            "failed to inspect devShell {shell:?} in {root:?}: nix eval exited {}",
            result.exit_code
        ));
    }
    match result.stdout.trim() {
        "true" => Ok(true),
        "false" => Ok(false),
        output => Err(anyhow!(
            "failed to inspect devShell {shell:?} in {root:?}: unexpected nix eval output {output:?}"
        )),
    }
}

/// The devShell selector actually used to wrap a task. A project-flake selector the
/// flake does not expose is dropped (falling back to the flake's `default`), so a
/// configured `#<shell>` can never hard-fail `nix develop`. Workspace-flake selectors
/// and the no-selector case are returned unchanged — the existence probe runs only
/// for a project flake with a resolved name.
fn effective_shell(target: &FlakeTarget, shell: Option<String>) -> AnyResult<Option<String>> {
    let named_shell_available = match shell.as_deref() {
        Some(name) if target.is_project_flake => flake_exposes_shell(&target.root, name)?,
        _ => true,
    };

    Ok(resolve_effective_shell(
        target,
        shell,
        named_shell_available,
    ))
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

fn host_file_contents(path: &str) -> Option<String> {
    exec_captured("cat", [path])
        .ok()
        .filter(|result| result.exit_code == 0)
        .map(|result| result.stdout)
}

/// Every `*.nix` under `dir`, or nothing when `dir` does not exist.
fn find_nix_files(dir: &str) -> Vec<String> {
    exec_captured("find", [dir, "-type", "f", "-name", "*.nix", "-print"])
        .ok()
        .filter(|result| result.exit_code == 0)
        .map(|result| {
            result
                .stdout
                .lines()
                .filter(|path| !path.is_empty())
                .map(str::to_owned)
                .collect()
        })
        .unwrap_or_default()
}

/// Read each path, labelling it relative to `label_root`.
///
/// The label must be relative: an absolute path would differ between CI and a
/// developer's machine and the cache key would never match across them.
fn read_nix_sources(mut paths: Vec<String>, label_root: &str) -> Vec<serde_json::Value> {
    paths.sort();
    paths.dedup();

    paths
        .into_iter()
        .filter_map(|path| {
            let contents = host_file_contents(&path)?;
            let relative_path = path
                .strip_prefix(label_root)
                .unwrap_or(&path)
                .trim_start_matches('/');
            Some(serde_json::json!({
                "path": relative_path,
                "contents": contents,
            }))
        })
        .collect()
}

/// Read the flake entry point and its conventional `nix/**/*.nix` modules so
/// edits to toolchain definitions invalidate Moon's task cache.
fn flake_source_contents(root: &str) -> Vec<serde_json::Value> {
    let mut paths = vec![format!("{root}/flake.nix")];
    paths.extend(find_nix_files(&format!("{root}/nix")));
    read_nix_sources(paths, root)
}

/// The workspace's shared `nix/**/*.nix` modules, hashed on behalf of a project
/// that resolves its own flake.
///
/// A project flake typically composes the workspace's devShells through a
/// relative `path:` input, and relative path inputs carry no `narHash`
/// (Nix 2.26+, which resolves them against the parent source tree instead of
/// fetching them as an independent locked tree). So when a shared module such as
/// `nix/cc.nix` is edited, the project's own `flake.nix` is unchanged AND its
/// `flake.lock` is unchanged — while the devShell it resolves to genuinely
/// changes. Without this the task hash stays byte-identical and Moon serves a
/// cache hit built with the previous toolchain.
///
/// Labelled relative to the workspace root, so the key is identical wherever it
/// is computed.
fn workspace_module_contents(workspace_root: &str) -> Vec<serde_json::Value> {
    read_nix_sources(
        find_nix_files(&format!("{workspace_root}/nix")),
        workspace_root,
    )
}

fn command_output(decision: WrapDecision, identity: &str) -> ExtendTaskCommandOutput {
    let mut output = ExtendTaskCommandOutput::default();
    let WrapDecision::Command { command, args } = decision else {
        return output;
    };

    output.command = Some(command);
    output.args = Some(Extend::Replace(args));
    output.env.insert(SENTINEL.into(), "1".into());
    output.env.insert(SHELL_IDENTITY.into(), identity.into());
    output
}

fn script_output(decision: WrapDecision, identity: &str) -> ExtendTaskScriptOutput {
    let mut output = ExtendTaskScriptOutput::default();
    let WrapDecision::Script(script) = decision else {
        return output;
    };

    output.script = Some(script);
    output.env.insert(SENTINEL.into(), "1".into());
    output.env.insert(SHELL_IDENTITY.into(), identity.into());
    output
}

#[plugin_fn]
pub fn extend_task_command(
    Json(input): Json<ExtendTaskCommandInput>,
) -> FnResult<Json<ExtendTaskCommandOutput>> {
    let config: NixToolchainConfig = parse_toolchain_config_schema(input.toolchain_config.clone())?;

    let Some(resolved) = resolve_wrap_target(&input.context, &input.project, &input.task, &config)?
    else {
        return Ok(Json(ExtendTaskCommandOutput::default()));
    };
    let decision = plan_command(
        Some(&resolved.target),
        resolved.shell.as_deref(),
        input.command,
        input.args,
    );

    Ok(Json(command_output(decision, &resolved.identity)))
}

#[plugin_fn]
pub fn extend_task_script(
    Json(input): Json<ExtendTaskScriptInput>,
) -> FnResult<Json<ExtendTaskScriptOutput>> {
    let config: NixToolchainConfig = parse_toolchain_config_schema(input.toolchain_config.clone())?;

    let Some(resolved) = resolve_wrap_target(&input.context, &input.project, &input.task, &config)?
    else {
        return Ok(Json(ExtendTaskScriptOutput::default()));
    };
    let decision = plan_script(
        Some(&resolved.target),
        resolved.shell.as_deref(),
        &input.script,
    );

    Ok(Json(script_output(decision, &resolved.identity)))
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
        let mut sources = flake_source_contents(&target.root);

        // A project flake's own sources and lock do not cover the workspace's
        // shared modules it composes through a relative `path:` input — those
        // inputs carry no narHash, so nothing under the project root moves when
        // one is edited. See `workspace_module_contents`.
        if target.is_project_flake {
            let workspace_root = canonical_workspace_root(&input.context)?
                .to_string_lossy()
                .into_owned();
            sources.extend(workspace_module_contents(&workspace_root));
        }

        contents.push(serde_json::json!({
            "flakeRoot": target.root,
            "shell": shell,
            "flakeLock": flake_lock_contents(&target.root),
            "flakeSources": sources,
        }));
    }

    Ok(Json(HashTaskContentsOutput { contents }))
}

#[plugin_fn]
pub fn setup_environment(
    Json(input): Json<SetupEnvironmentInput>,
) -> FnResult<Json<SetupEnvironmentOutput>> {
    let mut output = SetupEnvironmentOutput::default();

    if !command_exists(get_host_environment()?, "nix") {
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
                [
                    "develop",
                    "--option",
                    "eval-cache",
                    "false",
                    "--no-update-lock-file",
                    reference.as_str(),
                    "--command",
                    "true",
                ],
            ))
            .allow_failure()
            .label(format!("Pre-building nix devShell {reference}")),
        );
    }

    Ok(Json(output))
}
